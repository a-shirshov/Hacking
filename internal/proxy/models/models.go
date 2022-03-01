package models

type Request struct {
	Method   string                 `json:"method"`
	Path     string                 `json:"path"`
	Protocol string                 `json:"protocol"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Headers  map[string]interface{} `json:"headers,omitempty"`
	Cookies  map[string]interface{} `json:"cookies,omitempty"`
	Body     string                 `json:"body,omitempty"`
}

type Response struct {
	Protocol string                 `json:"protocol"`
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Headers  map[string]interface{} `json:"headers,omitempty"`
	Body     string                 `json:"body,omitempty"`
}
