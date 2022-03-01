package usecase

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	prxModels "proxy/internal/proxy/models"
	"proxy/internal/web"
	webModels "proxy/internal/web/models"
	"proxy/utils"
	"strings"

	"github.com/spf13/viper"
)

type Usecase struct {
	repo web.Repository
}

func NewUsecase(repo web.Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func htmlReplace(html string) string {
	htmlClean := html
	htmlClean = strings.ReplaceAll(htmlClean,"\\u003e",">")
	htmlClean = strings.ReplaceAll(htmlClean,"\\u003c","<")
	return htmlClean
}

func (u *Usecase) GetRequestsJson() (*webModels.RequestsJson, error) {
	requests, err := u.repo.GetRequestsJson()
	if err != nil {
		log.Print(err)
		return nil,err
	}

	for index := range requests.Requests {
		requests.Requests[index].Request = htmlReplace(requests.Requests[index].Request)
		requests.Requests[index].Response = htmlReplace(requests.Requests[index].Response)
		
	}

	return requests,nil
}

func (u *Usecase) GetRequestJson(id int) (*webModels.RequestJson, error) {
	request, err := u.repo.GetRequestJson(id)
	if err != nil {
		log.Print(err)
		return nil,err
	}
	request.Request = htmlReplace(request.Request)
	request.Response = htmlReplace(request.Response)

	return request,nil
}

func (u *Usecase) RepeatRequest(id int) (error) {
	request, err := u.repo.GetRequest(id)
	if err != nil {
		log.Print(err)
		return err
	}

	proxyHost := viper.GetString("proxy.host")
	proxyPort := viper.GetString("proxy.port")

	if !request.IsSecure {
		conn,err := net.Dial("tcp",proxyHost+":"+proxyPort)
		if err != nil {
			log.Print(err)
			return err
		}
		defer conn.Close()
		conn.Write([]byte(request.Request))
		_ = utils.ReadMessage(conn)

	} else {
		requestJson, err := u.repo.GetRequestJson(id)
		if err != nil {
			log.Print(err)
			return err
		}
		var connectMessage string
		requestStruct := &prxModels.Request{}
		json.Unmarshal([]byte(requestJson.Request),requestStruct)
		host := requestStruct.Headers["Host"].(string) + ":443"
		protocol := requestStruct.Protocol
		connectMessage = "CONNECT " + host + " " + protocol + "\r\n" + "Host: "+ host + "\r\n" +"\r\n"
		log.Print(connectMessage)

		
		conn,err := net.Dial("tcp",proxyHost+":"+proxyPort)
		if err != nil {
			log.Print(err)
		}
		defer conn.Close()
		
		conn.Write([]byte(connectMessage))
		message := utils.ReadMessage(conn)
		log.Print(message)

		confHost := &tls.Config{
			ServerName: host,
			InsecureSkipVerify: true,
		}
		
		tlsConn := tls.Client(conn,confHost)
		tlsConn.Handshake()
		conn = net.Conn(tlsConn)
		
		conn.Write([]byte(request.Request))
		response := utils.ReadMessage(conn)
		log.Print(response)
	}
	return nil

}