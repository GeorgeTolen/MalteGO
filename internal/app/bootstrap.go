// Package app provides shared startup logic for cmd binaries.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

// NewRegistry creates and registers all 13 GreyNoise transforms.
func NewRegistry() *transforms.Registry {
	r := transforms.NewRegistry()
	r.Register(&transforms.CommunityIPLookup{})
	r.Register(&transforms.NoiseIPLookupAllDetails{})
	r.Register(&transforms.NoiseIPLookupGetActor{})
	r.Register(&transforms.NoiseIPLookupGetCVEs{})
	r.Register(&transforms.NoiseIPLookupGetOrg{})
	r.Register(&transforms.NoiseIPLookupGetPorts{})
	r.Register(&transforms.NoiseIPLookupGetTags{})
	r.Register(&transforms.NoiseIPSims{})
	r.Register(&transforms.QueryByASN{})
	r.Register(&transforms.QueryByActor{})
	r.Register(&transforms.QueryByCVE{})
	r.Register(&transforms.QueryByTag{})
	r.Register(&transforms.RIOTIPLookup{})
	return r
}

// RunLocal executes a transform from the Maltego Desktop CLI.
// Maltego passes entity XML via stdin (with a 3s timeout workaround for Windows),
// falling back to os.Args value if stdin is empty.
func RunLocal(cfg *config.Config, registry *transforms.Registry, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s local <TransformName> [value]", os.Args[0])
	}
	name := args[0]

	// Read XML from stdin with a 3-second timeout.
	// Maltego on Windows may not close stdin, so we avoid blocking forever.
	stdinCh := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(os.Stdin)
		stdinCh <- data
	}()

	var xmlData []byte
	select {
	case xmlData = <-stdinCh:
	case <-time.After(3 * time.Second):
	}

	var req *maltego.Request
	if len(xmlData) > 10 {
		parsed, err := maltego.ParseRequest(xmlData)
		if err != nil {
			return fmt.Errorf("parse request: %w", err)
		}
		req = parsed
	} else {
		value := ""
		if len(args) > 1 {
			value = args[1]
		}
		req = &maltego.Request{
			Value:      value,
			EntityType: LocalInputType(name),
			Properties: map[string]string{},
			Settings:   map[string]string{},
			SoftLimit:  12,
			HardLimit:  12,
		}
	}

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

// LocalInputType returns the Maltego entity type for a given transform name.
func LocalInputType(name string) string {
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
