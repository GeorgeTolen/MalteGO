package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetActor struct{}

func (t *NoiseIPLookupGetActor) Name() string { return "GreyNoiseNoiseIPLookupGetActor" }

func (t *NoiseIPLookupGetActor) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	r, err := client.ContextIP(ctx, req.Value)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GreyNoise context lookup failed: %v", err))
		return resp, nil
	}

	if !r.Seen {
		resp.Inform(fmt.Sprintf("%s has not been seen by GreyNoise", req.Value))
		return resp, nil
	}

	if r.Actor == "" {
		resp.Inform(fmt.Sprintf("No actor attributed to %s", req.Value))
		return resp, nil
	}

	resp.AddEntity(maltego.EntityPerson, r.Actor).
		AddDisplayInfo("GreyNoise Actor", fmt.Sprintf("<b>Actor:</b> %s<br/><b>IP:</b> %s", r.Actor, r.IP)).
		AddProperty("ip", "Source IP", maltego.MatchingRuleLoose, r.IP).
		AddProperty("classification", "Classification", maltego.MatchingRuleLoose, r.Classification)

	resp.Inform(fmt.Sprintf("Actor for %s: %s", req.Value, r.Actor))
	return resp, nil
}
