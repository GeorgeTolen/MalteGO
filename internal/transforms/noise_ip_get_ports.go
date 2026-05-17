package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetPorts struct{}

func (t *NoiseIPLookupGetPorts) Name() string { return "GreyNoiseNoiseIPLookupGetPorts" }

func (t *NoiseIPLookupGetPorts) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	if len(r.RawData.Scan) == 0 {
		resp.Inform("The IP address " + req.Value + " has no associated Ports.")
		return resp, nil
	}

	for _, scan := range r.RawData.Scan {
		portStr := fmt.Sprintf("%d", scan.Port)
		resp.AddEntity(maltego.EntityPort, portStr).
			AddProperty("link#maltego.link.label", "Label", maltego.MatchingRuleLoose, "Scans For")
	}

	return resp, nil
}
