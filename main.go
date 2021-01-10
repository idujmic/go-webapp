package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)


type GameData struct{
	Data []Game `json:"data"`
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

	//json.NewEncoder(response).Encode(f)
	collection := client.Database("ivandb").Collection("games")

	for _,game :=range f.Data{
		game.Date = strings.Split(game.Date, "T")[0]
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		collection.InsertOne(ctx, game)
	}
}

func getGames(response http.ResponseWriter, request *http.Request){
	var games []Game = getAllGames()
	var gameData = GameData{
		Data: games,
	}
	t, _:= template.ParseFiles("index.html")
	t.Execute(response, gameData)
}
func postComment(response http.ResponseWriter, request * http.Request){
	fmt.Println("Ođeka")
	request.ParseForm()
	for key, value := range request.Form{
		fmt.Printf("%s = %s\n", key, value )
	}
}
func main() {
	openDBConncection()
	defer closeDBConnection()
	router := mux.NewRouter()
	router.HandleFunc("/api", GetApiGames).Methods("GET")
	router.HandleFunc("/", getGames).Methods("GET")
	router.HandleFunc("/postComment", postComment).Methods("POST")
	fileServer := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fileServer))
	http.ListenAndServe(":12345", router)
}