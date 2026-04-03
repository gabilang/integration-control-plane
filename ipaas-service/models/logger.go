package models

type Logger struct {
	ComponentName string   `json:"componentName,omitempty"`
	LogLevel      string   `json:"logLevel,omitempty"`
	RuntimeIDs    []string `json:"runtimeIds,omitempty"`
}

type LoggerList struct {
	Items []Logger `json:"items"`
}

type UpdateLogLevelInput struct {
	RuntimeIDs    []string `json:"runtimeIds"`
	ComponentName string   `json:"componentName"`
	LogLevel      string   `json:"logLevel"`
}

type UpdateLogLevelResult struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message,omitempty"`
	CommandIDs []string `json:"commandIds,omitempty"`
}
