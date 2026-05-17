package transforms

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

type QueryByCVE struct{}

func (t *QueryByCVE) Name() string { return "GreyNoiseQueryByCVE" }

func (t *QueryByCVE) Run(ctx context.Context, client greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	addInputEntity(resp, req, maltego.EntityCVE)

	cve := strings.TrimSpace(req.Value)
	if !regexp.MustCompile(`^CVE-\d{4}-\d{4,7}$`).MatchString(cve) {
		resp.Inform(req.Value + " is not a properly formatted CVE.")
		return resp, nil
	}

	fromDate, toDate := queryDateRange(req)
	query := fmt.Sprintf("cve:%s last_seen:[%s TO %s]", cve, fromDate, toDate)
	if asn := req.Settings["asn"]; asn != "" {
		query += " asn:AS" + asn
	}
	if actor := req.Settings["actor"]; actor != "" {
		query += " actor:'" + actor + "'"
	}
	if port := req.Settings["port"]; port != "" && port != "0" {
		query += " raw_data.scan.port:" + port
	}
	r, err := client.GNQL(ctx, query, req.HardLimit)
	if err != nil {
		resp.Inform(err.Error())
		return resp, nil
	}

	if r.Count <= 1 {
		resp.Inform("The Query " + query + " did not return any results.")
		return resp, nil
	}

	for _, entry := range r.Data {
		resp.AddEntity(maltego.EntityIPv4Address, entry.IP)
	}

	return resp, nil
}
