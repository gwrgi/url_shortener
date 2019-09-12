package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
	"log"
	"math/rand"
	"net/http"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const shortenedUrlLength = 16

func generateShortUrlId() string {
	b := make([]byte, shortenedUrlLength)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type ShortenedUrl struct {
	LongUrl string `json: "longUrl"`
}

var urlMap map[string]string
var router *httprouter.Router

func doesShortenedUrlIdExist(shortenedUrlId string) bool {
	if _, ok := urlMap[shortenedUrlId]; ok {
		return true
	}

	return false
}

func getLongUrl(shortUrlId string) (string, error) {
	if longUrl, ok := urlMap[shortUrlId]; ok {
		return longUrl, nil
	}

	return "", errors.New("No such URL exists")
}

func longUrlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlReq := ShortenedUrl{}

	err := json.NewDecoder(r.Body).Decode(&urlReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request Body")
		return
	}

	if !govalidator.IsURL(urlReq.LongUrl) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Long Url")
		return
	}

	var shortenedUrlId string
	for {
		shortenedUrlId = generateShortUrlId()
		if !doesShortenedUrlIdExist(shortenedUrlId) {
			break
		}
	}

	shortenedUrl := "http://localhost:8080/" + shortenedUrlId

	// TODO: Store this in a database, not in memory, or else it won't survive restarts
	urlMap[shortenedUrlId] = urlReq.LongUrl

	router.GET("/"+shortenedUrlId, shortUrlHandler)

	fmt.Fprintf(w, shortenedUrl)
}

func shortUrlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlId := r.URL.Path[len("/"):]
	longUrl, err := getLongUrl(urlId)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, longUrl, http.StatusTemporaryRedirect)
}

func pingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "pong")
}

func main() {
	urlMap = make(map[string]string)

	router = httprouter.New()
	router.GET("/v1/ping", pingHandler)
	router.POST("/v1/create", longUrlHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}
