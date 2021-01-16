package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

var conn *amqp.Connection
var ch *amqp.Channel

type GameData struct {
	Data []Game `json:"data"`
}

func GetApiGames(response http.ResponseWriter, request *http.Request) {

	url := "https://free-nba.p.rapidapi.com/games?page=0&per_page=25"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-key", "70239b9589msh86aeef721549996p1585b9jsn445e5e890b61")
	req.Header.Add("x-rapidapi-host", "free-nba.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var f GameData

	_ = json.Unmarshal(body, &f)

	//json.NewEncoder(response).Encode(f)
	collection := client.Database("ivandb").Collection("games")

	for _, game := range f.Data {
		game.Date = strings.Split(game.Date, "T")[0]
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		collection.InsertOne(ctx, game)
	}
}

func getGames(response http.ResponseWriter, request *http.Request) {
	var games []Game = getAllGames()
	var gameData = GameData{
		Data: games,
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(response, gameData)
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func postComment(response http.ResponseWriter, request *http.Request) {
	var comment Comment
	var gameId int
	data := make(map[string]interface{})
	body, _ := ioutil.ReadAll(request.Body)
	e := json.Unmarshal(body, &data)
	reqDump,_ := httputil.DumpRequest(request, true)
	if e != nil {
		log.Fatal(string(reqDump))
		log.Fatal(e)
	}
	gameId, _ = strconv.Atoi(data["game_id"].(string))
	comment = Comment{
		Content: data["content"].(string),
		Username: data["username"].(string),
	}

	createComment(comment, gameId)
	sendMessageToQueue(strconv.Itoa(gameId))
}
func sendMessageToQueue(message string) {
	err := ch.Publish(
		"comments", // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", message)

}
func setupQueue() {
	connCopy, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	conn = connCopy
	chCopy, err := connCopy.Channel()
	failOnError(err, "Failed to open a channel")
	ch = chCopy
	err = ch.ExchangeDeclare(
		"comments",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to open a channel")
}
func main() {
	openDBConncection()
	defer closeDBConnection()
	setupQueue()
	router := mux.NewRouter()
	router.HandleFunc("/api", GetApiGames).Methods("GET")
	router.HandleFunc("/", getGames).Methods("GET")
	router.HandleFunc("/postComment", postComment).Methods("POST")
	fileServer := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fileServer))
	http.ListenAndServe(":12345", router)
	defer conn.Close()
	defer ch.Close()

}
