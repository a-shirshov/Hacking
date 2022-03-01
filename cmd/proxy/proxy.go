package main

import (
	"log"
	"net"

	proxyHndl "proxy/internal/proxy/handler"
	proxyRepo "proxy/internal/proxy/repository"
	proxyUsec "proxy/internal/proxy/usecase"
	"proxy/utils"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
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

	proxyR := proxyRepo.NewRepository(db)
	proxyU := proxyUsec.NewUsecase(proxyR)
	proxyH := proxyHndl.NewHandler(proxyU)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}

		log.Println("From:", conn.RemoteAddr().String())
		log.Println("To proxy:", conn.LocalAddr().String())
		go proxyH.Handler(conn)
	}

}
