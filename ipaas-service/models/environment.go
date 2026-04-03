package models

type Environment struct {
	UID          string `json:"uid,omitempty"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName,omitempty"`
	DataPlaneRef string `json:"dataPlaneRef,omitempty"`
	IsProduction bool   `json:"isProduction,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
}

type EnvironmentList struct {
	Items []Environment `json:"items"`
}

// CreateEnvironmentRequest holds the fields required to create an environment.
type CreateEnvironmentRequest struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName,omitempty"`
	IsProduction bool   `json:"isProduction,omitempty"`
}

// UpdateEnvironmentRequest holds the mutable fields of an environment.
type UpdateEnvironmentRequest struct {
	DisplayName  string `json:"displayName,omitempty"`
	IsProduction *bool  `json:"isProduction,omitempty"`
}
