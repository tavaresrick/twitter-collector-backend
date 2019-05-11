package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
	"bufio"
	"io/ioutil"
	"strings"
	"github.com/gorilla/mux"
	"encoding/json"
	"log"
	_ "github.com/lib/pq"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestFetchhHandler(t *testing.T) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPass, dbName, sslmode)
	var err error
	db, err = sql.Open("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal("Could not ping database")
	}

	file, err := os.Open("test_data.csv")
	defer file.Close()
	if err != nil {
		t.Fatal(err)
	}

	TruncateTable()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		requests := strings.Split(string(scanner.Text()), "|")
		_, err = db.Exec("insert into twitter_collector.tweet (tag, user_name, user_followers) VALUES ($1, $2, $3)", requests[0], requests[1], requests[2])
		if err != nil {
			t.Fatal(err)
		}
	}

	req, err := http.NewRequest("GET", "/fetch/5", nil)
	req = mux.SetURLVars(req, map[string]string{
		"num": "5",
	})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(FetchHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	type fetchResponse struct {
		UserName string `json:"UserName"`
		FollowersCount int `json:"FollowersCount"`
	}
	tt := []fetchResponse {
		fetchResponse{UserName: "adn40", FollowersCount: 750021},
		fetchResponse{UserName: "Harish Chand", FollowersCount: 280947},
		fetchResponse{UserName: "Evan Kirstel", FollowersCount: 244293},
		fetchResponse{UserName: "Kirk Borne", FollowersCount: 230076},
		fetchResponse{UserName: "Ronald van Loon @ #SapphireNow", FollowersCount: 183580},
	}

	respBody := make([]fetchResponse,0)
	p, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	json.Unmarshal(p,&respBody)
	if err != nil {
		t.Fatal(err)
	}

	for index, element := range respBody {
		if element.UserName != tt[index].UserName || element.FollowersCount != tt[index].FollowersCount {
			t.Fatal("API response wrong")
		}
	}
	
}

func TestUpdatethHandler(t *testing.T) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPass, dbName, sslmode)
	var err error
	db, err = sql.Open("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal("Could not ping database")
	}

	req, err := http.NewRequest("GET", "/update", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
