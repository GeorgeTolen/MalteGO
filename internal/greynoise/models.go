package greynoise

// CommunityResponse — /v3/community/{ip}
type CommunityResponse struct {
	IP             string `json:"ip"`
	Noise          bool   `json:"noise"`
	Riot           bool   `json:"riot"`
	Classification string `json:"classification"` // malicious | benign | unknown
	Name           string `json:"name"`
	Link           string `json:"link"`
	LastSeen       string `json:"last_seen"`
	Message        string `json:"message"`
}

// ContextResponse — /v2/noise/context/{ip}
type ContextResponse struct {
	IP             string        `json:"ip"`
	Seen           bool          `json:"seen"`
	Classification string        `json:"classification"`
	Spoofable      bool          `json:"spoofable"`
	FirstSeen      string        `json:"first_seen"`
	LastSeen       string        `json:"last_seen"`
	Actor          string        `json:"actor"`
	Tags           []string      `json:"tags"`
	CVEs           []string      `json:"cve"`
	Ports          []int         `json:"ports"`
	Metadata       Metadata      `json:"metadata"`
	RawData        RawData       `json:"raw_data"`
}

type Metadata struct {
	ASN            string `json:"asn"`
	City           string `json:"city"`
	Country        string `json:"country"`
	CountryCode    string `json:"country_code"`
	Organization   string `json:"organization"`
	Category       string `json:"category"`
	Tor            bool   `json:"tor"`
	RDNS           string `json:"rdns"`
	OS             string `json:"os"`
	Region         string `json:"region"`
}

type RawData struct {
	Scan []ScanEntry `json:"scan"`
}

type ScanEntry struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// RIOTResponse — /v2/riot/{ip}
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

// SimilarityResponse — /v1/experimental/gnoise/similar/{ip}
type SimilarityResponse struct {
	IP      string        `json:"ip"`
	Similar []SimilarIP   `json:"similar_ips"`
}

type SimilarIP struct {
	IP         string   `json:"ip"`
	Score      float64  `json:"score"`
	Actor      string   `json:"actor"`
	Similarity []string `json:"features"`
}

// GNQLResponse — /v2/experimental/gnql
type GNQLResponse struct {
	Complete  bool       `json:"complete"`
	Count     int        `json:"count"`
	Message   string     `json:"message"`
	Query     string     `json:"query"`
	Scroll    string     `json:"scroll"`
	Data      []GNQLEntry `json:"data"`
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
