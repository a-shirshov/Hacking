package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"crypto/tls"
	"os"
)

//HTTPS
//Host is definetly {site.com}:443 and Method is CONNECT

//HTTP
//Host is http://{site.com}:80 or http://{site.com}

const okMessage = "HTTP/1.0 200 Connection established\r\n\r\n"

type request struct {
	Method string `json:"method"`
	Path string `json:"path"`
	Protocol string `json:"protocol"`
	Params map[string]interface{} `json:"params,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Cookies map[string]interface{} `json:"cookies,omitempty"`
	Body string `json:"body,omitempty"`
}

type response struct {
	Protocol string `json:"protocol"`
	Code string `json:"code"`
	Message string `json:"message"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body string `json:"body,omitempty"`
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

	protocol := strings.TrimSuffix(firstLineParts[2],"\r")

	log.Print("path:", path)
	log.Print("Method:", method)
	log.Print("Protocol:", protocol)
	return method, path, protocol
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

func ReadBytesN(reader io.Reader, size int64, message *string) {
	buf := make([]byte, size)
	_, err := io.ReadFull(reader, buf)
		if err != nil {
			log.Print(err.Error())
		}

	*message += string(buf[:])
}

func ReadMessage(conn net.Conn) string {
	var message string
	contentLength := 0

	//Case: Content-Length: len(body)
	//Case: Chunked: till \r\n
	//Case: None: 0
	bodyReadMode := "None"

	connReader := bufio.NewReader(conn)
	//Read Headers
	for {
		line, err := connReader.ReadString('\n')
		if err != nil {
			log.Print(err.Error())
			return ""
		}

		if strings.HasPrefix(line, "Content-Length:") {
			bodyReadMode = "Content-Length"
			lineParts := strings.Split(line, " ")
			strLength := strings.TrimSuffix(lineParts[1], "\r\n")
			contentLength, err = strconv.Atoi(strLength)

			if err != nil {
				log.Print(err.Error())
			}
		}

		if strings.HasPrefix(line, "Transfer-Encoding: chunked") {
			bodyReadMode = "Chunked"
		}

		message += line

		//Headers ended, going to body
		if line == "\r\n" {
			switch bodyReadMode {
			case "Content-Length":
				ReadBytesN(connReader,int64(contentLength),&message)
			case "Chunked":
				for {
					line, err := connReader.ReadString('\n')
					if err != nil {
						log.Print(err.Error())
						return ""
					}
					message += line
					number := strings.TrimSuffix(line, "\r\n")
					chunkLength, err := strconv.ParseInt(number, 16, 0)
					if err != nil {
						log.Print("not number", line)
						break
					}

					//chunkLength counts without \r\n so we need to add 2 bytes
					totalChunkSize := chunkLength + 2
					ReadBytesN(connReader,totalChunkSize,&message)

					if chunkLength == 0 {
						break
					}

				}

			default:

			}

			return message
		}
	}
}

func CopyMessage(to net.Conn, from net.Conn) (string,error) {
	message := ReadMessage(from)
	_, err := to.Write([]byte(message))
	if err != nil {
		return "",err
	}
	return message, nil
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

func putParams(params string) map[string]interface{} {
	mapParams := make(map[string]interface{})
	paramsParts := strings.Split(params,"&")
	for _,param := range paramsParts {
		keyAndValue := strings.Split(param,"=")
		key := keyAndValue[0]
		value := keyAndValue[1]
		mapParams[key] = value
	}
	return mapParams
}

func putCookies(cookies string) map[string]interface{} {
	mapCookies := make(map[string]interface{})
	cookiesParts := strings.Split(cookies,"; ")
	for _,cookie := range cookiesParts {
		keyAndValue := strings.Split(cookie,"=")
		key := keyAndValue[0]
		value := keyAndValue[1]
		mapCookies[key] = value
	}
	return mapCookies
}

func RequestToJson(message string) string {
	jsonRequest := &request{}
	mapHeaders := make(map[string]interface{})
	messageParts := strings.Split(message,"\r\n")

	//Parse 
	for index, line := range messageParts {
		if index == 0 {
			method, path,protocol := ParseFirstLine(line)
			jsonRequest.Method = method
			jsonRequest.Protocol = protocol
			pathParts := strings.Split(path,"?")
			jsonRequest.Path = pathParts[0]
			jsonRequest.Params = putParams(pathParts[1])
			continue
		}

		if strings.HasPrefix(line,"Cookie:") {
			cookies := strings.TrimPrefix(line,"Cookie: ")
			jsonRequest.Cookies = putCookies(cookies)
			continue
		}
		//Delimeter beetween headers and body
		if line != "" {
			lineParts := strings.Split(line,": ")
			headerName := lineParts[0]
			log.Print(headerName)
			headerValue := lineParts[1]
			log.Print(headerValue)
			mapHeaders[headerName] = headerValue
		} else {
			jsonRequest.Headers = mapHeaders
			body := strings.Join(messageParts[index:],"")
			jsonRequest.Body = body
			break
		}
	}
	
	result, err := json.Marshal(jsonRequest)
	if err != nil {
		log.Print(err.Error())
	}

	log.Print("Json:",string(result))
	return ""
}

func ResponseToJson(message string) string {
	jsonResponse := &response{}
	mapHeaders := make(map[string]interface{})
	messageParts := strings.Split(message,"\r\n")

	//Parse 
	for index, line := range messageParts {
		if index == 0 {
			firstLineParts := strings.Split(line," ")
			jsonResponse.Protocol = firstLineParts[0]
			jsonResponse.Code = firstLineParts[1]
			jsonResponse.Message = strings.Join(firstLineParts[2:]," ")
			continue
		}

		//Delimeter beetween headers and body
		if line != "" {
			lineParts := strings.Split(line,": ")
			headerName := lineParts[0]
			log.Print(headerName)
			headerValue := lineParts[1]
			log.Print(headerValue)
			mapHeaders[headerName] = headerValue
		} else {
			jsonResponse.Headers = mapHeaders
			body := strings.Join(messageParts[index:],"")
			jsonResponse.Body = body
			break
		}
	}
	
	result, err := json.Marshal(jsonResponse)
	if err != nil {
		log.Print(err.Error())
	}

	log.Print("Json:",string(result))
	return ""

}


func Handler(conn net.Conn) {
	defer conn.Close()
	message := ReadMessage(conn)
	log.Print("Message:\n", message)
	method, host, modMessage := ParseMessage(message)

	//HTTP
	if method != "CONNECT" {
		dest, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer dest.Close()

		_ = RequestToJson(modMessage)

		_, err = dest.Write([]byte(modMessage))
		if err != nil {
			log.Print(err.Error())
		}
		response,err := CopyMessage(conn,dest)
		if err != nil {
			log.Print(err.Error())
		}
		_ = ResponseToJson(response)
	//HTTPS
	} else {

		_, err := conn.Write([]byte(okMessage))
		if err != nil {
			log.Println(err.Error())
		}

		hostWithoutPort := strings.TrimSuffix(host, ":443")

		rand := rand.Intn(100000001)
		randStr := strconv.Itoa(rand)

		output, err := exec.Command("/bin/sh", "./gen_cert.sh", hostWithoutPort, randStr).Output()
		if err != nil {
			log.Print(err.Error())
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

		CopyMessage(dest,conn)
		CopyMessage(conn,dest)

	}
}

func main() {
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
		go Handler(conn)
	}

}
