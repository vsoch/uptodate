package utils

import (
	"io/ioutil"
	"log"
	"net/http"
)

func GetRequest(url string) string {

	response, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	// Read the response from the body, and return as string
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return string(body)
}
