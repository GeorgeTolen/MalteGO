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
	inputIP := addInputEntity(resp, req, maltego.EntityIPv4Address)

	r, err := client.ContextIP(ctx, req.Value)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if !r.Seen {
		resp.AddEntity("greynoise.noise", "No Noise Detected")
		resp.Inform("The IP address " + req.Value + " hasn't been seen by GreyNoise.")
		addNoiseDisplayInfo(inputIP, r.Classification, r.LastSeen, contextLink(r), r.Actor, r.Tags)
		return resp, nil
	}

	resp.AddEntity("greynoise.noise", "Noise Detected")
	if r.Actor != "" && r.Actor != "unknown" {
		resp.AddEntity(maltego.EntityPerson, r.Actor)
	}
	if r.Classification != "" {
		resp.AddEntity("greynoise.classification", r.Classification)
	}
	if r.Metadata.ASN != "" {
		resp.AddEntity(maltego.EntityAS, strings.TrimPrefix(r.Metadata.ASN, "AS"))
	}
	if r.Metadata.Organization != "" {
		resp.AddEntity(maltego.EntityCompany, r.Metadata.Organization)
	}
	if r.Metadata.City != "" && r.Metadata.Country != "" && r.Metadata.CountryCode != "" {
		resp.AddEntity(maltego.EntityLocation, fmt.Sprintf("%s, %s (%s)", r.Metadata.City, r.Metadata.Country, r.Metadata.CountryCode))
	}
	if r.VPN {
		resp.AddEntity(maltego.EntityService, "VPN Service: "+r.VPNService)
	}
	if r.Bot {
		resp.AddEntity(maltego.EntityService, "Common Bot Activity")
	}
	if r.Metadata.Tor {
		resp.AddEntity(maltego.EntityService, "Tor Exit Node")
	}
	for _, cve := range r.CVEs {
		resp.AddEntity(maltego.EntityCVE, cve).
			AddProperty("link#maltego.link.label", "Label", maltego.MatchingRuleLoose, "Probes For")
	}
	for _, item := range r.RawData.Scan {
		resp.AddEntity(maltego.EntityPort, fmt.Sprintf("%d", item.Port)).
			AddProperty("link#maltego.link.label", "Label", maltego.MatchingRuleLoose, "Scans For")
	}
	for _, tag := range r.Tags {
		resp.AddEntity(maltego.EntityPhrase, tag)
	}

	link := contextLink(r)
	inputIP.AddProperty("gn_url", "GreyNoise URL", maltego.MatchingRuleLoose, link).
		AddProperty("gn_last_seen", "GreyNoise last seen", maltego.MatchingRuleLoose, r.LastSeen)
	addNoiseDisplayInfo(inputIP, r.Classification, r.LastSeen, link, r.Actor, r.Tags)
	return resp, nil
}
