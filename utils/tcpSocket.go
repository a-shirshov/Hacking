package utils

import (
	"bufio"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

func readBytesN(reader io.Reader, size int64, message *string) {
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
				readBytesN(connReader, int64(contentLength), &message)
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
					readBytesN(connReader, totalChunkSize, &message)

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

func CopyMessage(to net.Conn, from net.Conn) (string, error) {
	message := ReadMessage(from)
	_, err := to.Write([]byte(message))
	if err != nil {
		return "", err
	}
	return message, nil
}
