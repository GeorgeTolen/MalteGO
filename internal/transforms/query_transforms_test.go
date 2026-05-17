package transforms

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

// gnqlMock returns a MockClient that records the last query string and
// responds with the provided response/error.
type gnqlCapture struct {
	lastQuery string
	lastSize  int
}

func gnqlMock(cap *gnqlCapture, r *greynoise.GNQLResponse, err error) *greynoise.MockClient {
	return &greynoise.MockClient{
		GNQLFn: func(_ context.Context, query string, size int) (*greynoise.GNQLResponse, error) {
			cap.lastQuery = query
			cap.lastSize = size
			if r != nil && r.Count == 0 && len(r.Data) > 0 {
				cp := *r
				cp.Count = len(r.Data)
				return &cp, err
			}
			return r, err
		},
	}
}

func gnqlEntries(ips ...string) []greynoise.GNQLEntry {
	entries := make([]greynoise.GNQLEntry, len(ips))
	for i, ip := range ips {
		entries[i] = greynoise.GNQLEntry{
			IP:             ip,
			Classification: "malicious",
			Actor:          "TestActor",
		}
	}
	return entries
}

// ──────────────────────────────────────────────────────────────────────────────
// QueryByASN
// ──────────────────────────────────────────────────────────────────────────────

func TestQueryByASN_Name(t *testing.T) {
	if got := (&QueryByASN{}).Name(); got != "GreyNoiseQueryByASN" {
		t.Errorf("Name() = %q", got)
	}
}

func TestQueryByASN_NormalisesPrefix_WithAS(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: gnqlEntries("1.1.1.1")}, nil)
	runTransform(t, &QueryByASN{}, mock, makeReq("AS1234"))
	if !strings.Contains(cap.lastQuery, "ASAS1234") {
		t.Errorf("query %q should match upstream AS prefix behaviour", cap.lastQuery)
	}
}

func TestQueryByASN_NormalisesPrefix_WithoutAS(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: gnqlEntries("1.1.1.1")}, nil)
	runTransform(t, &QueryByASN{}, mock, makeReq("1234"))
	if !strings.Contains(cap.lastQuery, "AS1234") {
		t.Errorf("query %q should contain AS1234 after normalisation", cap.lastQuery)
	}
}

func TestQueryByASN_ReturnsIPEntities(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{
		Data: gnqlEntries("10.0.0.1", "10.0.0.2", "10.0.0.3"),
	}, nil)

	px := runTransform(t, &QueryByASN{}, mock, makeReq("AS9000"))

	if len(px.Entities) != 4 {
		t.Fatalf("expected input entity and 3 IP entities, got %d", len(px.Entities))
	}
	for _, e := range px.Entities[1:] {
		if e.Type != maltego.EntityIPv4Address {
			t.Errorf("entity type = %q, want maltego.IPv4Address", e.Type)
		}
	}
}

func TestQueryByASN_EmptyResult_ReturnsInform(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: nil}, nil)
	px := runTransform(t, &QueryByASN{}, mock, makeReq("AS0"))
	assertInformWithInputEntity(t, px)
}

func TestQueryByASN_APIError_ReturnsFatalError(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, nil, errors.New("gnql error"))
	px := runTransform(t, &QueryByASN{}, mock, makeReq("AS123"))
	assertInformWithInputEntity(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// QueryByActor
// ──────────────────────────────────────────────────────────────────────────────

func TestQueryByActor_Name(t *testing.T) {
	if got := (&QueryByActor{}).Name(); got != "GreyNoiseQueryByActor" {
		t.Errorf("Name() = %q", got)
	}
}

func TestQueryByActor_QueryContainsActor(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: gnqlEntries("1.1.1.1")}, nil)
	runTransform(t, &QueryByActor{}, mock, makeReq("Mirai"))
	if !strings.Contains(cap.lastQuery, "Mirai") {
		t.Errorf("query %q does not contain actor name", cap.lastQuery)
	}
}

func TestQueryByActor_ReturnsIPEntities(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{
		Data: gnqlEntries("2.2.2.2", "3.3.3.3"),
	}, nil)

	px := runTransform(t, &QueryByActor{}, mock, makeReq("SomeActor"))
	if len(px.Entities) != 3 {
		t.Fatalf("expected input entity and 2 IP entities, got %d", len(px.Entities))
	}
}

func TestQueryByActor_EmptyResult_ReturnsInform(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: nil}, nil)
	px := runTransform(t, &QueryByActor{}, mock, makeReq("Unknown"))
	assertInformWithInputEntity(t, px)
}

func TestQueryByActor_APIError_ReturnsFatalError(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, nil, errors.New("actor lookup failed"))
	px := runTransform(t, &QueryByActor{}, mock, makeReq("BadActor"))
	assertInformWithInputEntity(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// QueryByCVE
// ──────────────────────────────────────────────────────────────────────────────

func TestQueryByCVE_Name(t *testing.T) {
	if got := (&QueryByCVE{}).Name(); got != "GreyNoiseQueryByCVE" {
		t.Errorf("Name() = %q", got)
	}
}

func TestQueryByCVE_RejectsLowercase(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: gnqlEntries("1.1.1.1")}, nil)
	px := runTransform(t, &QueryByCVE{}, mock, makeReq("cve-2021-44228"))
	if cap.lastQuery != "" {
		t.Errorf("query %q should not run for lowercase CVE", cap.lastQuery)
	}
	if len(px.Messages) != 1 || !strings.Contains(px.Messages[0].Value, "not a properly formatted CVE") {
		t.Errorf("expected malformed CVE message, got %#v", px.Messages)
	}
}

func TestQueryByCVE_ReturnsExploitingIPs(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{
		Data: gnqlEntries("4.4.4.4", "5.5.5.5"),
	}, nil)

	px := runTransform(t, &QueryByCVE{}, mock, makeReq("CVE-2022-0001"))
	if len(px.Entities) != 3 {
		t.Fatalf("expected input entity and 2 IPs exploiting CVE, got %d", len(px.Entities))
	}
}

func TestQueryByCVE_EmptyResult_ReturnsInform(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: nil}, nil)
	px := runTransform(t, &QueryByCVE{}, mock, makeReq("CVE-2099-0000"))
	assertInformWithInputEntity(t, px)
}

func TestQueryByCVE_APIError_ReturnsFatalError(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, nil, errors.New("cve query failed"))
	px := runTransform(t, &QueryByCVE{}, mock, makeReq("CVE-2021-44228"))
	assertInformWithInputEntity(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// QueryByTag
// ──────────────────────────────────────────────────────────────────────────────

func TestQueryByTag_Name(t *testing.T) {
	if got := (&QueryByTag{}).Name(); got != "GreyNoiseQueryByTag" {
		t.Errorf("Name() = %q", got)
	}
}

func TestQueryByTag_QueryContainsTag(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: gnqlEntries("1.1.1.1")}, nil)
	runTransform(t, &QueryByTag{}, mock, makeReq("mirai"))
	if !strings.Contains(cap.lastQuery, "mirai") {
		t.Errorf("query %q does not contain tag", cap.lastQuery)
	}
}

func TestQueryByTag_ReturnsTaggedIPs(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{
		Data: gnqlEntries("7.7.7.7", "8.8.8.8", "9.9.9.9"),
	}, nil)

	px := runTransform(t, &QueryByTag{}, mock, makeReq("scanner"))
	if len(px.Entities) != 4 {
		t.Fatalf("expected input entity and 3 entities, got %d", len(px.Entities))
	}
}

func TestQueryByTag_EmptyResult_ReturnsInform(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: nil}, nil)
	px := runTransform(t, &QueryByTag{}, mock, makeReq("nonexistent-tag"))
	assertInformWithInputEntity(t, px)
}

func TestQueryByTag_APIError_ReturnsFatalError(t *testing.T) {
	cap := &gnqlCapture{}
	mock := gnqlMock(cap, nil, errors.New("tag query error"))
	px := runTransform(t, &QueryByTag{}, mock, makeReq("mirai"))
	assertInformWithInputEntity(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// Cross-transform: HardLimit is forwarded to GNQL
// ──────────────────────────────────────────────────────────────────────────────

func TestQueryTransforms_HardLimitForwardedToGNQL(t *testing.T) {
	for _, tr := range []Transform{&QueryByASN{}, &QueryByActor{}, &QueryByCVE{}, &QueryByTag{}} {
		tr := tr
		t.Run(tr.Name(), func(t *testing.T) {
			cap := &gnqlCapture{}
			mock := gnqlMock(cap, &greynoise.GNQLResponse{Data: nil}, nil)
			value := "test"
			if tr.Name() == "GreyNoiseQueryByCVE" {
				value = "CVE-2024-0001"
			}
			req := makeReq(value)
			req.HardLimit = 25
			runTransform(t, tr, mock, req)
			if cap.lastSize != 25 {
				t.Errorf("%s: GNQL size = %d, want 25 (HardLimit)", tr.Name(), cap.lastSize)
			}
		})
	}
}
