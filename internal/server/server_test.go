package server

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/greynoise-maltego/maltego-go/internal/config"
	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
	"github.com/greynoise-maltego/maltego-go/internal/transforms"
)

// ──────────────────────────────────────────────────────────────────────────────
// Test helpers
// ──────────────────────────────────────────────────────────────────────────────

func testConfig() *config.Config {
	return &config.Config{
		Port:            "8080",
		GinMode:         "test",
		GreyNoiseAPIKey: "server-test-key",
		RequestTimeout:  5 * time.Second,
	}
}

func testConfigNoKey() *config.Config {
	c := testConfig()
	c.GreyNoiseAPIKey = ""
	return c
}

// echoTransform responds with one entity containing the input IP value.
type echoTransform struct{}

func (e *echoTransform) Name() string { return "EchoTransform" }
func (e *echoTransform) Run(_ context.Context, _ greynoise.Client, req *maltego.Request) (*maltego.Response, error) {
	resp := maltego.NewResponse()
	resp.AddEntity(maltego.EntityIPv4Address, req.Value)
	resp.Inform("echo ok")
	return resp, nil
}

func buildServer(cfg *config.Config) *Server {
	reg := transforms.NewRegistry()
	reg.Register(&echoTransform{})
	// factory returns a mock client that never panics
	factory := func(apiKey string, _ time.Duration) greynoise.Client {
		return &greynoise.MockClient{
			CommunityIPFn: func(_ context.Context, ip string) (*greynoise.CommunityResponse, error) {
				return &greynoise.CommunityResponse{IP: ip, Noise: true}, nil
			},
		}
	}
	return newWithClientFactory(cfg, reg, factory, nil)
}

// maltegoXML builds a minimal valid Maltego XML request body.
func maltegoXML(entityType, value, apiKey string) string {
	key := ""
	if apiKey != "" {
		key = `<Field Name="greynoise.api.key">` + apiKey + `</Field>`
	}
	return `<?xml version="1.0"?>
<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="` + entityType + `">
        <Value>` + value + `</Value>
        <Weight>100</Weight>
      </Entity>
    </Entities>
    <TransformFields>` + key + `</TransformFields>
    <Limits SoftLimit="12" HardLimit="12"/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>`
}

// parseXMLResponse extracts entities and messages from a raw XML response.
type xmlTestResponse struct {
	Entities []struct {
		Type  string `xml:"Type,attr"`
		Value string `xml:"Value"`
	} `xml:"MaltegoTransformResponseMessage>Entities>Entity"`
	Messages []struct {
		Type  string `xml:"MessageType,attr"`
		Value string `xml:",chardata"`
	} `xml:"MaltegoTransformResponseMessage>UIMessages>UIMessage"`
}

func parseRespXML(t *testing.T, body string) xmlTestResponse {
	t.Helper()
	clean := strings.TrimPrefix(body, xml.Header)
	var r xmlTestResponse
	if err := xml.Unmarshal([]byte(clean), &r); err != nil {
		t.Fatalf("parse response XML: %v\nBody:\n%s", err, body)
	}
	return r
}

func hasFatalError(r xmlTestResponse) bool {
	for _, m := range r.Messages {
		if m.Type == "FatalError" {
			return true
		}
	}
	return false
}

func doRequest(t *testing.T, srv *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "text/xml")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	srv.router.ServeHTTP(w, req)
	return w
}

// ──────────────────────────────────────────────────────────────────────────────
// GET / — index
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleIndex_Returns200(t *testing.T) {
	srv := buildServer(testConfig())
	w := doRequest(t, srv, http.MethodGet, "/", "")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandleIndex_ReturnsJSON(t *testing.T) {
	srv := buildServer(testConfig())
	w := doRequest(t, srv, http.MethodGet, "/", "")

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v\nBody: %s", err, w.Body.String())
	}
}

func TestHandleIndex_ListsRegisteredTransforms(t *testing.T) {
	srv := buildServer(testConfig())
	w := doRequest(t, srv, http.MethodGet, "/", "")

	var resp struct {
		Transforms []string `json:"transforms"`
		Count      int      `json:"count"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Count != 1 {
		t.Errorf("count = %d, want 1", resp.Count)
	}
	if len(resp.Transforms) != 1 || resp.Transforms[0] != "EchoTransform" {
		t.Errorf("transforms = %v, want [EchoTransform]", resp.Transforms)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /run/:name — unknown transform
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_UnknownTransform_ReturnsXMLFatalError(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "1.2.3.4", "some-key")
	w := doRequest(t, srv, http.MethodPost, "/run/NoSuchTransform", body)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (Maltego protocol always 200)", w.Code)
	}
	r := parseRespXML(t, w.Body.String())
	if !hasFatalError(r) {
		t.Error("expected FatalError for unknown transform")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /run/:name — malformed / missing request body
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_InvalidXML_ReturnsFatalError(t *testing.T) {
	srv := buildServer(testConfig())
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", "<not valid xml")

	r := parseRespXML(t, w.Body.String())
	if !hasFatalError(r) {
		t.Error("expected FatalError for invalid XML body")
	}
}

func TestHandleTransform_EmptyBody_ReturnsFatalError(t *testing.T) {
	srv := buildServer(testConfig())
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", "")

	r := parseRespXML(t, w.Body.String())
	if !hasFatalError(r) {
		t.Error("expected FatalError for empty body")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /run/:name — missing API key
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_NoAPIKey_ReturnsFatalError(t *testing.T) {
	srv := buildServer(testConfigNoKey())
	// API key neither in config nor in request.
	body := maltegoXML(maltego.EntityIPv4Address, "1.2.3.4", "")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", body)

	r := parseRespXML(t, w.Body.String())
	if !hasFatalError(r) {
		t.Error("expected FatalError when no API key provided")
	}
}

func TestHandleTransform_APIKeyInRequest_OverridesConfig(t *testing.T) {
	// Config has no key, but request carries one — should succeed.
	srv := buildServer(testConfigNoKey())
	body := maltegoXML(maltego.EntityIPv4Address, "1.2.3.4", "request-key")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", body)

	r := parseRespXML(t, w.Body.String())
	if hasFatalError(r) {
		t.Error("expected success when API key is in request TransformFields")
	}
	if len(r.Entities) != 1 || r.Entities[0].Value != "1.2.3.4" {
		t.Errorf("unexpected entities: %v", r.Entities)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /run/:name — empty entity value
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_EmptyEntityValue_ReturnsFatalError(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "   ", "test-key")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", body)

	r := parseRespXML(t, w.Body.String())
	if !hasFatalError(r) {
		t.Error("expected FatalError for empty entity value")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /run/:name — successful transform
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_Success_ReturnsXMLWithEntities(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "8.8.8.8", "test-key")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", body)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/xml") {
		t.Errorf("Content-Type = %q, want text/xml", ct)
	}

	r := parseRespXML(t, w.Body.String())
	if len(r.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(r.Entities))
	}
	if r.Entities[0].Value != "8.8.8.8" {
		t.Errorf("entity value = %q, want 8.8.8.8", r.Entities[0].Value)
	}
}

func TestHandleTransform_Success_ResponseOmitsXMLHeader(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "1.1.1.1", "test-key")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform", body)

	if strings.HasPrefix(w.Body.String(), "<?xml") {
		t.Error("response should omit XML declaration to match Python maltego-trx output")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Trailing slash
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_TrailingSlash_Works(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "9.9.9.9", "key")
	w := doRequest(t, srv, http.MethodPost, "/run/EchoTransform/", body)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 with trailing slash", w.Code)
	}
	r := parseRespXML(t, w.Body.String())
	if len(r.Entities) != 1 || r.Entities[0].Value != "9.9.9.9" {
		t.Error("trailing slash path did not return expected entity")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Transform name case insensitivity
// ──────────────────────────────────────────────────────────────────────────────

func TestHandleTransform_CaseInsensitiveName(t *testing.T) {
	srv := buildServer(testConfig())
	body := maltegoXML(maltego.EntityIPv4Address, "1.2.3.4", "key")

	for _, name := range []string{"echotransform", "ECHOTRANSFORM", "EchoTransform"} {
		w := doRequest(t, srv, http.MethodPost, "/run/"+name, body)
		r := parseRespXML(t, w.Body.String())
		if hasFatalError(r) {
			t.Errorf("path /run/%s returned FatalError, expected case-insensitive match", name)
		}
	}
}
