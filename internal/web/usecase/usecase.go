package usecase

import (
	"crypto/tls"
	"encoding/json"
	"github.com/spf13/viper"
	"log"
	"net"
	prxModels "proxy/internal/proxy/models"
	myParser "proxy/internal/proxy/parser"
	"proxy/internal/web"
	webModels "proxy/internal/web/models"
	"proxy/utils"
	"strings"
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
	htmlClean = strings.ReplaceAll(htmlClean, "\\u003e", ">")
	htmlClean = strings.ReplaceAll(htmlClean, "\\u003c", "<")
	return htmlClean
}

func sendAndGetHTTP(proxyHost string, proxyPort string, request string) error {
	conn, err := net.Dial("tcp", proxyHost+":"+proxyPort)
	if err != nil {
		log.Print(err)
		return err
	}
	defer conn.Close()
	conn.Write([]byte(request))
	_ = utils.ReadMessage(conn)
	return nil
}

func sendAndGetHTTPS(proxyHost string, proxyPort string, request string, requestStruct *prxModels.Request) error {
	var connectMessage string
	host := requestStruct.Headers["Host"].(string) + ":443"
	protocol := requestStruct.Protocol
	connectMessage = "CONNECT " + host + " " + protocol + "\r\n" + "Host: " + host + "\r\n" + "\r\n"
	log.Print(connectMessage)

	conn, err := net.Dial("tcp", proxyHost+":"+proxyPort)
	if err != nil {
		log.Print(err)
	}
	defer conn.Close()

	conn.Write([]byte(connectMessage))
	message := utils.ReadMessage(conn)
	log.Print(message)

	confHost := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}

	tlsConn := tls.Client(conn, confHost)
	tlsConn.Handshake()
	conn = net.Conn(tlsConn)

	conn.Write([]byte(request))
	response := utils.ReadMessage(conn)
	log.Print(response)
	return nil
}

func (u *Usecase) GetRequestsJson() (*webModels.RequestsJson, error) {
	requests, err := u.repo.GetRequestsJson()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	for index := range requests.Requests {
		requests.Requests[index].Request = htmlReplace(requests.Requests[index].Request)
		requests.Requests[index].Response = htmlReplace(requests.Requests[index].Response)

	}

	return requests, nil
}

func (u *Usecase) GetRequestJson(id int) (*webModels.RequestJson, error) {
	request, err := u.repo.GetRequestJson(id)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	request.Request = htmlReplace(request.Request)
	request.Response = htmlReplace(request.Response)

	resultRequest := &webModels.RequestJson{
		ID:       request.ID,
		Request:  request.Request,
		Response: request.Response,
	}

	return resultRequest, nil
}

func (u *Usecase) RepeatRequest(id int) error {
	request, err := u.repo.GetRequest(id)
	if err != nil {
		log.Print(err)
		return err
	}

	proxyHost := viper.GetString("proxy.host")
	proxyPort := viper.GetString("proxy.port")

	if !request.IsSecure {
		err := sendAndGetHTTP(proxyHost, proxyPort, request.Request)

		if err != nil {
			return err
		}

	} else {
		requestJson, err := u.repo.GetRequestJson(id)
		if err != nil {
			log.Print(err)
			return err
		}
		requestStruct := &prxModels.Request{}
		json.Unmarshal([]byte(requestJson.Request), requestStruct)
		request := myParser.JsonToRequest(requestStruct)
		err = sendAndGetHTTPS(proxyHost, proxyPort, request, requestStruct)
		if err != nil {
			return err
		}
	}
	return nil

}

func (u *Usecase) ScanRequest(id int, params *[]string) error {
	requestJson, err := u.repo.GetRequestJson(id)
	if err != nil {
		log.Print(err)
		return err
	}

	//Лицом по клавиатуре
	randStr := "uigrsbgrivbisbvilsbviebgbwe4irbi345ri2brwirbwb"

	proxyHost := viper.GetString("proxy.host")
	proxyPort := viper.GetString("proxy.port")

	request := &prxModels.Request{}
	json.Unmarshal([]byte(requestJson.Request), request)

	for _, param := range *params {
		exposedRequest := request
		exposedRequest.Params[param] = randStr
		exposedRequestStr := myParser.JsonToRequest(exposedRequest)
		if !requestJson.IsSecure {
			err := sendAndGetHTTP(proxyHost, proxyPort, exposedRequestStr)
			if err != nil {
				log.Print("Err:", err)
			}
		} else {
			err := sendAndGetHTTPS(proxyHost, proxyPort, exposedRequestStr, exposedRequest)
			if err != nil {
				log.Print("Err:", err)
				return err
			}
		}

	}
	return err
}
