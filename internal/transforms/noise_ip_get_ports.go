package transforms

import (
	"context"
	"fmt"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type NoiseIPLookupGetPorts struct{}

func (t *NoiseIPLookupGetPorts) Name() string { return "GreyNoiseNoiseIPLookupGetPorts" }

func (t *NoiseIPLookupGetPorts) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
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

	// Use raw scan data if available; fall back to ports list.
	if len(r.RawData.Scan) > 0 {
		for _, scan := range r.RawData.Scan {
			portStr := fmt.Sprintf("%d", scan.Port)
			resp.AddEntity(maltego.EntityPort, portStr).
				AddDisplayInfo("GreyNoise Port", fmt.Sprintf("<b>Port:</b> %d<br/><b>Protocol:</b> %s<br/><b>IP:</b> %s", scan.Port, scan.Protocol, r.IP)).
				AddProperty("protocol", "Protocol", maltego.MatchingRuleLoose, scan.Protocol).
				AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, r.IP)
		}
		resp.Inform(fmt.Sprintf("Found %d port(s) for %s", len(r.RawData.Scan), req.Value))
		return resp, nil
	}

	if len(r.Ports) == 0 {
		resp.Inform(fmt.Sprintf("No ports found for %s", req.Value))
		return resp, nil
	}

	for _, port := range r.Ports {
		portStr := fmt.Sprintf("%d", port)
		resp.AddEntity(maltego.EntityPort, portStr).
			AddDisplayInfo("GreyNoise Port", fmt.Sprintf("<b>Port:</b> %d<br/><b>IP:</b> %s", port, r.IP)).
			AddProperty("source_ip", "Source IP", maltego.MatchingRuleLoose, r.IP)
	}

	resp.Inform(fmt.Sprintf("Found %d port(s) for %s", len(r.Ports), req.Value))
	return resp, nil
}
