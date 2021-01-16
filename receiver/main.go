package main

import (
	"log"
	"net/http"

	"github.com/streadway/amqp"
	"github.com/gorilla/websocket"

)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

var wsGlobal *websocket.Conn
var clients = make(map[*websocket.Conn]bool)
var conn *amqp.Connection
var ch *amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func reader(ws *websocket.Conn){
	clients[ws] = true
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil{
			log.Println(err)
			return
		}
		log.Println(string(p))
		if err := ws.WriteMessage(messageType, p); err != nil{
			log.Println(err)
		}
	}


}
func wsEndpoint(response http.ResponseWriter, request *http.Request)  {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true}
	ws, err := upgrader.Upgrade(response, request, nil)
	if err != nil{
		log.Println(err)
	}
	log.Println("Client Successfully connected")
	reader(ws)
}

func main() {

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
	go func() {
		for d := range msgs {
			for client := range clients{
					if err := client.WriteMessage(websocket.TextMessage, d.Body); err != nil{
						log.Println(err)
						delete(clients, client)
					}
					log.Printf("Received a message: %s", d.Body)
			}
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	http.HandleFunc("/ws", wsEndpoint)
	log.Fatal(http.ListenAndServe(":8080", nil))

}