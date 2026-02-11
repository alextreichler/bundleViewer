package models

type TopicConfigsResponse []TopicConfig
type BrokerConfigsResponse []TopicConfig

type TopicConfig struct {
	Name       string        `json:"Name"`
	Err        interface{}   `json:"Err"`
	ErrMessage string        `json:"ErrMessage"`
	Configs    []ConfigEntry `json:"Configs"`
}

type ConfigEntry struct {
	Key       string          `json:"Key"`
	Value     string          `json:"Value"`
	Source    string          `json:"Source"`
	Sensitive bool            `json:"Sensitive"`
	Synonyms  []ConfigSynonym `json:"Synonyms"`
}

type ConfigSynonym struct {
	Key    string `json:"Key"`
	Value  string `json:"Value"`
	Source string `json:"Source"`
}
