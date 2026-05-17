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
	addInputEntity(resp, req, maltego.EntityPerson)

	fromDate, toDate := queryDateRange(req)
	query := fmt.Sprintf("actor:%s last_seen:[%s TO %s]", req.Value, fromDate, toDate)
	if asn := req.Settings["asn"]; asn != "" {
		query += " asn:AS" + asn
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
