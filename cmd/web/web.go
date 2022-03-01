package main

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"proxy/utils"

	webHndl "proxy/internal/web/handler"
	webUsec "proxy/internal/web/usecase"
	webRepo "proxy/internal/web/repository"

	_ "github.com/lib/pq"
)


func main() {

	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := utils.InitPostgresDB()
	if err != nil {
		log.Print(err)
		return
	}
	defer db.Close()

	webR := webRepo.NewRepository(db)
	webU := webUsec.NewUsecase(webR)
	webH := webHndl.NewHandler(webU)

	r := mux.NewRouter()
	r.HandleFunc("/requests",webH.GetRequests)
	r.HandleFunc("/request/{id}",webH.GetRequest)
	r.HandleFunc("/repeat/{id}",webH.RepeatRequest)

	err = http.ListenAndServe(":8000",r)
	if err != nil {
		log.Fatal(err)
	}
}