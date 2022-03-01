package usecase

import (
	"proxy/internal/proxy"
	"proxy/internal/proxy/parser"
)

type Usecase struct {
	repo proxy.Repository
}

func NewUsecase(repo proxy.Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (u *Usecase) Save(request string, response string, isSecure bool) {
	jsonRequest := parser.RequestToJson(request)
	jsonResponse := parser.ResponseToJson(response)
	u.repo.Save(request,response,jsonRequest, jsonResponse,isSecure)
}
