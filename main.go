package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	sslmode = "disable"
)

var (
	connString string
	db         *sql.DB
)

var (
	consumerKey    = os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret = os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken    = os.Getenv("TWITTER_APP_ACCESS_TOKEN")
	accessSecret   = os.Getenv("TWITTER_ACCESS_SECRET")
	dbHost         = os.Getenv("DB_HOST")
	dbPort         = os.Getenv("DB_PORT")
	dbUser         = os.Getenv("DB_USER")
	dbPass         = os.Getenv("DB_PASS")
	dbName         = os.Getenv("DB_NAME")
)

var tagList = [...]string{"cloud", "container", "devops", "aws", "microservices", "docker", "openstack", "automation", "gcp", "azure", "istio", "sre"}

type twitterUser struct {
	UserName       string
	FollowersCount int
}

func main() {
	log.Print("Starting application \n")
	connString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPass, dbName, sslmode)
	var err error
	db, err = sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Could not ping database")
	}

	r := mux.NewRouter()
	r.HandleFunc("/update", UpdateHandler).Methods("GET")
	r.HandleFunc("/fetch/{num}", FetchHandler).Methods("GET")
	r.HandleFunc("/health", HealthHandler).Methods("GET")
	log.Print("Application started \n")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	stats := db.Stats()
	ret, _ := json.Marshal("OK - Idle conns:" + strconv.Itoa(stats.Idle) + " - In use conns: " + strconv.Itoa(stats.InUse))
	w.Write(ret)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Updating database \n")
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	truncateTable()

	for _, tag := range tagList {
		searchTweetParams := &twitter.SearchTweetParams{
			Query:      "%23" + tag,
			ResultType: "recent",
			Count:      100,
		}
		search, _, _ := client.Search.Tweets(searchTweetParams)
		log.Printf("SEARCHING TWEETS (#%v): ", tag)
		for _, tweet := range search.Statuses {
			//fmt.Printf("%v - %v\n", tweet.User.Name, tweet.User.FollowersCount)
			data, err := json.Marshal(tweet)
			if err != nil {
				log.Fatalf("Could not marshall tweet data: %v\n", err)
			}
			_, err = db.Exec("insert into twitter_collector.tweet (tag, message, user_name, user_followers, data) VALUES ($1, $2, $3, $4, $5)", tag, tweet.Text, tweet.User.Name, tweet.User.FollowersCount, string(data))
			if err != nil {
				log.Fatal(err)
				log.Fatalf("Could not save tweet: %v\n", tweet)
			}
		}
		log.Printf("%v found\n", search.Metadata.Count)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	log.Print("Database update\n")
}

func truncateTable() {
	log.Print("Truncating data table \n")
	_, err := db.Exec("truncate table twitter_collector.tweet")
	if err != nil {
		log.Fatal("Could not truncate table\n")
		log.Fatal(err)
	}
	log.Print("Data table truncated \n")
}

func FetchHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Fetching data \n")
	num := mux.Vars(r)["num"]

	rows, err := db.Query("select distinct user_name, user_followers from twitter_collector.tweet order by user_followers desc limit $1", num)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []twitterUser

	for rows.Next() {
		var row twitterUser

		err = rows.Scan(&row.UserName, &row.FollowersCount)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result = append(result, row)
	}

	ret, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(ret)
	log.Print("Data fetched \n")
}
