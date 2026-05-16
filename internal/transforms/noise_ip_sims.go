package transforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPSims struct{}

func (t *NoiseIPSims) Name() string { return "GreyNoiseNoiseIPSims" }

func (t *NoiseIPSims) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	r, err := client.SimilarIPs(ctx, req.Value)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GreyNoise similarity lookup failed: %v", err))
		return resp, nil
	}

	if len(r.Similar) == 0 {
		resp.Inform(fmt.Sprintf("No similar IPs found for %s", req.Value))
		return resp, nil
	}

	limit := req.HardLimit
	if limit <= 0 || limit > len(r.Similar) {
		limit = len(r.Similar)
	}

	for _, sim := range r.Similar[:limit] {
		displayHTML := fmt.Sprintf(
			"<b>IP:</b> %s<br/><b>Similarity Score:</b> %.2f<br/><b>Actor:</b> %s<br/><b>Features:</b> %s",
			sim.IP, sim.Score, sim.Actor, strings.Join(sim.Similarity, ", "),
		)
		resp.AddEntity(maltego.EntityIPv4Address, sim.IP).
			AddDisplayInfo("GreyNoise Similar IP", displayHTML).
			AddProperty("similarity_score", "Similarity Score", maltego.MatchingRuleLoose, fmt.Sprintf("%.4f", sim.Score)).
			AddProperty("actor", "Actor", maltego.MatchingRuleLoose, sim.Actor).
			AddProperty("features", "Shared Features", maltego.MatchingRuleLoose, strings.Join(sim.Similarity, ", ")).
			AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, req.Value)
	}

	resp.Inform(fmt.Sprintf("Found %d similar IP(s) for %s", limit, req.Value))
	return resp, nil
}
