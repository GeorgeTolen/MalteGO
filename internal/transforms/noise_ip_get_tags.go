package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetTags struct{}

func (t *NoiseIPLookupGetTags) Name() string { return "GreyNoiseNoiseIPLookupGetTags" }

func (t *NoiseIPLookupGetTags) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	if len(r.Tags) == 0 {
		resp.Inform(fmt.Sprintf("No tags found for %s", req.Value))
		return resp, nil
	}

	for _, tag := range r.Tags {
		resp.AddEntity(maltego.EntityHashtag, tag).
			AddDisplayInfo("GreyNoise Tag", fmt.Sprintf("<b>Tag:</b> %s<br/><b>IP:</b> %s", tag, r.IP)).
			AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, r.IP).
			AddProperty("classification", "Classification", maltego.MatchingRuleLoose, r.Classification)
	}

	resp.Inform(fmt.Sprintf("Found %d tag(s) for %s", len(r.Tags), req.Value))
	return resp, nil
}
