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
