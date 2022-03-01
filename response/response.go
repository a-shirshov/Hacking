package response

import (
	"net/http"
	"log"
	"encoding/json"
)

type Response struct {
	Message string `json:"message"`
}


func SendResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(response)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Print(err)
	}
}