package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type CommunityIPLookup struct{}

func (t *CommunityIPLookup) Name() string { return "GreyNoiseCommunityIPLookup" }

func (t *CommunityIPLookup) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	r, err := client.CommunityIP(ctx, req.Value)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GreyNoise Community lookup failed: %v", err))
		return resp, nil
	}

	if r.Message != "" && !r.Noise && !r.Riot {
		resp.Inform(r.Message)
		return resp, nil
	}

	displayHTML := fmt.Sprintf(
		"<b>IP:</b> %s<br/><b>Noise:</b> %v<br/><b>RIOT:</b> %v<br/><b>Classification:</b> %s<br/><b>Actor:</b> %s<br/><b>Last Seen:</b> %s",
		r.IP, r.Noise, r.Riot, r.Classification, r.Name, r.LastSeen,
	)

	resp.AddEntity(maltego.EntityIPv4Address, r.IP).
		AddDisplayInfo("GreyNoise Community", displayHTML).
		AddProperty("noise", "Noise", maltego.MatchingRuleLoose, fmt.Sprintf("%v", r.Noise)).
		AddProperty("riot", "RIOT", maltego.MatchingRuleLoose, fmt.Sprintf("%v", r.Riot)).
		AddProperty("classification", "Classification", maltego.MatchingRuleLoose, r.Classification).
		AddProperty("actor", "Actor", maltego.MatchingRuleLoose, r.Name).
		AddProperty("last_seen", "Last Seen", maltego.MatchingRuleLoose, r.LastSeen).
		AddProperty("link", "GreyNoise Link", maltego.MatchingRuleLoose, r.Link)

	resp.Inform(fmt.Sprintf("Community lookup for %s complete", req.Value))
	return resp, nil
}
