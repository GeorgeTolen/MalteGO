package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type RIOTIPLookup struct{}

func (t *RIOTIPLookup) Name() string { return "GreyNoiseRIOTIPLookup" }

func (t *RIOTIPLookup) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()

	r, err := client.RIOT(ctx, req.Value)
	if err != nil {
		resp.FatalError(fmt.Sprintf("GreyNoise RIOT lookup failed: %v", err))
		return resp, nil
	}

	if !r.Riot {
		resp.Inform(fmt.Sprintf("%s is not in the GreyNoise RIOT dataset", req.Value))
		return resp, nil
	}

	displayHTML := fmt.Sprintf(
		"<b>IP:</b> %s<br/><b>RIOT:</b> %v<br/><b>Name:</b> %s<br/><b>Category:</b> %s<br/>"+
			"<b>Trust Level:</b> %s<br/><b>Description:</b> %s<br/><b>Last Updated:</b> %s",
		r.IP, r.Riot, r.Name, r.Category, r.TrustLevel, r.Description, r.LastUpdated,
	)

	resp.AddEntity(maltego.EntityIPv4Address, r.IP).
		AddDisplayInfo("GreyNoise RIOT", displayHTML).
		AddProperty("riot", "RIOT", maltego.MatchingRuleLoose, fmt.Sprintf("%v", r.Riot)).
		AddProperty("name", "Service Name", maltego.MatchingRuleLoose, r.Name).
		AddProperty("category", "Category", maltego.MatchingRuleLoose, r.Category).
		AddProperty("trust_level", "Trust Level", maltego.MatchingRuleLoose, r.TrustLevel).
		AddProperty("description", "Description", maltego.MatchingRuleLoose, r.Description).
		AddProperty("reference", "Reference", maltego.MatchingRuleLoose, r.Reference).
		AddProperty("last_updated", "Last Updated", maltego.MatchingRuleLoose, r.LastUpdated)

	resp.Inform(fmt.Sprintf("RIOT lookup for %s: %s (%s)", req.Value, r.Name, r.Category))
	return resp, nil
}
