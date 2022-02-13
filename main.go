package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

func parseFirstLine(reader *bufio.Reader, message *string) string {
	line, errorMessage := reader.ReadString('\n')
	if errorMessage != nil {
		log.Fatalln(errorMessage.Error())
	}
	
	//Достаём host Убираем http:// и :80
	metaInfo := strings.Split(line, " ")
	fullUrl := metaInfo[1]
	fullUrl = strings.Replace(fullUrl, "http://", "", 1)
	fullUrl = strings.TrimSuffix(fullUrl, "/")
	hostNameArray := strings.Split(fullUrl,"/")
	//Находим относительный урл
	url := "/"
	if len(hostNameArray) != 1 {
		url = "/" + strings.Join(hostNameArray[1:],"/")
	}
	log.Println("Url:",url)
	log.Println("HostName:",hostNameArray)
	//Записываем в сообщение первую строку с новым урл
	*message += metaInfo[0] + " " + url + " " + metaInfo[2]
	hostName := strings.TrimSuffix(hostNameArray[0],":80")
	return hostName
}

func Handler(conn net.Conn) {
	var message string
	defer conn.Close()
	//Считываем сообщение
	connReader := bufio.NewReader(conn)
	//Парсим первую строку для нахождения куда отправить
	//PS. Почему-то если не начать считывать, то ты длину сообщения не поймёшь
	hostName := parseFirstLine(connReader,&message)
	//Коннектимся к серверу
	dest, err := net.Dial("tcp", "["+hostName+"]:80")
	if err != nil {
		log.Fatalln(err.Error())
	}

	defer dest.Close()

	var messageSize int
	//Длина того, что есть в ридере, то есть наше сообщение
	messageSize = connReader.Buffered()

	//Считываем сообщение и удаляем заголовок
	for {
		line, errorMessage := connReader.ReadString('\n')
		if errorMessage != nil {
			log.Fatalln(err.Error())
		}
		if !strings.HasPrefix(line, "Proxy-Connection:") {
			message += line
		}
		messageSize -= len(line)
		if messageSize == 0 {
			break
		}
	}
	//Отправляем
	log.Print("Message:\n", message)
	dest.Write([]byte(message))

	//Копируем ответ
	_, err = io.Copy(conn, dest)
	if err != nil {
		log.Fatalln(err.Error())
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
