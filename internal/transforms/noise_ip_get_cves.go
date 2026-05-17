package transforms

import (
	"context"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetCVEs struct{}

func (t *NoiseIPLookupGetCVEs) Name() string { return "GreyNoiseNoiseIPLookupGetCVEs" }

func (t *NoiseIPLookupGetCVEs) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	if len(r.CVEs) == 0 {
		resp.Inform("The IP address " + req.Value + " has no associated CVEs.")
		return resp, nil
	}

	for _, cve := range r.CVEs {
		resp.AddEntity(maltego.EntityCVE, cve).
			AddProperty("link#maltego.link.label", "Label", maltego.MatchingRuleLoose, "Probes For")
	}

	return resp, nil
}
