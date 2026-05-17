package transforms

import (
	"context"
	"errors"
	"testing"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

// ──────────────────────────────────────────────────────────────────────────────
// CommunityIPLookup
// ──────────────────────────────────────────────────────────────────────────────

func communityMock(r *greynoise.CommunityResponse, err error) *greynoise.MockClient {
	return &greynoise.MockClient{
		CommunityIPFn: func(_ context.Context, _ string) (*greynoise.CommunityResponse, error) {
			return r, err
		},
	}
}

func TestCommunityIPLookup_Name(t *testing.T) {
	if got := (&CommunityIPLookup{}).Name(); got != "GreyNoiseCommunityIPLookup" {
		t.Errorf("Name() = %q", got)
	}
}

func TestCommunityIPLookup_Success_NoisyMalicious(t *testing.T) {
	mock := communityMock(&greynoise.CommunityResponse{
		IP:             "1.2.3.4",
		Noise:          true,
		Riot:           false,
		Classification: "malicious",
		Name:           "BadActor",
		LastSeen:       "2024-01-01",
	}, nil)

	px := runTransform(t, &CommunityIPLookup{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 4 {
		t.Fatalf("expected 4 entities, got %d", len(px.Entities))
	}
	e := px.Entities[0]
	if e.Type != maltego.EntityIPv4Address {
		t.Errorf("entity Type = %q", e.Type)
	}
	if e.Value != "1.2.3.4" {
		t.Errorf("entity Value = %q, want 1.2.3.4", e.Value)
	}
	if e.Properties["gn_last_seen"] != "2024-01-01" {
		t.Errorf("gn_last_seen = %q, want 2024-01-01", e.Properties["gn_last_seen"])
	}
	if px.Entities[1].Type != "greynoise.noise" || px.Entities[1].Value != "Noise Detected" {
		t.Errorf("noise entity = %#v", px.Entities[1])
	}
	if px.Entities[2].Type != maltego.EntityOrganization || px.Entities[2].Value != "BadActor" {
		t.Errorf("organization entity = %#v", px.Entities[2])
	}
	if px.Entities[3].Type != "greynoise.classification" || px.Entities[3].Value != "malicious" {
		t.Errorf("classification entity = %#v", px.Entities[3])
	}
}

func TestCommunityIPLookup_Success_RIOTBenign(t *testing.T) {
	mock := communityMock(&greynoise.CommunityResponse{
		IP:             "8.8.8.8",
		Noise:          false,
		Riot:           true,
		Classification: "benign",
		Name:           "Google Public DNS",
	}, nil)

	px := runTransform(t, &CommunityIPLookup{}, mock, makeReq("8.8.8.8"))

	if len(px.Entities) != 4 {
		t.Fatalf("expected 4 entities, got %d", len(px.Entities))
	}
	if px.Entities[1].Type != "greynoise.noise" || px.Entities[1].Value != "Common Business Detected" {
		t.Errorf("riot noise entity = %#v", px.Entities[1])
	}
}

func TestCommunityIPLookup_IPNotSeen_ReturnsInform(t *testing.T) {
	// GreyNoise community returns message and noise=false, riot=false for unknown IPs.
	mock := communityMock(&greynoise.CommunityResponse{
		IP:      "192.168.0.1",
		Message: "IP not observed scanning the internet",
	}, nil)

	px := runTransform(t, &CommunityIPLookup{}, mock, makeReq("192.168.0.1"))
	if len(px.Entities) != 2 {
		t.Fatalf("expected input IP and no-noise entities, got %d", len(px.Entities))
	}
	if px.Entities[1].Type != "greynoise.noise" || px.Entities[1].Value != "No Noise Detected" {
		t.Errorf("no-noise entity = %#v", px.Entities[1])
	}
	if len(px.Messages) != 1 || px.Messages[0].Type != maltego.MsgTypeInform {
		t.Errorf("expected Inform UIMessage, got %#v", px.Messages)
	}
}

func TestCommunityIPLookup_APIError_ReturnsFatalError(t *testing.T) {
	mock := communityMock(nil, errors.New("network timeout"))
	px := runTransform(t, &CommunityIPLookup{}, mock, makeReq("1.2.3.4"))
	if len(px.Entities) != 1 {
		t.Fatalf("expected copied input entity, got %d", len(px.Entities))
	}
	if len(px.Messages) != 1 || px.Messages[0].Type != maltego.MsgTypeInform {
		t.Errorf("expected Inform UIMessage, got %#v", px.Messages)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupAllDetails
// ──────────────────────────────────────────────────────────────────────────────

func contextMock(r *greynoise.ContextResponse, err error) *greynoise.MockClient {
	return &greynoise.MockClient{
		ContextIPFn: func(_ context.Context, _ string) (*greynoise.ContextResponse, error) {
			return r, err
		},
	}
}

func TestNoiseIPLookupAllDetails_Name(t *testing.T) {
	if got := (&NoiseIPLookupAllDetails{}).Name(); got != "GreyNoiseNoiseIPLookupAllDetails" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupAllDetails_Success_FullContext(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:             "5.5.5.5",
		Seen:           true,
		Classification: "malicious",
		Actor:          "Mirai",
		Tags:           []string{"mirai", "scanner"},
		CVEs:           []string{"CVE-2021-44228"},
		Ports:          []int{22, 80},
		FirstSeen:      "2023-01-01",
		LastSeen:       "2024-01-01",
		Metadata: greynoise.Metadata{
			ASN:          "AS12345",
			Organization: "Evil Corp",
			Country:      "RU",
			Tor:          false,
		},
	}, nil)

	px := runTransform(t, &NoiseIPLookupAllDetails{}, mock, makeReq("5.5.5.5"))

	if len(px.Entities) != 9 {
		t.Fatalf("expected 9 entities, got %d", len(px.Entities))
	}
	e := px.Entities[0]
	if e.Properties["gn_last_seen"] != "2024-01-01" {
		t.Errorf("gn_last_seen = %q", e.Properties["gn_last_seen"])
	}
	if px.Entities[1].Type != "greynoise.noise" || px.Entities[1].Value != "Noise Detected" {
		t.Errorf("noise entity = %#v", px.Entities[1])
	}
	if px.Entities[2].Type != maltego.EntityPerson || px.Entities[2].Value != "Mirai" {
		t.Errorf("actor entity = %#v", px.Entities[2])
	}
	if px.Entities[4].Type != maltego.EntityAS || px.Entities[4].Value != "12345" {
		t.Errorf("asn entity = %#v", px.Entities[4])
	}
}

func TestNoiseIPLookupAllDetails_NotSeen_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "10.0.0.1", Seen: false}, nil)
	px := runTransform(t, &NoiseIPLookupAllDetails{}, mock, makeReq("10.0.0.1"))
	if len(px.Entities) != 2 {
		t.Fatalf("expected input IP and no-noise entities, got %d", len(px.Entities))
	}
	if px.Entities[1].Type != "greynoise.noise" || px.Entities[1].Value != "No Noise Detected" {
		t.Errorf("no-noise entity = %#v", px.Entities[1])
	}
	if len(px.Messages) != 1 || px.Messages[0].Type != maltego.MsgTypeInform {
		t.Errorf("expected Inform UIMessage, got %#v", px.Messages)
	}
}

func TestNoiseIPLookupAllDetails_APIError_ReturnsFatalError(t *testing.T) {
	mock := contextMock(nil, errors.New("api error"))
	px := runTransform(t, &NoiseIPLookupAllDetails{}, mock, makeReq("5.5.5.5"))
	if len(px.Entities) != 1 {
		t.Fatalf("expected copied input entity, got %d", len(px.Entities))
	}
	if len(px.Messages) != 1 || px.Messages[0].Type != maltego.MsgTypeInform {
		t.Errorf("expected Inform UIMessage, got %#v", px.Messages)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupGetActor
// ──────────────────────────────────────────────────────────────────────────────

func TestNoiseIPLookupGetActor_Name(t *testing.T) {
	if got := (&NoiseIPLookupGetActor{}).Name(); got != "GreyNoiseNoiseIPLookupGetActor" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupGetActor_HasActor(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "5.5.5.5", Seen: true, Actor: "APT28"}, nil)
	px := runTransform(t, &NoiseIPLookupGetActor{}, mock, makeReq("5.5.5.5"))

	if len(px.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(px.Entities))
	}
	if px.Entities[0].Type != maltego.EntityPerson {
		t.Errorf("entity type = %q, want maltego.Person", px.Entities[0].Type)
	}
	if px.Entities[0].Value != "APT28" {
		t.Errorf("entity value = %q, want APT28", px.Entities[0].Value)
	}
}

func TestNoiseIPLookupGetActor_NoActor_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "5.5.5.5", Seen: true, Actor: ""}, nil)
	px := runTransform(t, &NoiseIPLookupGetActor{}, mock, makeReq("5.5.5.5"))
	assertInformNoEntities(t, px)
}

func TestNoiseIPLookupGetActor_NotSeen_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "5.5.5.5", Seen: false}, nil)
	px := runTransform(t, &NoiseIPLookupGetActor{}, mock, makeReq("5.5.5.5"))
	assertInformNoEntities(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupGetCVEs
// ──────────────────────────────────────────────────────────────────────────────

func TestNoiseIPLookupGetCVEs_Name(t *testing.T) {
	if got := (&NoiseIPLookupGetCVEs{}).Name(); got != "GreyNoiseNoiseIPLookupGetCVEs" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupGetCVEs_MultipleCVEs(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:   "1.2.3.4",
		Seen: true,
		CVEs: []string{"CVE-2021-44228", "CVE-2022-0001", "CVE-2023-9999"},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetCVEs{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 3 {
		t.Fatalf("expected 3 CVE entities, got %d", len(px.Entities))
	}
	for _, e := range px.Entities {
		if e.Type != maltego.EntityCVE {
			t.Errorf("entity type = %q, want maltego.CVE", e.Type)
		}
	}
}

func TestNoiseIPLookupGetCVEs_NoCVEs_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "1.2.3.4", Seen: true, CVEs: nil}, nil)
	px := runTransform(t, &NoiseIPLookupGetCVEs{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupGetOrg
// ──────────────────────────────────────────────────────────────────────────────

func TestNoiseIPLookupGetOrg_Name(t *testing.T) {
	if got := (&NoiseIPLookupGetOrg{}).Name(); got != "GreyNoiseNoiseIPLookupGetOrg" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupGetOrg_HasOrganization(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:   "1.2.3.4",
		Seen: true,
		Metadata: greynoise.Metadata{
			ASN:          "AS9999",
			Organization: "Acme Corp",
			Country:      "US",
			Category:     "hosting",
		},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetOrg{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(px.Entities))
	}
	if px.Entities[0].Type != maltego.EntityOrganization {
		t.Errorf("entity type = %q, want maltego.Organization", px.Entities[0].Type)
	}
	if px.Entities[0].Value != "Acme Corp" {
		t.Errorf("entity value = %q, want Acme Corp", px.Entities[0].Value)
	}
	if px.Entities[0].Properties["asn"] != "AS9999" {
		t.Errorf("asn = %q", px.Entities[0].Properties["asn"])
	}
}

func TestNoiseIPLookupGetOrg_OrgEmpty_FallsBackToASN(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:       "1.2.3.4",
		Seen:     true,
		Metadata: greynoise.Metadata{ASN: "AS777", Organization: ""},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetOrg{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 1 {
		t.Fatalf("expected 1 entity for ASN fallback, got %d", len(px.Entities))
	}
	if px.Entities[0].Value != "AS777" {
		t.Errorf("fallback value = %q, want AS777", px.Entities[0].Value)
	}
}

func TestNoiseIPLookupGetOrg_NoOrgNoASN_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "1.2.3.4", Seen: true}, nil)
	px := runTransform(t, &NoiseIPLookupGetOrg{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupGetPorts
// ──────────────────────────────────────────────────────────────────────────────

func TestNoiseIPLookupGetPorts_Name(t *testing.T) {
	if got := (&NoiseIPLookupGetPorts{}).Name(); got != "GreyNoiseNoiseIPLookupGetPorts" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupGetPorts_FromRawScan(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:   "1.2.3.4",
		Seen: true,
		RawData: greynoise.RawData{
			Scan: []greynoise.ScanEntry{
				{Port: 22, Protocol: "TCP"},
				{Port: 443, Protocol: "TCP"},
			},
		},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetPorts{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 2 {
		t.Fatalf("expected 2 port entities, got %d", len(px.Entities))
	}
	for _, e := range px.Entities {
		if e.Type != maltego.EntityPort {
			t.Errorf("entity type = %q, want maltego.Port", e.Type)
		}
	}
	vals := entityValues(px)
	if !contains(vals, "22") || !contains(vals, "443") {
		t.Errorf("ports %v missing expected 22 or 443", vals)
	}
}

func TestNoiseIPLookupGetPorts_FromPortsList_WhenNoRawScan(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:    "1.2.3.4",
		Seen:  true,
		Ports: []int{80, 8080},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetPorts{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 2 {
		t.Fatalf("expected 2 port entities, got %d", len(px.Entities))
	}
}

func TestNoiseIPLookupGetPorts_NoPorts_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "1.2.3.4", Seen: true}, nil)
	px := runTransform(t, &NoiseIPLookupGetPorts{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPLookupGetTags
// ──────────────────────────────────────────────────────────────────────────────

func TestNoiseIPLookupGetTags_Name(t *testing.T) {
	if got := (&NoiseIPLookupGetTags{}).Name(); got != "GreyNoiseNoiseIPLookupGetTags" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPLookupGetTags_MultipleTags(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{
		IP:   "1.2.3.4",
		Seen: true,
		Tags: []string{"scanner", "mirai", "smb-scanner"},
	}, nil)

	px := runTransform(t, &NoiseIPLookupGetTags{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 3 {
		t.Fatalf("expected 3 tag entities, got %d", len(px.Entities))
	}
	for _, e := range px.Entities {
		if e.Type != maltego.EntityHashtag {
			t.Errorf("entity type = %q, want maltego.Hashtag", e.Type)
		}
	}
	vals := entityValues(px)
	if !contains(vals, "scanner") {
		t.Error("tag 'scanner' not found in entities")
	}
}

func TestNoiseIPLookupGetTags_NoTags_ReturnsInform(t *testing.T) {
	mock := contextMock(&greynoise.ContextResponse{IP: "1.2.3.4", Seen: true, Tags: nil}, nil)
	px := runTransform(t, &NoiseIPLookupGetTags{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// NoiseIPSims
// ──────────────────────────────────────────────────────────────────────────────

func simMock(r *greynoise.SimilarityResponse, err error) *greynoise.MockClient {
	return &greynoise.MockClient{
		SimilarIPsFn: func(_ context.Context, _ string) (*greynoise.SimilarityResponse, error) {
			return r, err
		},
	}
}

func TestNoiseIPSims_Name(t *testing.T) {
	if got := (&NoiseIPSims{}).Name(); got != "GreyNoiseNoiseIPSims" {
		t.Errorf("Name() = %q", got)
	}
}

func TestNoiseIPSims_ReturnsSimilarIPs(t *testing.T) {
	mock := simMock(&greynoise.SimilarityResponse{
		IP: "1.2.3.4",
		Similar: []greynoise.SimilarIP{
			{IP: "5.5.5.5", Score: 0.95, Actor: "Mirai"},
			{IP: "6.6.6.6", Score: 0.82, Actor: ""},
		},
	}, nil)

	px := runTransform(t, &NoiseIPSims{}, mock, makeReq("1.2.3.4"))

	if len(px.Entities) != 2 {
		t.Fatalf("expected 2 similar IPs, got %d", len(px.Entities))
	}
	if px.Entities[0].Properties["similarity_score"] == "" {
		t.Error("similarity_score property missing")
	}
}

func TestNoiseIPSims_RespectsHardLimit(t *testing.T) {
	similar := make([]greynoise.SimilarIP, 20)
	for i := range similar {
		similar[i] = greynoise.SimilarIP{IP: "1.1.1.1", Score: 0.5}
	}

	mock := simMock(&greynoise.SimilarityResponse{IP: "1.2.3.4", Similar: similar}, nil)

	req := makeReq("1.2.3.4")
	req.HardLimit = 5

	px := runTransform(t, &NoiseIPSims{}, mock, req)
	if len(px.Entities) != 5 {
		t.Errorf("expected 5 entities (HardLimit), got %d", len(px.Entities))
	}
}

func TestNoiseIPSims_NoSimilar_ReturnsInform(t *testing.T) {
	mock := simMock(&greynoise.SimilarityResponse{IP: "1.2.3.4", Similar: nil}, nil)
	px := runTransform(t, &NoiseIPSims{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

func TestNoiseIPSims_APIError_ReturnsFatalError(t *testing.T) {
	mock := simMock(nil, errors.New("sim api error"))
	px := runTransform(t, &NoiseIPSims{}, mock, makeReq("1.2.3.4"))
	assertFatalError(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// RIOTIPLookup
// ──────────────────────────────────────────────────────────────────────────────

func riotMock(r *greynoise.RIOTResponse, err error) *greynoise.MockClient {
	return &greynoise.MockClient{
		RIOTFn: func(_ context.Context, _ string) (*greynoise.RIOTResponse, error) {
			return r, err
		},
	}
}

func TestRIOTIPLookup_Name(t *testing.T) {
	if got := (&RIOTIPLookup{}).Name(); got != "GreyNoiseRIOTIPLookup" {
		t.Errorf("Name() = %q", got)
	}
}

func TestRIOTIPLookup_KnownService(t *testing.T) {
	mock := riotMock(&greynoise.RIOTResponse{
		IP:          "8.8.8.8",
		Riot:        true,
		Name:        "Google Public DNS",
		Category:    "public_dns",
		TrustLevel:  "1",
		Description: "Google DNS resolver",
		LastUpdated: "2024-01-01",
	}, nil)

	px := runTransform(t, &RIOTIPLookup{}, mock, makeReq("8.8.8.8"))

	if len(px.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(px.Entities))
	}
	e := px.Entities[0]
	if e.Properties["name"] != "Google Public DNS" {
		t.Errorf("name = %q, want Google Public DNS", e.Properties["name"])
	}
	if e.Properties["riot"] != "true" {
		t.Errorf("riot = %q, want true", e.Properties["riot"])
	}
	if e.Properties["trust_level"] != "1" {
		t.Errorf("trust_level = %q, want 1", e.Properties["trust_level"])
	}
}

func TestRIOTIPLookup_UnknownIP_ReturnsInform(t *testing.T) {
	mock := riotMock(&greynoise.RIOTResponse{IP: "1.2.3.4", Riot: false}, nil)
	px := runTransform(t, &RIOTIPLookup{}, mock, makeReq("1.2.3.4"))
	assertInformNoEntities(t, px)
}

func TestRIOTIPLookup_APIError_ReturnsFatalError(t *testing.T) {
	mock := riotMock(nil, errors.New("riot api down"))
	px := runTransform(t, &RIOTIPLookup{}, mock, makeReq("1.2.3.4"))
	assertFatalError(t, px)
}

// ──────────────────────────────────────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────────────────────────────────────

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
