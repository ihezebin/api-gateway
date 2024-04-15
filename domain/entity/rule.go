package entity

type Rule struct {
	Domains []string          `json:"domains" mapstructure:"domains"`
	Headers map[string]string `json:"headers" mapstructure:"headers"`
	// Timeout seconds
	Timeout int   `json:"timeout" mapstructure:"timeout"`
	Uris    []Uri `json:"uris" mapstructure:"uris"`
}
