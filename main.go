package main

import (
	//"bufio"
	//"bytes"
	//"encoding/binary"
	//"bytes"
	//"bufio"
	"bufio"
	"io"
	"log"
	"net"

	//"os"
	"os/exec"
	"strconv"
	"strings"

	"os"
	"time"

	//"time"
	//"strings"
	"crypto/tls"
)

/*
func readSNI(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5*time.Second))
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf,conn,1+2+2); err != nil {
		log.Println(err)
		return
	}
	length := binary.BigEndian.Uint16(buf.Bytes()[3:5])
	if _, err := io.CopyN(&buf,conn,int64(length)); err != nil {
		log.Println(err)
		return
	}

	ch, ok := ParseClientHello(buf.Bytes())
	if ok {
		log.Println(ch.SNI)
	}
}*/

//HTTPS
//Host is definetly {site.com}:443 and Method is CONNECT

//HTTP
//Host is http://{site.com}:80 or http://{site.com}

const okMessage = "HTTP/1.0 200 Connection established\r\n\r\n"

func ParseFirstLine(firstLine string) (string, string, string) {
	firstLineParts := strings.Split(firstLine, " ")
	method := firstLineParts[0]

	link := firstLineParts[1]
	link = strings.TrimPrefix(link, "http://")
	url := "/"
	linkParts := strings.Split(link, "/")
	if len(linkParts) != 1 {
		url = "/" + strings.Join(linkParts[1:], "/")
	}

	protocol := firstLineParts[2]

	log.Print("Url:", url)
	log.Print("Method:", method)
	log.Print("Protocol:", protocol)
	return method, url, protocol
}

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

func MyRead(conn net.Conn) string {
	var buf [64]byte
	var message string
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		size, err := conn.Read(buf[:])
		log.Println("Size", size)
		if err != nil {
			log.Print(err)
			return message
		}
		message += string(buf[:size])

	}
}

func CopyFromConn(conn net.Conn) string {
	var message string
	contentLength := 0

	//Case: Content-Length: len(body)
	//Case: Chunked: till \r\n
	//Case: None: 0
	bodyReadMode := "None"

	connReader := bufio.NewReader(conn)
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

		if line == "\r\n" {
			switch bodyReadMode {
			case "Content-Length":
				if contentLength != 0 {
					buf := make([]byte, contentLength)
					_, err := io.ReadFull(connReader, buf)
					if err != nil {
						log.Print(err.Error())
					}

					message += string(buf[:])
				}

			case "Chunked":
				for {
					line, err := connReader.ReadString('\n')
					if err != nil {
						log.Print(err.Error())
						return ""
					}

					//log.Print("Line:", line)
					message += line

					number := strings.TrimSuffix(line, "\r\n")
					chunkLength, err := strconv.ParseInt(number, 16, 0)
					if err != nil {
						log.Print("not number", line)
						break
					}

					log.Print("Chunk Length:", chunkLength)

					buf := make([]byte, chunkLength)
					_, err = io.ReadFull(connReader, buf)
					if err != nil {
						log.Print(err.Error())
					}

					message += string(buf[:])

					line, err = connReader.ReadString('\n')
					if err != nil {
						log.Print(err.Error())
						return ""
					}

					log.Print("Line:", line)
					message += line

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

func ParseMessage(message string) (string, string, string) {
	var method string
	var host string
	var url string
	var protocol string

	var resultMessage string
	MessageLines := strings.Split(message, "\n")
	for numOfLine, line := range MessageLines {
		if numOfLine == 0 {
			method, url, protocol = ParseFirstLine(line)
			firstLine := method + " " + url + " " + protocol + "\n"
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
	log.Print("Result Message:\n", resultMessage)
	log.Print("Result Bytes:", []byte(resultMessage))
	return method, host, resultMessage
}

func Handler(conn net.Conn) {
	defer conn.Close()
	message := CopyFromConn(conn)
	log.Print("Message:\n", message)
	method, host, modMessage := ParseMessage(message)

	if method != "CONNECT" {
		dest, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer dest.Close()

		_, err = dest.Write([]byte(modMessage))
		if err != nil {
			log.Println(err.Error())
		}

		responseFromServer := CopyFromConn(dest)

		log.Print("Response:\n", responseFromServer)
		_, err = conn.Write([]byte(responseFromServer))
		if err != nil {
			log.Println(err.Error())
		}
	} else {

		_, err := conn.Write([]byte(okMessage))
		if err != nil {
			log.Println(err.Error())
		}

		hostWithoutPort := strings.TrimSuffix(host, ":443")

		output, err := exec.Command("/bin/sh", "./gen_cert.sh", hostWithoutPort, "1000").Output()
		if err != nil {
			log.Print(err.Error())
		}

		os.WriteFile("certs/"+hostWithoutPort+".crt", output, 0666)

		/*
			cert, err := tls.LoadX509KeyPair("ca.crt","ca.key")
			if err != nil {
				log.Print(err.Error())
			}

			conf := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		*/
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

		superMessage := CopyFromConn(conn)
		log.Print("Super Message:\n", superMessage)

		dest, err := tls.Dial("tcp", host, confHost)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer dest.Close()

		dest.Write([]byte(superMessage))

		replyMessage := CopyFromConn(dest)
		log.Print("Reply Message:\n", replyMessage)

		conn.Write([]byte(replyMessage))

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
