package maltego

import (
	"testing"
)

// --- helpers ---

func mustParse(t *testing.T, xml string) *Request {
	t.Helper()
	req, err := ParseRequest([]byte(xml))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	return req
}

// --- test fixtures ---

const fullXML = `<?xml version="1.0" encoding="UTF-8"?>
<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>8.8.8.8</Value>
        <Weight>100</Weight>
        <AdditionalFields>
          <Field Name="ip">8.8.8.8</Field>
          <Field Name="country">US</Field>
        </AdditionalFields>
      </Entity>
    </Entities>
    <TransformFields>
      <Field Name="greynoise.api.key">test-key-123</Field>
      <Field Name="other.setting">value</Field>
    </TransformFields>
    <Limits SoftLimit="20" HardLimit="50"/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>`

const minimalXML = `<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>1.2.3.4</Value>
      </Entity>
    </Entities>
    <TransformFields/>
    <Limits/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>`

const noLimitsXML = `<MaltegoMessage>
  <MaltegoTransformRequestMessage>
    <Entities>
      <Entity Type="maltego.IPv4Address">
        <Value>5.5.5.5</Value>
      </Entity>
    </Entities>
    <TransformFields/>
  </MaltegoTransformRequestMessage>
</MaltegoMessage>`

// --- ParseRequest tests ---

func TestParseRequest_FullXML(t *testing.T) {
	req := mustParse(t, fullXML)

	if req.Value != "8.8.8.8" {
		t.Errorf("Value = %q, want 8.8.8.8", req.Value)
	}
	if req.EntityType != "maltego.IPv4Address" {
		t.Errorf("EntityType = %q, want maltego.IPv4Address", req.EntityType)
	}
	if req.Properties["ip"] != "8.8.8.8" {
		t.Errorf("Properties[ip] = %q, want 8.8.8.8", req.Properties["ip"])
	}
	if req.Properties["country"] != "US" {
		t.Errorf("Properties[country] = %q, want US", req.Properties["country"])
	}
	if req.Settings["greynoise.api.key"] != "test-key-123" {
		t.Errorf("Settings[greynoise.api.key] = %q, want test-key-123", req.Settings["greynoise.api.key"])
	}
	if req.SoftLimit != 20 {
		t.Errorf("SoftLimit = %d, want 20", req.SoftLimit)
	}
	if req.HardLimit != 50 {
		t.Errorf("HardLimit = %d, want 50", req.HardLimit)
	}
}

func TestParseRequest_MinimalXML_DefaultLimits(t *testing.T) {
	req := mustParse(t, minimalXML)

	if req.Value != "1.2.3.4" {
		t.Errorf("Value = %q, want 1.2.3.4", req.Value)
	}
	// Zero limits in XML → default to 12.
	if req.SoftLimit != 12 {
		t.Errorf("SoftLimit = %d, want 12 (default)", req.SoftLimit)
	}
	if req.HardLimit != 12 {
		t.Errorf("HardLimit = %d, want 12 (default)", req.HardLimit)
	}
}

func TestParseRequest_NoLimitsElement_DefaultLimits(t *testing.T) {
	req := mustParse(t, noLimitsXML)
	if req.SoftLimit != 12 {
		t.Errorf("SoftLimit = %d, want 12 (default)", req.SoftLimit)
	}
	if req.HardLimit != 12 {
		t.Errorf("HardLimit = %d, want 12 (default)", req.HardLimit)
	}
}

func TestParseRequest_EmptyAdditionalFields(t *testing.T) {
	req := mustParse(t, minimalXML)
	if len(req.Properties) != 0 {
		t.Errorf("Properties = %v, want empty map", req.Properties)
	}
}

func TestParseRequest_EmptyTransformFields(t *testing.T) {
	req := mustParse(t, minimalXML)
	if len(req.Settings) != 0 {
		t.Errorf("Settings = %v, want empty map", req.Settings)
	}
}

func TestParseRequest_InvalidXML_ReturnsError(t *testing.T) {
	_, err := ParseRequest([]byte(`<not valid xml`))
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestParseRequest_EmptyBody_ReturnsError(t *testing.T) {
	_, err := ParseRequest([]byte(``))
	if err == nil {
		t.Error("expected error for empty body, got nil")
	}
}

func TestParseRequest_WrongRootElement_ReturnsError(t *testing.T) {
	// xml.Unmarshal is lenient on root tag but structure won't match; Value will be empty.
	req, err := ParseRequest([]byte(`<SomethingElse><Foo/></SomethingElse>`))
	if err != nil {
		// also acceptable to return error
		return
	}
	if req.Value != "" {
		t.Errorf("Value = %q from wrong root XML, want empty", req.Value)
	}
}

// --- APIKey tests ---

func TestAPIKey_FromSettings(t *testing.T) {
	req := mustParse(t, fullXML)
	if got := req.APIKey("fallback-key"); got != "test-key-123" {
		t.Errorf("APIKey = %q, want test-key-123", got)
	}
}

func TestAPIKey_FromUpstreamGNApiKeySetting(t *testing.T) {
	req := &Request{Settings: map[string]string{"GNApiKey": "test-key-123"}}
	if got := req.APIKey("fallback-key"); got != "test-key-123" {
		t.Errorf("APIKey = %q, want test-key-123", got)
	}
}

func TestAPIKey_FallbackWhenNotInSettings(t *testing.T) {
	req := mustParse(t, minimalXML)
	if got := req.APIKey("env-key"); got != "env-key" {
		t.Errorf("APIKey = %q, want env-key", got)
	}
}

func TestAPIKey_EmptyFallback_ReturnsEmpty(t *testing.T) {
	req := mustParse(t, minimalXML)
	if got := req.APIKey(""); got != "" {
		t.Errorf("APIKey = %q, want empty string", got)
	}
}
