package handler

import (
	"log"
	"math/rand"
	"net"
	"os/exec"
	"proxy/internal/proxy"
	"proxy/internal/proxy/parser"
	"proxy/utils"
	"strconv"
	"strings"

	"crypto/tls"
	"os"

	_ "github.com/lib/pq"
)

type Proxy struct {
	usecase proxy.Usecase
}

func NewHandler(usecase proxy.Usecase) *Proxy {
	return &Proxy{
		usecase: usecase,
	}
}

//HTTPS
//Host is definetly {site.com}:443 and Method is CONNECT

//HTTP
//Host is http://{site.com}:80 or http://{site.com}

const okMessage = "HTTP/1.0 200 Connection established\r\n\r\n"

func (p *Proxy) Handler(conn net.Conn) {
	defer conn.Close()
	message := utils.ReadMessage(conn)
	log.Print("Message:\n", message)
	method, host, modMessage := parser.ParseMessage(message)

	//HTTP
	if method != "CONNECT" {
		dest, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer dest.Close()

		_, err = dest.Write([]byte(modMessage))
		if err != nil {
			log.Print(err.Error())
		}

		response, err := utils.CopyMessage(conn, dest)
		if err != nil {
			log.Print(err.Error())
		}

		p.usecase.Save(modMessage, response, false)
		//HTTPS
	} else {
		_, err := conn.Write([]byte(okMessage))
		if err != nil {
			log.Println(err.Error())
		}

		hostWithoutPort := strings.TrimSuffix(host, ":443")

		rand := rand.Intn(100000001)
		randStr := strconv.Itoa(rand)

		output, err := exec.Command("/bin/sh", "gen_cert.sh", hostWithoutPort, randStr).Output()
		if err != nil {
			log.Print("Command:", err.Error())
		}
		os.WriteFile("certs/"+hostWithoutPort+".crt", output, 0666)

		certHost, err := tls.LoadX509KeyPair("certs/"+hostWithoutPort+".crt", "cert.key")
		if err != nil {
			log.Print(err.Error())
		}

		confHost := &tls.Config{
			Certificates: []tls.Certificate{certHost},
		}

		tlsConn := tls.Server(conn, confHost)
		tlsConn.Handshake()
		conn = net.Conn(tlsConn)

		dest, err := tls.Dial("tcp", host, confHost)
		if err != nil {
			log.Fatalln(err.Error())
		}

		defer dest.Close()

		request, _ := utils.CopyMessage(dest, conn)
		log.Print("Request:\n", request)
		response, _ := utils.CopyMessage(conn, dest)
		p.usecase.Save(request, response, true)

	}
}
