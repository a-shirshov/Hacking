package web

import (
	"proxy/internal/web/models"
)

type Repository interface {
	GetRequestsJson() (*models.RequestsJson, error)
	GetRequestJson(id int) (*models.RequestJsonWithSecure, error)
	GetRequest(id int) (*models.Request, error)
}
