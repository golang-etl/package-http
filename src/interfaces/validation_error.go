package interfaces

type ValidationError struct {
	Property string `json:"property"`
	Path     string `json:"path"`
	Rule     string `json:"rule"`
	Message  string `json:"message,omitempty"`
}
