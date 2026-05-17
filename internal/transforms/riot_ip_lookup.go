package transforms

import (
	"context"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type RIOTIPLookup struct{}

func (t *RIOTIPLookup) Name() string { return "GreyNoiseRIOTIPLookup" }

func (t *RIOTIPLookup) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	inputIP := addInputEntity(resp, req, maltego.EntityIPv4Address)

	r, err := client.RIOT(ctx, req.Value)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if !r.Riot {
		resp.AddEntity("greynoise.noise", "Not a Common Business Service")
		resp.Inform("The IP address " + req.Value + " is not found in GreyNoise RIOT IP Dataset.")
		addRIOTDisplayInfo(inputIP, "", r.LastUpdated, "", r.Name)
		return resp, nil
	}

	resp.AddEntity("greynoise.noise", "Common Business Service Detected")
	if r.Name != "" && r.Name != "unknown" {
		resp.AddEntity(maltego.EntityOrganization, r.Name)
	}

	classification := "RIOT"
	if r.TrustLevel == "1" {
		classification = "RIOT - Reasonably Ignore"
	} else if r.TrustLevel == "2" {
		classification = "RIOT - Commonly Seen"
	}
	resp.AddEntity("greynoise.classification", classification)

	link := "https://www.greynoise.io/viz/riot/" + r.IP
	inputIP.AddProperty("gn_url", "GreyNoise URL", maltego.MatchingRuleLoose, link).
		AddProperty("gn_last_updated", "GreyNoise last updated", maltego.MatchingRuleLoose, r.LastUpdated)
	addRIOTDisplayInfo(inputIP, "", r.LastUpdated, link, r.Name)
	return resp, nil
}
