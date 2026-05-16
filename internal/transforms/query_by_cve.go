package transforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type QueryByCVE struct{}

func (t *QueryByCVE) Name() string { return "GreyNoiseQueryByCVE" }

func (t *QueryByCVE) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	cve := strings.ToUpper(strings.TrimSpace(req.Value))
	query := fmt.Sprintf("cve:%s", cve)
	r, err := client.GNQL(ctx, query, req.HardLimit)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GNQL query failed: %v", err))
		return resp, nil
	}

	if len(r.Data) == 0 {
		resp.Inform(fmt.Sprintf("No IPs found exploiting %s", cve))
		return resp, nil
	}

	for _, entry := range r.Data {
		resp.AddEntity(maltego.EntityIPv4Address, entry.IP).
			AddProperty("classification", "Classification", maltego.MatchingRuleLoose, entry.Classification).
			AddProperty("actor", "Actor", maltego.MatchingRuleLoose, entry.Actor).
			AddProperty("country", "Country", maltego.MatchingRuleLoose, entry.Country).
			AddProperty("organization", "Organization", maltego.MatchingRuleLoose, entry.Organization).
			AddProperty("last_seen", "Last Seen", maltego.MatchingRuleLoose, entry.LastSeen)
	}

	resp.Inform(fmt.Sprintf("Found %d IP(s) exploiting %s", len(r.Data), cve))
	return resp, nil
}
