package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mailgun/mailgun-go"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

var apiKey =  "4efbe29e9259350f967c4b334cfaca3d-28d78af2-1ce9dbde"

var domain ="sandboxa5aefc9af0da411ea218db36551cd093.mailgun.org"
var conn *amqp.Connection
var ch *amqp.Channel
var client *mongo.Client
type Game struct {
	ID               int    `json:"id"`
	Date             string `json:"date,omitempty"`
	HomeTeam         Team   `json:"home_team,omitempty"`
	HomeTeamScore    int    `json:"home_team_score"`
	Period           int    `json:"period"`
	PostSeason       bool   `json:"post_season"`
	Season           int    `json:"season"`
	Status           string `json:"status"`
	Time             string `json:"time"`
	VisitorTeam      Team   `json:"visitor_team"`
	VisitorTeamScore int    `json:"visitor_team_score"`
	Comments [] Comment `json:"comments,omitempty"`
}
type Team struct{
	ID int `json:"id"`
	Abbrevation string `json:"abbrevation"`
	City string `json:"city"`
	Conference string `json:"conference"`
	Division string `json:"division"`
	FullName string `json:"full_name"`
	Name string `json:"name"`
}
type Comment struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string `json:"username"`
	Content string `json:"content"`
}

func getGameById(gameId int) Game{
	var game Game
	collection := client.Database("ivandb").Collection("games")
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	err := collection.FindOne(ctx, bson.M{"id" : gameId}).Decode(&game)
	if err!=nil{
		log.Fatal(err)
	}
	return game
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func openDBConncection(){
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client,err = mongo.Connect(ctx, clientOptions)
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println("Connection with the database is open")
}
func getHtmlPreview(game Game) string {
	lastComment := game.Comments[len(game.Comments)-1]
	lastCommentUsername := lastComment.Username
	lastCommentContent := lastComment.Content
	homeTeam := game.HomeTeam.FullName
	visitorTeam := game.VisitorTeam.FullName

	 htmlMsg := "<h2>" + homeTeam + " - " + visitorTeam+ "</h2>"+
		"<p>"  + lastCommentUsername +  " : " + lastCommentContent + "</p>"
	 return htmlMsg
}
func SendSimpleMessage(domain, apiKey, gameId string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	gameIdInt,_ := strconv.Atoi(gameId)
	game := getGameById(gameIdInt)
	htmlMsg := getHtmlPreview(game)
	sender := "nbaapp@dujma.hr"
	subject := "Comment added!"
	plainText :="A new comment has been added\n"
	recipient := "ivan.dujmic96@gmail.com"
	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(sender, subject,plainText, recipient)
	message.SetHtml(htmlMsg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
	return id, err
}
func main(){
	openDBConncection()
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"comments",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"comments", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			id, err := SendSimpleMessage(domain, apiKey, string(d.Body))
			fmt.Println(id)
			fmt.Println(err)
		}
	}()
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever

}