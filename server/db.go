package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

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
func closeDBConnection(){
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := client.Disconnect(ctx)
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println("Connection closed")
}
func getAllGames() []Game {
	var games []Game
	collection := client.Database("ivandb").Collection("games")
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var game Game
		cursor.Decode(&game)
		games=append(games, game)
	}
	if err := cursor.Err(); err != nil{
		log.Fatal(err)
	}
	return games
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
func updateGame(game Game){
	collection := client.Database("ivandb").Collection("games")
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"id": game.ID},
		bson.D{
			{"$set", bson.D{{"comments", game.Comments}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
}
func createComment(comment Comment, gameId int){
	var game Game
	game = getGameById(gameId)
	game.Comments = append(game.Comments, comment)
	updateGame(game)
}
func getCommentsForGameId(id int) []Comment{
	var game Game
	var comments []Comment
	game = getGameById(id)
	comments = game.Comments
	return comments
}