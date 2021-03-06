package parser

import (
	"encoding/json"
	"log"
	prxModels "proxy/internal/proxy/models"
	"strings"
)

func RequestToJson(message string) string {
	jsonRequest := &prxModels.Request{}
	mapHeaders := make(map[string]interface{})
	messageParts := strings.Split(message, "\r\n")

	//Parse
	for index, line := range messageParts {
		if index == 0 {
			method, path, protocol := ParseFirstLine(line)
			jsonRequest.Method = method
			jsonRequest.Protocol = protocol
			pathParts := strings.Split(path, "?")
			jsonRequest.Path = pathParts[0]
			if len(pathParts) > 1 {
				log.Println("Before Params")
				jsonRequest.Params = putParams(pathParts[1])
			}
			continue
		}

		if strings.HasPrefix(line, "Cookie:") {
			cookies := strings.TrimPrefix(line, "Cookie: ")
			jsonRequest.Cookies = putCookies(cookies)
			continue
		}
		//Delimeter beetween headers and body
		if line != "" {
			lineParts := strings.Split(line, ": ")
			headerName := lineParts[0]
			headerValue := lineParts[1]
			mapHeaders[headerName] = headerValue
		} else {
			jsonRequest.Headers = mapHeaders
			body := strings.Join(messageParts[index:], "")
			jsonRequest.Body = body
			break
		}
	}

	result, err := json.Marshal(jsonRequest)
	if err != nil {
		log.Print(err.Error())
	}

	return string(result)
}

func ResponseToJson(message string) string {
	jsonResponse := &prxModels.Response{}
	mapHeaders := make(map[string]interface{})
	messageParts := strings.Split(message, "\r\n")

	//Parse
	for index, line := range messageParts {
		if index == 0 {
			firstLineParts := strings.Split(line, " ")
			jsonResponse.Protocol = firstLineParts[0]
			jsonResponse.Code = firstLineParts[1]
			jsonResponse.Message = strings.Join(firstLineParts[2:], " ")
			continue
		}

		//Delimeter beetween headers and body
		if line != "" {
			lineParts := strings.Split(line, ": ")
			headerName := lineParts[0]
			headerValue := lineParts[1]
			mapHeaders[headerName] = headerValue
		} else {
			jsonResponse.Headers = mapHeaders
			body := strings.Join(messageParts[index:], "")
			jsonResponse.Body = body
			break
		}
	}

	result, err := json.Marshal(jsonResponse)
	if err != nil {
		log.Print(err.Error())
	}

	return string(result)
}

func JsonToRequest(request *prxModels.Request) string {
	var resultRequest string
	var firstLine string
	paramsCount := len(request.Params)
	if paramsCount > 0 {
		counter := 0
		url := request.Path + "?"
		for key, value := range request.Params {
			param := key + "=" + value.(string)
			url += param
			if counter != (paramsCount - 1) {
				url += "&"
			}
		}
		firstLine = request.Method + " " + url + " " + request.Protocol + "\r\n"
	} else {
		firstLine = request.Method + " " + request.Path + " " + request.Protocol + "\r\n"
	}

	resultRequest += firstLine
	for key, value := range request.Headers {
		line := key + ": " + value.(string) + "\r\n"
		resultRequest += line
	}

	resultRequest += "\r\n"
	resultRequest += request.Body
	return resultRequest
}
