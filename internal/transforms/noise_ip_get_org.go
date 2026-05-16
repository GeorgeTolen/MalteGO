package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetOrg struct{}

func (t *NoiseIPLookupGetOrg) Name() string { return "GreyNoiseNoiseIPLookupGetOrg" }

func (t *NoiseIPLookupGetOrg) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	org := r.Metadata.Organization
	if org == "" {
		org = r.Metadata.ASN
	}
	if org == "" {
		resp.Inform(fmt.Sprintf("No organization info for %s", req.Value))
		return resp, nil
	}

	displayHTML := fmt.Sprintf(
		"<b>Organization:</b> %s<br/><b>ASN:</b> %s<br/><b>Country:</b> %s<br/><b>Category:</b> %s",
		r.Metadata.Organization, r.Metadata.ASN, r.Metadata.Country, r.Metadata.Category,
	)

	resp.AddEntity(maltego.EntityOrganization, org).
		AddDisplayInfo("GreyNoise Organization", displayHTML).
		AddProperty("asn", "ASN", maltego.MatchingRuleLoose, r.Metadata.ASN).
		AddProperty("country", "Country", maltego.MatchingRuleLoose, r.Metadata.Country).
		AddProperty("category", "Category", maltego.MatchingRuleLoose, r.Metadata.Category).
		AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, r.IP)

	resp.Inform(fmt.Sprintf("Organization for %s: %s", req.Value, org))
	return resp, nil
}
