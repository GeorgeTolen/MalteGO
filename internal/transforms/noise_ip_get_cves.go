package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetCVEs struct{}

func (t *NoiseIPLookupGetCVEs) Name() string { return "GreyNoiseNoiseIPLookupGetCVEs" }

func (t *NoiseIPLookupGetCVEs) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	if len(r.CVEs) == 0 {
		resp.Inform(fmt.Sprintf("No CVEs associated with %s", req.Value))
		return resp, nil
	}

	for _, cve := range r.CVEs {
		resp.AddEntity(maltego.EntityCVE, cve).
			AddDisplayInfo("GreyNoise CVE", fmt.Sprintf("<b>CVE:</b> %s<br/><b>Exploited by IP:</b> %s", cve, r.IP)).
			AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, r.IP)
	}

	resp.Inform(fmt.Sprintf("Found %d CVE(s) for %s", len(r.CVEs), req.Value))
	return resp, nil
}
