package api

import "encoding/json"

type ElasticProxyConfig struct {
	LogPath            string                               `json:"logPath"`
	LogFlags           *ElasticProxyLogFlagsConfig          `json:"logFlags"`
	Elastic            *ElasticProxyElasticConfig           `json:"elastic,omitempty"`
	Sneller            *ElasticProxySnellerConfig           `json:"sneller,omitempty"`
	Mapping            map[string]ElasticProxyMappingConfig `json:"mapping"`
	CompareWithElastic bool                                 `json:"compareWithElastic,omitempty"`
}

type ElasticProxyLogFlagsConfig struct {
	LogRequest         bool `json:"logRequest"`
	LogQueryParameters bool `json:"logQueryParameters"`
	LogSQL             bool `json:"logSQL"`
	LogSnellerResult   bool `json:"logSnellerResult"`
	LogPreprocessed    bool `json:"logPreprocessed"`
	LogResult          bool `json:"logResult"`
}

type ElasticProxyElasticConfig struct {
	EndPoint   string `json:"endpoint,omitempty"`
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	ESPassword string `json:"esPassword,omitempty"`
	IgnoreCert bool   `json:"ignoreCert,omitempty"`
}

type ElasticProxySnellerConfig struct {
	EndPoint string `json:"endpoint,omitempty"`
	Token    string `json:"token,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`
}

type ElasticProxyMappingConfig struct {
	Database               string                             `json:"database"`
	Table                  string                             `json:"table"`
	IgnoreTotalHits        bool                               `json:"ignoreTotalHits"`
	IgnoreSumOtherDocCount bool                               `json:"ignoreSumOtherDocCount"`
	TypeMapping            map[string]ElasticProxyTypeMapping `json:"typeMapping,omitempty"`
}

type ElasticProxyTypeMapping struct {
	Type   string            `json:"type"`
	Fields map[string]string `json:"fields,omitempty"`
}

func (tm *ElasticProxyTypeMapping) UnmarshalJSON(data []byte) error {
	type _elasticProxyTypeMapping ElasticProxyTypeMapping
	if err := json.Unmarshal(data, (*_elasticProxyTypeMapping)(tm)); err != nil {
		var typeName string
		if err := json.Unmarshal(data, &typeName); err != nil {
			return err
		}
		tm.Type = typeName
		tm.Fields = make(map[string]string, 0)
	}
	return nil
}

func (tm *ElasticProxyTypeMapping) MarshalJSON() ([]byte, error) {
	if len(tm.Fields) == 0 {
		return json.Marshal(tm.Type)
	}

	type _elasticProxyTypeMapping ElasticProxyTypeMapping
	return json.Marshal((*_elasticProxyTypeMapping)(tm))
}
