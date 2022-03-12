package utils

import (
	"bufio"
	"log"
	"os"
)

func GetParamsFromFile(path string) []string {
	var params []string
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		params = append(params, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Print(err)
	}

	return params
}
