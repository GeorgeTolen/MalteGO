package transforms

import (
	"context"
	"errors"
	"testing"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

// stubTransform is a minimal Transform used only in registry tests.
type stubTransform struct {
	name string
	resp *maltego.Response
	err  error
}

func (s *stubTransform) Name() string { return s.name }
func (s *stubTransform) Run(_ context.Context, _ greynoise.Client, _ *maltego.Request) (*maltego.Response, error) {
	return s.resp, s.err
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubTransform{name: "MyTransform"})

	_, ok := reg.Get("MyTransform")
	if !ok {
		t.Error("Get(MyTransform) = false, want true")
	}
}

func TestRegistry_Get_CaseInsensitive(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubTransform{name: "MyTransform"})

	for _, name := range []string{"mytransform", "MYTRANSFORM", "MyTransform", "myTRANSFORM"} {
		if _, ok := reg.Get(name); !ok {
			t.Errorf("Get(%q) = false, want case-insensitive match", name)
		}
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	reg := NewRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) = true, want false")
	}
}

func TestRegistry_Names_ContainsAllRegistered(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubTransform{name: "Alpha"})
	reg.Register(&stubTransform{name: "Beta"})
	reg.Register(&stubTransform{name: "Gamma"})

	names := reg.Names()
	if len(names) != 3 {
		t.Fatalf("Names() len = %d, want 3", len(names))
	}

	nameSet := map[string]bool{}
	for _, n := range names {
		nameSet[n] = true
	}
	for _, want := range []string{"Alpha", "Beta", "Gamma"} {
		if !nameSet[want] {
			t.Errorf("Names() missing %q", want)
		}
	}
}

func TestRegistry_Names_EmptyRegistry(t *testing.T) {
	reg := NewRegistry()
	if names := reg.Names(); len(names) != 0 {
		t.Errorf("Names() = %v, want empty", names)
	}
}

func TestRegistry_Run_NotFound_ReturnsError(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Run(context.Background(), "ghost", nil, makeReq("1.2.3.4"))
	if err == nil {
		t.Error("Run(unknown) expected error, got nil")
	}
}

func TestRegistry_Run_PropagatesTransformError(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubTransform{name: "Broken", err: errors.New("boom")})

	_, err := reg.Run(context.Background(), "Broken", nil, makeReq("1.2.3.4"))
	if err == nil {
		t.Error("expected error from broken transform")
	}
}

func TestRegistry_Run_ReturnsTransformResponse(t *testing.T) {
	r := maltego.NewResponse()
	r.AddEntity(maltego.EntityIPv4Address, "9.9.9.9")

	reg := NewRegistry()
	reg.Register(&stubTransform{name: "Good", resp: r})

	resp, err := reg.Run(context.Background(), "Good", nil, makeReq("1.2.3.4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := resp.ToXML()
	px := mustParseXML(t, data)
	if len(px.Entities) != 1 || px.Entities[0].Value != "9.9.9.9" {
		t.Errorf("unexpected entities: %v", px.Entities)
	}
}

func TestRegistry_Register_OverwritesSameName(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&stubTransform{name: "Dup"})
	reg.Register(&stubTransform{name: "Dup"}) // second registration overwrites first

	if len(reg.Names()) != 1 {
		t.Errorf("Names() len = %d, want 1 after duplicate registration", len(reg.Names()))
	}
}
