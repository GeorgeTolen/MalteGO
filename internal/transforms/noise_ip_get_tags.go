package transforms

import (
	"context"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetTags struct{}

func (t *NoiseIPLookupGetTags) Name() string { return "GreyNoiseNoiseIPLookupGetTags" }

func (t *NoiseIPLookupGetTags) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	addInputEntity(resp, req, maltego.EntityIPv4Address)

	r, err := client.ContextIP(ctx, req.Value)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if !r.Seen {
		resp.Inform("The IP address " + req.Value + " hasn't been seen by GreyNoise.")
		return resp, nil
	}

	if len(r.Tags) == 0 {
		resp.Inform("The IP address " + req.Value + " has no associated Tags.")
		return resp, nil
	}

	for _, tag := range r.Tags {
		resp.AddEntity(maltego.EntityPhrase, tag)
	}

	return resp, nil
}
