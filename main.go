/*https://tutorialedge.net/golang/creating-restful-api-with-golang/#building-our-router*/
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/all", returnAllArticles)
	myRouter.HandleFunc("/articles", returnAllArticles)
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")
	myRouter.HandleFunc("/article/{id}", returnSingleArticle)
	myRouter.HandleFunc("/api/{git_name}/{git_repo}", returnRepo)
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	fmt.Println("Rest API v2.0 - Mux Routers")
	Articles = []Article{
		Article{ID: "1", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{ID: "2", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}
	handleRequests()
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	for _, article := range Articles {
		if article.ID == key {
			json.NewEncoder(w).Encode(article)
		}
	}
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	Articles = append(Articles, article)
	json.NewEncoder(w).Encode(article)
}

func getRepoData(user string, repo string, contents chan []FileMetadata) {
	url := "https://api.github.com/repos/" + user + "/" + repo + "/contents/"
	response, err := http.Get(url)
	fmt.Println(url)
	fmt.Println(response.StatusCode)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(responseData))
	var repoData []FileMetadata
	json.Unmarshal(responseData, &repoData)
	// data appears to not be loading in time, returns {[]}
	fmt.Println(repoData)

	contents <- repoData
}

func returnRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contents := make(chan []FileMetadata)
	go getRepoData(vars["git_name"], vars["git_repo"], contents)
	Contents = <-contents

	json.NewEncoder(w).Encode(Contents)
}

func decodeString(encodedString string) string {
	decodedString, _ := base64.StdEncoding.DecodeString(encodedString)
	fmt.Printf(string(decodedString))
	return string(decodedString)
}

// Article ...
type Article struct {
	ID      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

// type RepoContents struct {
// 	Files []File
// }

// FileMetadata ...
type FileMetadata struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Links       Links  `json:"_links"`
}

// File ...
type File struct {
	FileMetadata
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// Links ...
type Links struct {
	Self string `json:"self"`
	Git  string `json:"git"`
	HTML string `json:"html"`
}

// Articles ...
var Articles []Article

// Contents ...
var Contents []FileMetadata
