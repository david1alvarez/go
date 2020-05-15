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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/{git_name}/{git_repo}", returnRepo)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	// getFileDataFromLocalFile()
	handleRequests()
}

func getFileDataFromLocalFile() []File {
	fmt.Println("getting data from local file...")
	data, err := ioutil.ReadFile("./google-uuid.contents.json")
	if err != nil {
		fmt.Print(err.Error())
	}

	var fileData []File
	err = json.Unmarshal(data, &fileData)
	if err != nil {
		fmt.Print("error: ", err)
	}

	return fileData
}

func getFileData(user string, repo string, filePath string, contents chan File) {
	url := "https://api.github.com/repos/" + user + "/" + repo + "/contents/" + filePath
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var fileData File
	json.Unmarshal(responseData, &fileData)

	contents <- fileData
}

func getRepoData(user string, repo string, contents chan []FileMetadata) {
	url := "https://api.github.com/repos/" + user + "/" + repo + "/contents/"
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var repoData []FileMetadata
	json.Unmarshal(responseData, &repoData)

	// this would need to be updated for repos dealing with folders
	// this will currently only get data for top level files

	files := make(chan File, len(repoData))
	for _, file := range repoData {
		if file.Type == "file" {
			go getFileData(user, repo, file.Path, files)
			Files = append(Files, <-files)
		}
	}

	contents <- repoData
}

func returnRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contents := make(chan []FileMetadata)
	go getRepoData(vars["git_name"], vars["git_repo"], contents)
	Contents = <-contents

	if Contents == nil {
		fmt.Println("contents empty, Github's rate limiter likely reached. Pulling data from saved file")
		Files = getFileDataFromLocalFile()
	}

	for i := range Files {
		Files[i].Content = decodeString(Files[i].Content)
	}
	json.NewEncoder(w).Encode(Files)
}

func decodeString(encodedString string) string {
	decodedString, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		fmt.Println(err)
	}
	return string(decodedString)
}

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

// Contents ...
var Contents []FileMetadata

// Files ...
var Files []File
