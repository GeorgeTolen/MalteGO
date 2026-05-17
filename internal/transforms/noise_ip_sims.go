package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPSims struct{}

func (t *NoiseIPSims) Name() string { return "GreyNoiseNoiseIPSims" }

func (t *NoiseIPSims) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	inputIP := addInputEntity(resp, req, maltego.EntityIPv4Address)

	limit := intSetting(req, "limit", 50)
	minimumScore := intSetting(req, "minimum_score", 90)

	r, err := client.SimilarIPs(ctx, req.Value, minimumScore, limit)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if len(r.Similar) == 0 {
		resp.Inform("The IP address " + req.Value + " has no similar IPs within GreyNoise.")
		return resp, nil
	}

	for _, sim := range r.Similar {
		resp.AddEntity(maltego.EntityIPv4Address, sim.IP)
	}

	ipAddress := r.IP
	if ipAddress == "" {
		ipAddress = req.Value
	}
	displayHTML := fmt.Sprintf(
		`<h3><a href="https://viz.greynoise.io/ip-similarity/%s">See Similarity results in GreyNoise</a></h3><br/>%s is %d%% or more similar to %d other IP addresses in the GreyNoise.<br/>`,
		ipAddress, ipAddress, minimumScore, len(r.Similar),
	)
	inputIP.AddDisplayInfo("GreyNoise Similarity", displayHTML)
	return resp, nil
}
