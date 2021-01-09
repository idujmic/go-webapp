package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
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
	//fmt.Println(res)
	//fmt.Println(string(body))

	//json.NewEncoder(response).Encode(f)
	collection := client.Database("ivandb").Collection("games")

	for _,game :=range f.Data{
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
func main() {
	openDBConncection()
	router := mux.NewRouter()
	router.HandleFunc("/api", GetApiGames).Methods("GET")
	router.HandleFunc("/", getGames).Methods("GET")
	fileServer := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", fileServer))
	http.ListenAndServe(":12345", router)
}