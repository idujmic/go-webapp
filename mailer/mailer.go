package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"strconv"
	"time"
)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}


var conn *amqp.Connection
var ch *amqp.Channel
var Session *mgo.Session

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

type MongoConfig struct {
	mongoHost string
	mongoPort string
	mongoDb   string
	username  string
	password  string
	collection    string
}


var mongoConfig MongoConfig

func readMongoConfig() {
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal(err)
	}
	mongoConfig = MongoConfig{
		mongoHost:  os.Getenv("MONGO_HOST"),
		mongoPort:  os.Getenv("MONGO_PORT"),
		mongoDb:    os.Getenv("MONGO_DB"),
		username:   os.Getenv("MONGO_USER"),
		password:   os.Getenv("MONGO_PASS"),
		collection: os.Getenv("MONGO_COLLECTION"),
	}

}

func openDBConncection(){
	readMongoConfig()
	info := &mgo.DialInfo{
		Addrs:    []string{mongoConfig.mongoHost},
		Database: mongoConfig.mongoDb,
		Username: mongoConfig.username,
		Password: mongoConfig.password,
	}
	s, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Printf("ERROR connecting mongo, %s ", err.Error())
		return
	}
	s.SetMode(mgo.Monotonic, true)
	Session = s
}
func closeDBConnection(){

	Session.Close()
}
func getGameById(gameId int) Game{
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()
	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.collection)
	game := Game{}
	err := coll.Find(bson.M{"id" : gameId}).One(&game)
	if err != nil{
		log.Printf("ERROR: no game with id, %s", gameId)
	}
	fmt.Printf("INFO: found game, %+v", game)
	return game
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
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
	fmt.Println(domain)
	fmt.Println(apiKey)
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
	return id, err
}
func main(){
	openDBConncection()

	domain := os.Getenv("MAIL_DOMAIN")
	fmt.Println(domain)
	apiKey :=os.Getenv("KEY")
	fmt.Println(apiKey)
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