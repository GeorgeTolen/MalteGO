package main

import (
	"log"

	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/server"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

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

	srv := server.New(cfg, registry)

	log.Printf("MalteGO starting on port %s", cfg.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
