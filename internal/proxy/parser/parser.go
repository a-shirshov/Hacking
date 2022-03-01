package parser

import (
	"strings"
	"log"
)



func putParams(params string) map[string]interface{} {
	mapParams := make(map[string]interface{})
	paramsParts := strings.Split(params, "&")
	for _, param := range paramsParts {
		keyAndValue := strings.Split(param, "=")
		key := keyAndValue[0]
		value := keyAndValue[1]
		mapParams[key] = value
	}
	return mapParams
}

func putCookies(cookies string) map[string]interface{} {
	mapCookies := make(map[string]interface{})
	cookiesParts := strings.Split(cookies, "; ")
	for _, cookie := range cookiesParts {
		keyAndValue := strings.Split(cookie, "=")
		key := keyAndValue[0]
		value := keyAndValue[1]
		mapCookies[key] = value
	}
	return mapCookies
}

//returns example.com:80 or example.com:443
func getHost(hostLine string) string {
	hostLineParts := strings.Split(hostLine, " ")
	host := hostLineParts[1]
	//Delete \n
	host = host[:len(host)-1]
	if !strings.HasSuffix(host, ":443") {
		host = host + ":80"
	}
	log.Print("Host:", host)
	return host
}

func ParseFirstLine(firstLine string) (string, string, string) {
	firstLineParts := strings.Split(firstLine, " ")
	method := firstLineParts[0]

	link := firstLineParts[1]
	link = strings.TrimPrefix(link, "http://")
	path := "/"
	linkParts := strings.Split(link, "/")
	if len(linkParts) != 1 {
		path = "/" + strings.Join(linkParts[1:], "/")
	}

	protocol := strings.TrimSuffix(firstLineParts[2], "\r")

	log.Print("path:", path)
	log.Print("Method:", method)
	log.Print("Protocol:", protocol)
	return method, path, protocol
}


func ParseMessage(message string) (string, string, string) {
	var method string
	var host string
	var path string
	var protocol string

	var resultMessage string
	MessageLines := strings.Split(message, "\n")
	for numOfLine, line := range MessageLines {
		if numOfLine == 0 {
			method, path, protocol = ParseFirstLine(line)
			firstLine := method + " " + path + " " + protocol + "\r\n"
			resultMessage += firstLine
			continue
		}

		if strings.HasPrefix(line, "Host:") {
			host = getHost(line)
		}

		if strings.HasPrefix(line, "Proxy-Connection:") {
			continue
		}

		if numOfLine != len(MessageLines)-1 {
			resultMessage += (line + "\n")
		} else {
			resultMessage += line
		}

	}
	return method, host, resultMessage
}

