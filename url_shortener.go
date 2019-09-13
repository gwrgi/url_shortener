package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const shortenedUrlLength = 7

var database *sql.DB
var router *httprouter.Router

func generateShortUrlId() string {
	b := make([]byte, shortenedUrlLength)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type ShortenedUrl struct {
	ShortId       string `json: "shortId"`
	LongUrl       string `json: "longUrl"`
	Visits24Hours int    `json: "visits24Hours"`
	Visits7Days   int    `json: "visits7Days"`
	VisitsAllTime  int    `json: "visitsAllTime"`
}

func doesShortenedUrlIdExist(shortenedUrlId string) (bool, error) {
	var count int
	err := database.QueryRow("SELECT COUNT(1) FROM urlmap WHERE shorturlid = ?", shortenedUrlId).Scan(&count)
	if err != nil {
		return false, errors.New("Database error occurred: 0x2")
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func saveHistory(shortUrlId string) error {
	visitedTime := time.Now().Unix()

	statement, err := database.Prepare("INSERT INTO history (shorturlid, visited) VALUES (?, ?)")
	if err == nil {
		_, err = statement.Exec(shortUrlId, visitedTime)
	}

	return err
}

/*
  Retrieves the number of times a given short url has been vistied since now minus duration seconds.
  If duration is negative, count of all visits is returned
*/
func getHistoryCount(shortUrlId string, duration int64) (int, error) {
	var since int64 = 0
	if duration > 0 {
		since = time.Now().Unix() - duration
	}
	
	var count int
	err := database.QueryRow("SELECT COUNT(1) FROM history WHERE shorturlid = ? AND visited > ?", shortUrlId, since).Scan(&count)
	if err != nil {
		return -1, errors.New("Database error occurred: 0x3")
	}

	return count, nil
}

func getLongUrl(shortUrlId string) (string, error) {
	rows, err := database.Query("SELECT longurl FROM urlmap WHERE shorturlid = ?", shortUrlId)
	defer rows.Close()

	if err != nil {
		return "", errors.New("Database error occurred: 0x1")
	}

	var longUrl string
	for rows.Next() {
		rows.Scan(&longUrl)
		break // Should only be one...
	}

	return longUrl, nil
}

func insertIntoDb(shortUrlId, longUrl string) error {
	statement, err := database.Prepare("INSERT INTO urlmap (shorturlid, longurl) VALUES (?, ?)")
	if err == nil {
		_, err = statement.Exec(shortUrlId, longUrl)
	}

	return err
}

func saveUrl(longUrl string) (string, error) {
	var shortId string
	for {
		shortId = generateShortUrlId()
		exists, err := doesShortenedUrlIdExist(shortId)
		if err != nil {
			return "", err
		}
		if !exists {
			break
		}
	}

	err := insertIntoDb(shortId, longUrl)
	if err != nil {
		return "", err
	}

	return shortId, nil
}

func longUrlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlReq := ShortenedUrl{}

	err := json.NewDecoder(r.Body).Decode(&urlReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Request Body")
		return
	}

	_, err = url.ParseRequestURI(urlReq.LongUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Long Url")
		return
	}

	shortId, err := saveUrl(urlReq.LongUrl)

	router.GET("/"+shortId, shortUrlHandler)

	shortUrl := "http://localhost:8080/" + shortId
	fmt.Fprintf(w, shortUrl)
}

func shortUrlHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlId := r.URL.Path[len("/"):]
	longUrl, err := getLongUrl(urlId)

	err = saveHistory(urlId)
	if err != nil {
		// This is not great, and should be looked into if it ever occurs, but should NOT
		// stop the user from getting their long url
		log.Printf("Error occurred while saving history. urlId: '%s', err: %s", urlId, err)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	if longUrl == "" {
		// This represents a buggy state.  The router does not have wildcards, so the only way to get to this function is by having the urlId
		// actually have been seen before.  If the longUrl it maps to is empty, then either an empty longUrl was erroneously accepted, or
		// somehow it didn't get saved to the database.  As far as the user is concerned, however, the url they are looking for was not found
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longUrl, http.StatusTemporaryRedirect)
}

func infoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlId := ps.ByName("shortId")
	longUrl, err := getLongUrl(urlId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	visits24Hours, err := getHistoryCount(urlId, 60 * 60 * 24)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	visits7Days, err := getHistoryCount(urlId, 60 * 60 * 24 * 7)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	visitsAllTime, err := getHistoryCount(urlId, -1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	data := ShortenedUrl{
		LongUrl: longUrl,
		ShortId: urlId,
		Visits24Hours: visits24Hours,
		Visits7Days: visits7Days,
		VisitsAllTime: visitsAllTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func pingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "pong")
}

func initializeDatabase() error {
	db, err := sql.Open("sqlite3", "./database.db")
	database = db

	if err == nil {
		createUrlMapStatment, err := database.Prepare("CREATE TABLE IF NOT EXISTS urlmap (id INTEGER PRIMARY KEY, shorturlid TEXT, longurl TEXT)")
		createHistoryStatment, err := database.Prepare("CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, shorturlid TEXT, visited INTEGER)")

		if err == nil {
			_, err = createUrlMapStatment.Exec()
			if err == nil {
				_, err = createHistoryStatment.Exec()
			}
		}

	}

	return err
}

func addExistingRoutesFromDatabase() error {
	rows, err := database.Query("SELECT shorturlid FROM urlmap")
	if err == nil {
		defer rows.Close()
		var shortUrlId string
		for rows.Next() {
			rows.Scan(&shortUrlId)
			router.GET("/"+shortUrlId, shortUrlHandler)
		}
	}

	return err
}

func main() {
	router = httprouter.New()
	router.GET("/v1/ping", pingHandler)
	router.POST("/v1/create", longUrlHandler)
	router.GET("/v1/get/*shortId", infoHandler)

	err := initializeDatabase()
	if err != nil {
		log.Fatal("An error occurred in database initialization: %s", err.Error())
		return
	}

	err = addExistingRoutesFromDatabase()
	if err != nil {
		log.Fatal("An error occurred while setting up existing routes: %s", err.Error())
		return
	}

	log.Printf("Ready...")

	log.Fatal(http.ListenAndServe(":8080", router))
}
