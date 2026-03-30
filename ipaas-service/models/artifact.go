package models

import "encoding/json"

// ArtifactList is the REST response for artifact queries.
// Items are kept as raw JSON to avoid defining separate structs for every MI
// artifact type (RestApi, Sequence, ProxyService, etc.).
type ArtifactList struct {
	Items []json.RawMessage `json:"items"`
}
