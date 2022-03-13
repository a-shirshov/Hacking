package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"proxy/utils"

	webHndl "proxy/internal/web/handler"
	webRepo "proxy/internal/web/repository"
	webUsec "proxy/internal/web/usecase"

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

	params := utils.GetParamsFromFile("params.txt")

	webR := webRepo.NewRepository(db)
	webU := webUsec.NewUsecase(webR)
	webH := webHndl.NewHandler(webU, &params)

	r := mux.NewRouter()
	r.HandleFunc("/requests", webH.GetRequests)
	r.HandleFunc("/request/{id}", webH.GetRequest)
	r.HandleFunc("/repeat/{id}", webH.RepeatRequest)
	r.HandleFunc("/scan/{id}", webH.ScanRequest)

	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal(err)
	}
}
