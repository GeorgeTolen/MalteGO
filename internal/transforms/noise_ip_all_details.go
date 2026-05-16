package transforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupAllDetails struct{}

func (t *NoiseIPLookupAllDetails) Name() string { return "GreyNoiseNoiseIPLookupAllDetails" }

func (t *NoiseIPLookupAllDetails) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	displayHTML := fmt.Sprintf(
		"<b>IP:</b> %s<br/><b>Classification:</b> %s<br/><b>Actor:</b> %s<br/>"+
			"<b>First Seen:</b> %s<br/><b>Last Seen:</b> %s<br/>"+
			"<b>ASN:</b> %s<br/><b>Org:</b> %s<br/><b>Country:</b> %s<br/>"+
			"<b>Tags:</b> %s<br/><b>CVEs:</b> %s<br/><b>Spoofable:</b> %v<br/><b>Tor:</b> %v",
		r.IP, r.Classification, r.Actor,
		r.FirstSeen, r.LastSeen,
		r.Metadata.ASN, r.Metadata.Organization, r.Metadata.Country,
		strings.Join(r.Tags, ", "), strings.Join(r.CVEs, ", "),
		r.Spoofable, r.Metadata.Tor,
	)

	resp.AddEntity(maltego.EntityIPv4Address, r.IP).
		AddDisplayInfo("GreyNoise Context", displayHTML).
		AddProperty("classification", "Classification", maltego.MatchingRuleLoose, r.Classification).
		AddProperty("actor", "Actor", maltego.MatchingRuleLoose, r.Actor).
		AddProperty("first_seen", "First Seen", maltego.MatchingRuleLoose, r.FirstSeen).
		AddProperty("last_seen", "Last Seen", maltego.MatchingRuleLoose, r.LastSeen).
		AddProperty("asn", "ASN", maltego.MatchingRuleLoose, r.Metadata.ASN).
		AddProperty("organization", "Organization", maltego.MatchingRuleLoose, r.Metadata.Organization).
		AddProperty("country", "Country", maltego.MatchingRuleLoose, r.Metadata.Country).
		AddProperty("tags", "Tags", maltego.MatchingRuleLoose, strings.Join(r.Tags, ", ")).
		AddProperty("cves", "CVEs", maltego.MatchingRuleLoose, strings.Join(r.CVEs, ", ")).
		AddProperty("spoofable", "Spoofable", maltego.MatchingRuleLoose, fmt.Sprintf("%v", r.Spoofable)).
		AddProperty("tor", "Tor", maltego.MatchingRuleLoose, fmt.Sprintf("%v", r.Metadata.Tor)).
		AddProperty("os", "OS", maltego.MatchingRuleLoose, r.Metadata.OS)

	resp.Inform(fmt.Sprintf("Context lookup for %s complete", req.Value))
	return resp, nil
}
