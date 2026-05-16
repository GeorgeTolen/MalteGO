package transforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type QueryByASN struct{}

func (t *QueryByASN) Name() string { return "GreyNoiseQueryByASN" }

func (t *QueryByASN) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	asn := strings.TrimSpace(req.Value)
	// Normalise: GNQL expects "AS12345" format.
	if !strings.HasPrefix(strings.ToUpper(asn), "AS") {
		asn = "AS" + asn
	}

	query := fmt.Sprintf("metadata.asn:%s", asn)
	r, err := client.GNQL(ctx, query, req.HardLimit)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GNQL query failed: %v", err))
		return resp, nil
	}

	if len(r.Data) == 0 {
		resp.Inform(fmt.Sprintf("No results for ASN %s", asn))
		return resp, nil
	}

	for _, entry := range r.Data {
		resp.AddEntity(maltego.EntityIPv4Address, entry.IP).
			AddProperty("classification", "Classification", maltego.MatchingRuleLoose, entry.Classification).
			AddProperty("actor", "Actor", maltego.MatchingRuleLoose, entry.Actor).
			AddProperty("asn", "ASN", maltego.MatchingRuleLoose, entry.ASN).
			AddProperty("last_seen", "Last Seen", maltego.MatchingRuleLoose, entry.LastSeen)
	}

	resp.Inform(fmt.Sprintf("Found %d IP(s) for ASN %s", len(r.Data), asn))
	return resp, nil
}
