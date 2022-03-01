package repository

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

const insertRequestQuery = `insert into "request-response" (request, response, requestJson, responseJson, isSecure) 
	values ($1,$2,$3,$4,$5)`

func (r *Repository) Save(request string, response string,reqJson string, resJson string, isSecure bool) {
	_, err := r.db.Exec(insertRequestQuery, request, response, reqJson, resJson, isSecure)
	if err != nil {
		log.Fatal(err.Error())
	}

}
