package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var client *mongo.Client
type Game struct {
	ID int `json:"id"`
	Date string `json:"date,omitempty"`
	HomeTeam Team `json:"home_team,omitempty"`
	HomeTeamScore int `json:"home_team_score"`
	Period int `json:"period"`
	PostSeason bool `json:"post_season"`
	Season int `json:"season"`
	Status string `json:"status"`
	Time string `json:"time"`
	VisitorTeam Team `json:"visitor_team"`
	VisitorTeamScore int `json:"visitor_team_score"`
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
func openDBConncection(){
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client,_ = mongo.Connect(ctx, clientOptions)
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