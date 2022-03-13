package web

import (
	"proxy/internal/web/models"
)

type Usecase interface {
	GetRequestsJson() (*models.RequestsJson, error)
	GetRequestJson(id int) (*models.RequestJson, error)
	RepeatRequest(id int) error
	ScanRequest(id int, params *[]string) (*[]string, error)
}
