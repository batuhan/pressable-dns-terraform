package pressable

type Site struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	DisplayName       string `json:"displayName"`
	URL               string `json:"url"`
	State             string `json:"state"`
	PHPVersion        string `json:"phpVersion"`
	WPEnvironmentType string `json:"wpEnvironmentType"`
	MultisiteSupport  bool   `json:"multisiteSupport"`
	MCPEnabled        bool   `json:"mcpEnabled"`
	Created           string `json:"created"`
	Updated           string `json:"updated"`
	Raw               map[string]any
}

type Zone struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Created string      `json:"created"`
	Updated string      `json:"updated"`
	Records []DNSRecord `json:"records"`
}

type DNSRecord struct {
	ID       int    `json:"id"`
	ZoneID   int    `json:"zoneId"`
	Primary  bool   `json:"primary"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int64 `json:"priority"`
	Weight   *int64 `json:"weight"`
	Port     *int64 `json:"port"`
	TTL      int64  `json:"ttl"`
	SiteID   *int64 `json:"siteId"`
	Created  string `json:"created"`
	Raw      map[string]any
}

type SiteDomain struct {
	ID           int    `json:"id"`
	DomainName   string `json:"domainName"`
	Primary      bool   `json:"primary"`
	Healthy      bool   `json:"healthy"`
	DNSConfirmed bool   `json:"dnsConfirmed"`
	Provisioned  bool   `json:"provisioned"`
	ZoneID       int    `json:"zoneId"`
	ZoneName     string `json:"zoneName"`
	CreatedAt    string `json:"createdAt"`
	Raw          map[string]any
}
