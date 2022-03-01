package repository

import (
	"log"
	"proxy/internal/web/models"

	"github.com/jmoiron/sqlx"
)

const (
	getRequestsJsonQuery = `select id, requestJson, responseJson from "request-response"`
	getRequestJsonByID = `select id, requestJson, responseJson from "request-response" where id = $1`
	getRequestByID = `select id, request, response, isSecure from "request-response" where id = $1`
) 


type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetRequestsJson() (*models.RequestsJson,error) {
	rows, err := r.db.Queryx(getRequestsJsonQuery)
	if err != nil {
		log.Print(err.Error())
	}
	defer rows.Close()

	requests := &models.RequestsJson{}

	for rows.Next() {
		request := &models.RequestJson{}
	
		err := rows.Scan(&request.ID,&request.Request,&request.Response)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		requests.Requests = append(requests.Requests, *request)
	}
	return requests, nil
}

func (r *Repository) GetRequestJson(id int) (*models.RequestJson,error) {
	row := r.db.QueryRow(getRequestJsonByID,id)
	request := &models.RequestJson{}
	err := row.Scan(&request.ID,&request.Request,&request.Response)
	if err != nil {
		log.Print(err)
		return nil,err
	}
	return request, nil
}

func (r *Repository) GetRequest(id int) (*models.Request, error) {
	row := r.db.QueryRow(getRequestByID,id)
	request := &models.Request{}
	err := row.Scan(&request.ID,&request.Request,&request.Response,&request.IsSecure)
	if err != nil {
		log.Print(err)
		return nil,err
	}
	return request, nil
}