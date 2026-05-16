package transforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

// Transform is the interface every transform must implement.
type Transform interface {
	Name() string
	Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error)
}

// Registry holds all registered transforms by lowercase name.
type Registry struct {
	transforms map[string]Transform
}

func NewRegistry() *Registry {
	return &Registry{transforms: make(map[string]Transform)}
}

func (r *Registry) Register(t Transform) {
	r.transforms[strings.ToLower(t.Name())] = t
}

func (r *Registry) Get(name string) (Transform, bool) {
	t, ok := r.transforms[strings.ToLower(name)]
	return t, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.transforms))
	for _, t := range r.transforms {
		names = append(names, t.Name())
	}
	return names
}

func (r *Registry) Run(ctx context.Context, name string, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("transform %q not found", name)
	}
	return t.Run(ctx, client, req)
}
