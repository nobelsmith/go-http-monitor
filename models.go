package main

// Config has been created
type Config struct {
	SlackUrl       string `yaml:"slack_url"`
	Insecure       bool   `yaml:"insecure"`
	TimeoutRequest int    `yaml:"timeout_seconds"`
	Verbose        bool   `yaml:"verbose"`
	Checks         []struct {
		URL          string  `yaml:"url"`
		StatusCode   *int    `yaml:"status_code"`
		Match        *string `yaml:"match"`
		ResponseTime *int    `yaml:"response_time"`
		TCP          string  `yaml:"tcp"`
		Port         *int    `yaml:"port"`
	} `yaml:"checks"`
}

// Config has been created
type CheckOutput struct {
	Resource string `json:"resource"`
	Status   string `json:"available"`
	Elapsed  string `json:"elapsed"`
	Error    string `json:"error"`
}

type JsonOutput struct {
	Results []CheckOutput `json:"checks"`
}

type Field struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Section struct {
	Type string `json:"type"`
	Text Field  `json:"text"`
}

type SlackMessage struct {
	Blocks []Section `json:"blocks"`
}
