package greynoise

// MockServerData returns realistic fake responses for demo/testing purposes.
// Used by greynoise-api service when MOCK_MODE=true.

func MockCommunity(ip string) *CommunityResponse {
	return &CommunityResponse{
		IP:             ip,
		Noise:          true,
		Riot:           false,
		Classification: "malicious",
		Name:           "Unknown",
		Link:           "https://viz.greynoise.io/ip/" + ip,
		LastSeen:       "2026-05-26",
		Message:        "This IP is commonly included in our NOISE dataset",
	}
}

func MockContext(ip string) *ContextResponse {
	return &ContextResponse{
		IP:             ip,
		Seen:           true,
		Classification: "malicious",
		Actor:          "Mirai Botnet",
		FirstSeen:      "2026-01-14",
		LastSeen:       "2026-05-26",
		Spoofable:      false,
		Bot:            true,
		VPN:            false,
		Tags:           []string{"Mirai", "IoT Scanner", "Telnet Scanner"},
		CVEs:           []string{"CVE-2023-46604", "CVE-2021-44228", "CVE-2022-42889"},
		Metadata: Metadata{
			ASN:          "AS4134",
			Country:      "China",
			CountryCode:  "CN",
			City:         "Beijing",
			Organization: "Chinanet",
			Category:     "isp",
			OS:           "Linux 2.2-3.x",
			Region:       "Beijing",
			RDNS:         "",
			Tor:          false,
		},
	}
}

func MockRIOT(ip string) *RIOTResponse {
	return &RIOTResponse{
		IP:          ip,
		Riot:        false,
		Category:    "",
		Name:        "",
		Description: "This IP is not a common business service",
		Explanation: "This IP has not been identified as belonging to a known benign service",
		LastUpdated: "2026-05-26",
		TrustLevel:  "",
	}
}

func MockSimilar(ip string) *SimilarityResponse {
	return &SimilarityResponse{
		IP:    ip,
		Total: 4,
		Similar: []SimilarIP{
			{IP: "45.142.212.100", Score: 0.97, Actor: "Mirai Botnet", Tags: []string{"Mirai", "IoT Scanner"}, Similarity: []string{"ports", "tags", "asn"}},
			{IP: "185.220.101.45", Score: 0.94, Actor: "Mirai Botnet", Tags: []string{"Telnet Scanner"}, Similarity: []string{"ports", "tags"}},
			{IP: "91.92.248.150", Score: 0.91, Actor: "Unknown", Tags: []string{"IoT Scanner", "Brute Force"}, Similarity: []string{"tags", "os"}},
			{IP: "194.165.16.78", Score: 0.88, Actor: "Unknown", Tags: []string{"Mirai"}, Similarity: []string{"ports"}},
		},
	}
}

func MockGNQL(query string) *GNQLResponse {
	return &GNQLResponse{
		Complete: true,
		Count:    5,
		Message:  "Query results",
		Query:    query,
		Data: []GNQLEntry{
			{IP: "45.142.212.100", Classification: "malicious", Actor: "Mirai Botnet", Tags: []string{"Mirai", "IoT Scanner"}, FirstSeen: "2026-01-10", LastSeen: "2026-05-26", ASN: "AS9009", Organization: "M247 Ltd", Country: "Romania"},
			{IP: "185.220.101.45", Classification: "malicious", Actor: "Mirai Botnet", Tags: []string{"Telnet Scanner"}, FirstSeen: "2026-02-01", LastSeen: "2026-05-25", ASN: "AS60068", Organization: "Datacamp Limited", Country: "United Kingdom"},
			{IP: "91.92.248.150", Classification: "malicious", Actor: "Unknown", Tags: []string{"IoT Scanner"}, FirstSeen: "2026-03-15", LastSeen: "2026-05-24", ASN: "AS211680", Organization: "ALEXHOST SRL", Country: "Moldova"},
			{IP: "194.165.16.78", Classification: "malicious", Actor: "Unknown", Tags: []string{"Brute Force", "SSH Scanner"}, FirstSeen: "2026-01-20", LastSeen: "2026-05-26", ASN: "AS35624", Organization: "Petersburg Internet Network", Country: "Russia"},
			{IP: "103.168.230.45", Classification: "malicious", Actor: "Hafnium", Tags: []string{"ProxyShell", "Exchange Scanner"}, FirstSeen: "2026-04-01", LastSeen: "2026-05-22", ASN: "AS55720", Organization: "Gigabit Hosting", Country: "Malaysia"},
		},
	}
}
