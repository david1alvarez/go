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
	myRouter.HandleFunc("/api/{git_name}/{git_repo}/info", returnRepoMeta)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
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

func getRepoMetaData(user string, repo string, channel chan RepoMetadata) {
	url := "https://api.github.com/repos/" + user + "/" + repo
	response, err := http.Get(url)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var repoMetadata RepoMetadata
	json.Unmarshal(responseData, &repoMetadata)
	channel <- repoMetadata
}

func returnRepoMeta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metaData := make(chan RepoMetadata)
	go getRepoMetaData(vars["git_name"], vars["git_repo"], metaData)
	repoMetadata := <-metaData

	var starsAndFiles StarsAndFiles
	starsAndFiles.Stars = repoMetadata.StargazersCount

	fileData := make(chan []FileMetadata)
	go getRepoData(vars["git_name"], vars["git_repo"], fileData)
	files := <-fileData

	for i, file := range files {
		if file.Type == "file" {
			starsAndFiles.Files = append(starsAndFiles.Files, files[i].Name)
		}
	}

	json.NewEncoder(w).Encode(starsAndFiles)
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

// StarsAndFiles ...
type StarsAndFiles struct {
	Stars int
	Files []string
}

// RepoMetadata ...
type RepoMetadata struct {
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	Name             string `json:"name"`
	FullName         string `json:"full_name"`
	Private          bool   `json:"private"`
	Owner            Owner  `json:"owner"`
	HTMLURL          string `json:"html_url"`
	Description      string `json:"description"`
	Fork             bool   `json:"fork"`
	URL              string `json:"url"`
	ForksURL         string `json:"forks_url"`
	KeysURL          string `json:"keys_url"`
	CollaboratorsURL string `json:"collaborators_url"`
	TeamsURL         string `json:"teams_url"`
	HooksURL         string `json:"hooks_url"`
	IssueEventsURL   string `json:"issue_events_url"`
	EventsURL        string `json:"events_url"`
	AssigneesURL     string `json:"assignees_url"`
	BranchesURL      string `json:"branches_url"`
	TagsURL          string `json:"tags_url"`
	BlobsURL         string `json:"blobs_url"`
	GitTagsURL       string `json:"git_tags_url"`
	GitRefsURL       string `json:"git_refs_url"`
	TreesURL         string `json:"trees_url"`
	StatusesURL      string `json:"statuses_url"`
	LanguagesURL     string `json:"languages_url"`
	StargazersURL    string `json:"stargazers_url"`
	ContributorsURL  string `json:"contributors_url"`
	SubscribersURL   string `json:"subscribers_url"`
	SubscriptionURL  string `json:"subscription_url"`
	CommitsURL       string `json:"commits_url"`
	GitCommitsURL    string `json:"git_commits_url"`
	CommentsURL      string `json:"comments_url"`
	IssueCommentURL  string `json:"issue_comment_url"`
	ContentsURL      string `json:"contents_url"`
	CompareURL       string `json:"compare_url"`
	MergesURL        string `json:"merges_url"`
	ArchiveURL       string `json:"archive_url"`
	DownloadsURL     string `json:"downloads_url"`
	IssuesURL        string `json:"issues_url"`
	PullsURL         string `json:"pulls_url"`
	MilestonesURL    string `json:"milestones_url"`
	NotificationsURL string `json:"notifications_url"`
	LabelsURL        string `json:"labels_url"`
	ReleasesURL      string `json:"releases_url"`
	DeploymentsURL   string `json:"deployments_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	PushedAt         string `json:"pushed_at"`
	GitURL           string `json:"git_url"`
	SSHURL           string `json:"ssh_url"`
	CloneURL         string `json:"clone_url"`
	SvnURL           string `json:"svn_url"`
	Homepage         string `json:"homepage"`
	Size             int    `json:"size"`
	StargazersCount  int    `json:"stargazers_count"`
	WatchersCount    int    `json:"watchers_count"`
	Language         string `json:"language"`
	HasIssues        bool   `json:"has_issues"`
	HasProjects      bool   `json:"has_projects"`
	HasDownloads     bool   `json:"has_downloads"`
	HasWiki          bool   `json:"has_wiki"`
	HasPages         bool   `json:"has_pages"`
	ForksCount       int    `json:"forks_count"`
	MirrorURL        string `json:"mirror_url"`
	Archived         bool   `json:"archived"`
	Disabled         bool   `json:"disabled"`
	OpenIssuesCount  int    `json:"open_issues_count"`
	License          string `json:"license"`
	Forks            int    `json:"forks"`
	OpenIssues       int    `json:"open_issues"`
	Watchers         int    `json:"watchers"`
	DefaultBranch    string `json:"default_branch"`
	TempCloneToken   string `json:"temp_clone_token"`
	NetworkCount     int    `json:"network_count"`
	SubscribersCount int    `json:"subscribers_count"`
}

// Owner ...
type Owner struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// Contents ...
var Contents []FileMetadata

// Files ...
var Files []File
