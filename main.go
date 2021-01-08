package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"time"
	"html/template"
)

var client *mongo.Client

type GameData struct{
	Data []Game `json:"data"`
}
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

func GetApiGames(response http.ResponseWriter, request *http.Request){

	url := "https://free-nba.p.rapidapi.com/games?page=0&per_page=25"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-key", "70239b9589msh86aeef721549996p1585b9jsn445e5e890b61")
	req.Header.Add("x-rapidapi-host", "free-nba.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var f GameData

	_ = json.Unmarshal(body, &f)
	//fmt.Println(res)
	//fmt.Println(string(body))

	//json.NewEncoder(response).Encode(f)
	collection := client.Database("ivandb").Collection("games")

	for _,game :=range f.Data{
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		collection.InsertOne(ctx, game)
	}
	t,_ := template.ParseFiles("index.html")
	t.Execute(response, f)
}
func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/api", GetApiGames).Methods("GET")
	fileServer := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fileServer))
	http.ListenAndServe(":12345", router)
}