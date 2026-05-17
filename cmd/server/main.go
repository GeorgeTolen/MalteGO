package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
	"github.com/greynoise-maltego/maltego-go/internal/server"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	registry := newRegistry()

	if len(os.Args) > 1 && os.Args[1] == "local" {
		if err := runLocal(cfg, registry, os.Args[2:]); err != nil {
			xmlErr, _ := maltego.ErrorResponse(err.Error())
			fmt.Print(string(xmlErr))
			os.Exit(1)
		}
		return
	}

	srv := server.New(cfg, registry)

	log.Printf("MalteGO starting on port %s", cfg.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func newRegistry() *transforms.Registry {
	registry := transforms.NewRegistry()
	registry.Register(&transforms.CommunityIPLookup{})
	registry.Register(&transforms.NoiseIPLookupAllDetails{})
	registry.Register(&transforms.NoiseIPLookupGetActor{})
	registry.Register(&transforms.NoiseIPLookupGetCVEs{})
	registry.Register(&transforms.NoiseIPLookupGetOrg{})
	registry.Register(&transforms.NoiseIPLookupGetPorts{})
	registry.Register(&transforms.NoiseIPLookupGetTags{})
	registry.Register(&transforms.NoiseIPSims{})
	registry.Register(&transforms.QueryByASN{})
	registry.Register(&transforms.QueryByActor{})
	registry.Register(&transforms.QueryByCVE{})
	registry.Register(&transforms.QueryByTag{})
	registry.Register(&transforms.RIOTIPLookup{})
	return registry
}

func runLocal(cfg *config.Config, registry *transforms.Registry, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s local <TransformName> <Value>", os.Args[0])
	}

	name := args[0]
	value := args[1]
	if cfg.GreyNoiseAPIKey == "" {
		return fmt.Errorf("GreyNoise API key not configured. Set GREYNOISE_API_KEY before running local transforms.")
	}

	req := &maltego.Request{
		Value:      value,
		EntityType: localInputType(name),
		Properties: map[string]string{},
		Settings:   map[string]string{"GNApiKey": cfg.GreyNoiseAPIKey},
		SoftLimit:  12,
		HardLimit:  12,
	}
	client := greynoise.NewClient(cfg.GreyNoiseAPIKey, cfg.RequestTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	resp, err := registry.Run(ctx, name, client, req)
	if err != nil {
		return err
	}
	xmlData, err := resp.ToXML()
	if err != nil {
		return err
	}
	fmt.Print(string(xmlData))
	return nil
}

func localInputType(name string) string {
	switch name {
	case "GreyNoiseQueryByASN":
		return maltego.EntityAS
	case "GreyNoiseQueryByActor":
		return maltego.EntityPerson
	case "GreyNoiseQueryByCVE":
		return maltego.EntityCVE
	case "GreyNoiseQueryByTag":
		return maltego.EntityPhrase
	default:
		return maltego.EntityIPv4Address
	}
}
