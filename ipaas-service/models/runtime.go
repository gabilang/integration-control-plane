package models

type Runtime struct {
	RuntimeID       string            `json:"runtimeId"`
	RuntimeType     string            `json:"runtimeType,omitempty"`
	Status          string            `json:"status,omitempty"`
	Version         string            `json:"version,omitempty"`
	PlatformName    string            `json:"platformName,omitempty"`
	PlatformVersion string            `json:"platformVersion,omitempty"`
	PlatformHome    string            `json:"platformHome,omitempty"`
	OsName          string            `json:"osName,omitempty"`
	OsVersion       string            `json:"osVersion,omitempty"`
	RegisteredAt    string            `json:"registrationTime,omitempty"`
	LastHeartbeat   string            `json:"lastHeartbeat,omitempty"`
	Component       *RuntimeComponent `json:"component,omitempty"`
}

type RuntimeComponent struct {
	DisplayName string `json:"displayName,omitempty"`
}

type RuntimeList struct {
	Items []Runtime `json:"items"`
}
