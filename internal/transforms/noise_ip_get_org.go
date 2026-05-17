package transforms

import (
	"context"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetOrg struct{}

func (t *NoiseIPLookupGetOrg) Name() string { return "GreyNoiseNoiseIPLookupGetOrg" }

func (t *NoiseIPLookupGetOrg) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	if r.Metadata.Organization == "" {
		resp.Inform("The IP address " + req.Value + " has no associated Organization.")
		return resp, nil
	}

	resp.AddEntity(maltego.EntityCompany, r.Metadata.Organization)
	return resp, nil
}
