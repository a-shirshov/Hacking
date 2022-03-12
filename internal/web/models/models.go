package models

import ()

type Request struct {
	ID       int
	Request  string
	Response string
	IsSecure bool
}

type RequestJson struct {
	ID       int    `json:"id"`
	Request  string `json:"request,omitempty"`
	Response string `json:"response,omitempty"`
}

type RequestJsonWithSecure struct {
	ID       int    `json:"id"`
	Request  string `json:"request,omitempty"`
	Response string `json:"response,omitempty"`
	IsSecure bool
}

type RequestsJson struct {
	Requests []RequestJson `json:"requests"`
}
