package usecase

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	prxModels "proxy/internal/proxy/models"
	myParser "proxy/internal/proxy/parser"
	"proxy/internal/web"
	webModels "proxy/internal/web/models"
	"proxy/utils"
	"strings"
	"time"

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
	htmlClean = strings.ReplaceAll(htmlClean, "\\u003e", ">")
	htmlClean = strings.ReplaceAll(htmlClean, "\\u003c", "<")
	return htmlClean
}

func sendAndGetHTTP(proxyHost string, proxyPort string, request string) (string, error) {
	conn, err := net.Dial("tcp", proxyHost+":"+proxyPort)
	if err != nil {
		log.Print(err)
		return "", err
	}
	defer conn.Close()
	conn.Write([]byte(request))
	response := utils.ReadMessage(conn)
	return response, nil
}

func makeSecureConn(proxyHost string, proxyPort string, requestStruct *prxModels.Request) (net.Conn, error) {
	var connectMessage string
	host := requestStruct.Headers["Host"].(string) + ":443"
	protocol := requestStruct.Protocol
	connectMessage = "CONNECT " + host + " " + protocol + "\r\n" + "Host: " + host + "\r\n" + "\r\n"
	log.Print(connectMessage)

	conn, err := net.DialTimeout("tcp", proxyHost+":"+proxyPort, time.Second*10)
	if err != nil {
		log.Print(err)
		return nil, err
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
	return conn, nil
}

func sendAndGetHTTPS(conn net.Conn, request string) (string, error) {
	conn.Write([]byte(request))
	log.Print(request)
	response := utils.ReadMessage(conn)
	log.Print(response)
	return response, nil
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
		_, err := sendAndGetHTTP(proxyHost, proxyPort, request.Request)

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

		conn, err := makeSecureConn(proxyHost, proxyPort, requestStruct)
		if err != nil {
			return err
		}
		_, err = sendAndGetHTTPS(conn, request)
		if err != nil {
			return err
		}
	}
	return nil

}

func checkResponse(response string, randStr *string) bool {
	return strings.Contains(response, *randStr)
}

func (u *Usecase) ScanRequest(id int, params *[]string) (*[]string, error) {
	requestJson, err := u.repo.GetRequestJson(id)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	//Лицом по клавиатуре
	randStr := "uigrsbgrivbisbvilsbviebgbwe4irbi345ri2brwirbwb"

	proxyHost := viper.GetString("proxy.host")
	proxyPort := viper.GetString("proxy.port")

	request := &prxModels.Request{}
	json.Unmarshal([]byte(requestJson.Request), request)
	var exposedParams []string

	for _, param := range *params {
		var response string

		exposedRequest := &prxModels.Request{
			Method:   request.Method,
			Path:     request.Path,
			Protocol: request.Protocol,
			Params:   request.Params,
			Headers:  request.Headers,
			Cookies:  request.Cookies,
			Body:     request.Body,
		}

		if exposedRequest.Params == nil {
			exposedRequest.Params = make(map[string]interface{})
		}
		exposedRequest.Params[param] = randStr
		exposedRequestStr := myParser.JsonToRequest(exposedRequest)

		if !requestJson.IsSecure {
			response, err = sendAndGetHTTP(proxyHost, proxyPort, exposedRequestStr)
		} else {
			conn, err := makeSecureConn(proxyHost, proxyPort, exposedRequest)
			if err != nil {
				return nil, err
			}
			defer conn.Close()

			response, err = sendAndGetHTTPS(conn, exposedRequestStr)
			if err != nil {
				log.Print(err)
				return nil, err
			}
			conn.Close()
		}

		if err != nil {
			log.Print("Err:", err)
			return nil, err
		}

		if checkResponse(response, &randStr) {
			exposedParams = append(exposedParams, param)
		}

	}
	return &exposedParams, err
}
