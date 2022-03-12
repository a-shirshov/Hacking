package handler

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxy/internal/web"
	"proxy/response"
	"strconv"
)

type Handler struct {
	usecase web.Usecase
	params  *[]string
}

func NewHandler(usecase web.Usecase, params *[]string) *Handler {
	return &Handler{
		usecase: usecase,
		params:  params,
	}
}

func (h *Handler) GetRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.usecase.GetRequestsJson()
	if err != nil {
		log.Print(err)
		return
	}

	response.SendResponse(w, requests)

}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		return
	}
	request, err := h.usecase.GetRequestJson(id)
	if err != nil {
		log.Print(err)
		return
	}
	response.SendResponse(w, request)
}

func (h *Handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		return
	}

	err = h.usecase.RepeatRequest(id)
	if err != nil {
		log.Print(err.Error())
		return
	}
	response.SendResponse(w, response.Response{Message: "OK"})

}

func (h *Handler) ScanRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		return
	}

	err = h.usecase.ScanRequest(id, h.params)
	if err != nil {
		log.Print(err)
		return
	}
}
