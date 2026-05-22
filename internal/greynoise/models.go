package greynoise

import "encoding/json"

// ── Community API (/v3/community/{ip}) ──────────────────────────────────────

type CommunityResponse struct {
	IP             string `json:"ip"`
	Noise          bool   `json:"noise"`
	Riot           bool   `json:"riot"`
	Classification string `json:"classification"`
	Name           string `json:"name"`
	Link           string `json:"link"`
	LastSeen       string `json:"last_seen"`
	Message        string `json:"message"`
}

// ── v3 Full IP response (/v3/ip/{ip}) ───────────────────────────────────────

type V3IPResponse struct {
	IP                          string                `json:"ip"`
	BusinessServiceIntelligence *V3BusinessIntel      `json:"business_service_intelligence"`
	InternetScannerIntelligence *V3ScannerIntel       `json:"internet_scanner_intelligence"`
	RequestMetadata             V3RequestMetadata     `json:"request_metadata"`
}

type V3BusinessIntel struct {
	Found       bool   `json:"found"`
	TrustLevel  string `json:"trust_level"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Explanation string `json:"explanation"`
	LastUpdated string `json:"last_updated"`
	Reference   string `json:"reference"`
}

type V3ScannerIntel struct {
	Found             bool       `json:"found"`
	Classification    string     `json:"classification"`
	Actor             string     `json:"actor"`
	FirstSeen         string     `json:"first_seen"`
	LastSeen          string     `json:"last_seen"`
	LastSeenTimestamp string     `json:"last_seen_timestamp"`
	Spoofable         bool       `json:"spoofable"`
	Bot               bool       `json:"bot"`
	VPN               bool       `json:"vpn"`
	VPNService        string     `json:"vpn_service"`
	Tor               bool       `json:"tor"`
	Tags              []V3Tag    `json:"tags"`
	CVEs              []string   `json:"cves"`
	Metadata          V3Metadata `json:"metadata"`
}

type V3Tag struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Intention   string   `json:"intention"`
	CVEs        []string `json:"cves"`
}

type V3Metadata struct {
	ASN               string `json:"asn"`
	SourceCountry     string `json:"source_country"`
	SourceCountryCode string `json:"source_country_code"`
	SourceCity        string `json:"source_city"`
	Domain            string `json:"domain"`
	Organization      string `json:"organization"`
	Category          string `json:"category"`
	OS                string `json:"os"`
	Region            string `json:"region"`
	RDNS              string `json:"rdns"`
}

type V3RequestMetadata struct {
	RestrictedFields []string `json:"restricted_fields"`
}

// ToContextResponse maps the v3 response to our ContextResponse (used by transforms).
func (v *V3IPResponse) ToContextResponse() *ContextResponse {
	r := &ContextResponse{IP: v.IP}
	if v.InternetScannerIntelligence == nil {
		return r
	}
	isi := v.InternetScannerIntelligence
	r.Seen = isi.Found
	r.Classification = isi.Classification
	r.Actor = isi.Actor
	r.FirstSeen = isi.FirstSeen
	r.LastSeen = isi.LastSeen
	r.Spoofable = isi.Spoofable
	r.Bot = isi.Bot
	r.VPN = isi.VPN
	r.VPNService = isi.VPNService
	r.CVEs = isi.CVEs
	for _, t := range isi.Tags {
		r.Tags = append(r.Tags, t.Name)
	}
	r.Metadata = Metadata{
		ASN:          isi.Metadata.ASN,
		Country:      isi.Metadata.SourceCountry,
		CountryCode:  isi.Metadata.SourceCountryCode,
		City:         isi.Metadata.SourceCity,
		Organization: isi.Metadata.Organization,
		Category:     isi.Metadata.Category,
		OS:           isi.Metadata.OS,
		Region:       isi.Metadata.Region,
		RDNS:         isi.Metadata.RDNS,
		Tor:          isi.Tor,
	}
	return r
}

// ToRIOTResponse maps the v3 response to our RIOTResponse (used by transforms).
func (v *V3IPResponse) ToRIOTResponse() *RIOTResponse {
	r := &RIOTResponse{IP: v.IP}
	if v.BusinessServiceIntelligence == nil {
		return r
	}
	bsi := v.BusinessServiceIntelligence
	r.Riot = bsi.Found
	r.TrustLevel = bsi.TrustLevel
	r.Category = bsi.Category
	r.Name = bsi.Name
	r.Description = bsi.Description
	r.Explanation = bsi.Explanation
	r.LastUpdated = bsi.LastUpdated
	r.Reference = bsi.Reference
	return r
}

// ── Internal models used by transforms (unchanged interface) ─────────────────

type ContextResponse struct {
	IP             string   `json:"ip"`
	Seen           bool     `json:"seen"`
	Classification string   `json:"classification"`
	Spoofable      bool     `json:"spoofable"`
	FirstSeen      string   `json:"first_seen"`
	LastSeen       string   `json:"last_seen"`
	Actor          string   `json:"actor"`
	Tags           []string `json:"tags"`
	CVEs           []string `json:"cve"`
	Ports          []int    `json:"ports"`
	Metadata       Metadata `json:"metadata"`
	RawData        RawData  `json:"raw_data"`
	VPN            bool     `json:"vpn"`
	VPNService     string   `json:"vpn_service"`
	Bot            bool     `json:"bot"`
}

type Metadata struct {
	ASN          string `json:"asn"`
	City         string `json:"city"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Organization string `json:"organization"`
	Category     string `json:"category"`
	Tor          bool   `json:"tor"`
	RDNS         string `json:"rdns"`
	OS           string `json:"os"`
	Region       string `json:"region"`
}

type RawData struct {
	Scan []ScanEntry `json:"scan"`
}

type ScanEntry struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type RIOTResponse struct {
	IP          string `json:"ip"`
	Riot        bool   `json:"riot"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Explanation string `json:"explanation"`
	LastUpdated string `json:"last_updated"`
	Reference   string `json:"reference"`
	TrustLevel  string `json:"trust_level"`
}

// ── Similarity (/v3/similarity/ips/{ip}) ─────────────────────────────────────

type SimilarityResponse struct {
	IP      string      `json:"ip"`
	Similar []SimilarIP `json:"similar_ips"`
	Total   int         `json:"total"`
	Message string      `json:"message"`
}

func (r *SimilarityResponse) UnmarshalJSON(data []byte) error {
	type alias SimilarityResponse
	var raw struct {
		IP json.RawMessage `json:"ip"`
		*alias
	}
	raw.alias = (*alias)(r)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.IP) == 0 {
		return nil
	}
	var ipString string
	if err := json.Unmarshal(raw.IP, &ipString); err == nil {
		r.IP = ipString
		return nil
	}
	var ipObject struct{ IP string `json:"ip"` }
	if err := json.Unmarshal(raw.IP, &ipObject); err != nil {
		return err
	}
	r.IP = ipObject.IP
	return nil
}

type SimilarIP struct {
	IP         string   `json:"ip"`
	Score      float64  `json:"score"`
	Actor      string   `json:"actor"`
	Similarity []string `json:"features"`
	Tags       []string `json:"tags"`
}

// ── GNQL (/v2/experimental/gnql — no v3 equivalent) ─────────────────────────

type GNQLResponse struct {
	Complete bool        `json:"complete"`
	Count    int         `json:"count"`
	Message  string      `json:"message"`
	Query    string      `json:"query"`
	Scroll   string      `json:"scroll"`
	Data     []GNQLEntry `json:"data"`
}

type GNQLEntry struct {
	IP             string   `json:"ip"`
	Classification string   `json:"classification"`
	Actor          string   `json:"actor"`
	Tags           []string `json:"tags"`
	FirstSeen      string   `json:"first_seen"`
	LastSeen       string   `json:"last_seen"`
	ASN            string   `json:"asn"`
	Organization   string   `json:"organization"`
	Country        string   `json:"country"`
}
