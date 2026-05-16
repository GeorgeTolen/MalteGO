package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type QueryByActor struct{}

func (t *QueryByActor) Name() string { return "GreyNoiseQueryByActor" }

func (t *QueryByActor) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	query := fmt.Sprintf("actor:%s", req.Value)
	r, err := client.GNQL(ctx, query, req.HardLimit)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GNQL query failed: %v", err))
		return resp, nil
	}

	if len(r.Data) == 0 {
		resp.Inform(fmt.Sprintf("No IPs found for actor %q", req.Value))
		return resp, nil
	}

	for _, entry := range r.Data {
		resp.AddEntity(maltego.EntityIPv4Address, entry.IP).
			AddProperty("classification", "Classification", maltego.MatchingRuleLoose, entry.Classification).
			AddProperty("actor", "Actor", maltego.MatchingRuleLoose, entry.Actor).
			AddProperty("country", "Country", maltego.MatchingRuleLoose, entry.Country).
			AddProperty("last_seen", "Last Seen", maltego.MatchingRuleLoose, entry.LastSeen)
	}

	resp.Inform(fmt.Sprintf("Found %d IP(s) for actor %q", len(r.Data), req.Value))
	return resp, nil
}
