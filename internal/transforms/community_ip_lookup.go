package transforms

import (
	"context"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type CommunityIPLookup struct{}

func (t *CommunityIPLookup) Name() string { return "GreyNoiseCommunityIPLookup" }

func (t *CommunityIPLookup) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	inputIP := addInputEntity(resp, req, maltego.EntityIPv4Address)

	r, err := client.CommunityIP(ctx, req.Value)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if r.Noise || r.Riot {
		if r.Noise {
			resp.AddEntity("greynoise.noise", "Noise Detected")
		}
		if r.Riot {
			resp.AddEntity("greynoise.noise", "Common Business Detected")
		}

		if r.Name != "" && r.Name != "unknown" {
			resp.AddEntity(maltego.EntityOrganization, r.Name)
		}

		if r.Classification != "" {
			resp.AddEntity("greynoise.classification", r.Classification)
		}

		inputIP.AddProperty("gn_url", "GreyNoise URL", maltego.MatchingRuleLoose, r.Link).
			AddProperty("gn_last_seen", "GreyNoise last seen", maltego.MatchingRuleLoose, r.LastSeen)
	} else {
		resp.AddEntity("greynoise.noise", "No Noise Detected")
		resp.Inform("The IP address " + req.Value + " hasn't been seen by GreyNoise.")
	}

	addCommunityDisplayInfo(inputIP, r.Classification, r.LastSeen, r.Link, r.Name)
	return resp, nil
}
