package handler

import (
	"log"
	"net/http"
	"proxy/internal/web"
	"proxy/response"
	"github.com/gorilla/mux"
	"strconv"
)

type Handler struct {
	usecase web.Usecase
}

func NewHandler(usecase web.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) GetRequests(w http.ResponseWriter,r* http.Request) {
	requests, err := h.usecase.GetRequestsJson()
	if err != nil {
		log.Print(err)
		return
	}

	response.SendResponse(w,requests)

}

func (h *Handler) GetRequest(w http.ResponseWriter, r* http.Request) {
	idStr := mux.Vars(r)["id"]
	id,err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		return
	}
	request, err := h.usecase.GetRequestJson(id)
	if err != nil {
		log.Print(err)
		return 
	}
	response.SendResponse(w,request)
}

func (h *Handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id,err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		return
	}
	
	err = h.usecase.RepeatRequest(id)
	if err != nil {
		log.Print(err.Error())
		return
	}
	response.SendResponse(w,response.Response{Message: "OK"})
	
}