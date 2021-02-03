package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"log"
	"os"
)


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
func getAllGames() []Game {
	var games []Game
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()
	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.collection)
	err := coll.Find(bson.M{}).All(&games)
	if err != nil {
		fmt.Printf("ERROR: fail get msgs, %s", err.Error())
	}
	return games
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
func updateGame(game Game){
	fmt.Println("uso u update game")
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()
	var coll = sessionCopy.DB(mongoConfig.mongoDb).C(mongoConfig.collection)
	err := coll.Update(bson.M{"id": game.ID}, bson.M{"$set": bson.M{"comments": game.Comments}})
	if err != nil {
		log.Fatal(err)
	}
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
