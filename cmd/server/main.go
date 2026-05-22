package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

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

	srv := server.New(cfg, registry, nil)

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
	if len(args) < 1 {
		return fmt.Errorf("usage: %s local <TransformName>", os.Args[0])
	}
	name := args[0]

	// Read XML from stdin with a 3-second timeout.
	// Maltego sends XML via stdin but on Windows may not close it,
	// so we use a goroutine + timer to avoid blocking forever.
	stdinCh := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(os.Stdin)
		stdinCh <- data
	}()

	var xmlData []byte
	select {
	case xmlData = <-stdinCh:
	case <-time.After(3 * time.Second):
		// stdin not closed — fall through with empty data
	}

	var req *maltego.Request
	if len(xmlData) > 10 {
		parsed, err := maltego.ParseRequest(xmlData)
		if err != nil {
			return fmt.Errorf("parse request: %w", err)
		}
		req = parsed
	} else {
		// Fallback: value from second arg (for manual/CLI testing).
		value := ""
		if len(args) > 1 {
			value = args[1]
		}
		req = &maltego.Request{
			Value:      value,
			EntityType: localInputType(name),
			Properties: map[string]string{},
			Settings:   map[string]string{},
			SoftLimit:  12,
			HardLimit:  12,
		}
	}

	// API key: prefer from request TransformFields, fallback to config.
	if req.APIKey(cfg.GreyNoiseAPIKey) == "" {
		return fmt.Errorf("GreyNoise API key not configured")
	}

	client := greynoise.NewClient(req.APIKey(cfg.GreyNoiseAPIKey), cfg.RequestTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	resp, err := registry.Run(ctx, name, client, req)
	if err != nil {
		return err
	}
	out, err := resp.ToXML()
	if err != nil {
		return err
	}
	fmt.Print(string(out))
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
