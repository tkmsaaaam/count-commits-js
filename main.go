package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"os"
	"strconv"

	"github.com/slack-go/slack"
)

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

type Repositories []struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    Owner `json:"owner"`
	HTMLURL                  string        `json:"html_url"`
	Description              interface{}   `json:"description"`
	Fork                     bool          `json:"fork"`
	URL                      string        `json:"url"`
	ForksURL                 string        `json:"forks_url"`
	KeysURL                  string        `json:"keys_url"`
	CollaboratorsURL         string        `json:"collaborators_url"`
	TeamsURL                 string        `json:"teams_url"`
	HooksURL                 string        `json:"hooks_url"`
	IssueEventsURL           string        `json:"issue_events_url"`
	EventsURL                string        `json:"events_url"`
	AssigneesURL             string        `json:"assignees_url"`
	BranchesURL              string        `json:"branches_url"`
	TagsURL                  string        `json:"tags_url"`
	BlobsURL                 string        `json:"blobs_url"`
	GitTagsURL               string        `json:"git_tags_url"`
	GitRefsURL               string        `json:"git_refs_url"`
	TreesURL                 string        `json:"trees_url"`
	StatusesURL              string        `json:"statuses_url"`
	LanguagesURL             string        `json:"languages_url"`
	StargazersURL            string        `json:"stargazers_url"`
	ContributorsURL          string        `json:"contributors_url"`
	SubscribersURL           string        `json:"subscribers_url"`
	SubscriptionURL          string        `json:"subscription_url"`
	CommitsURL               string        `json:"commits_url"`
	GitCommitsURL            string        `json:"git_commits_url"`
	CommentsURL              string        `json:"comments_url"`
	IssueCommentURL          string        `json:"issue_comment_url"`
	ContentsURL              string        `json:"contents_url"`
	CompareURL               string        `json:"compare_url"`
	MergesURL                string        `json:"merges_url"`
	ArchiveURL               string        `json:"archive_url"`
	DownloadsURL             string        `json:"downloads_url"`
	IssuesURL                string        `json:"issues_url"`
	PullsURL                 string        `json:"pulls_url"`
	MilestonesURL            string        `json:"milestones_url"`
	NotificationsURL         string        `json:"notifications_url"`
	LabelsURL                string        `json:"labels_url"`
	ReleasesURL              string        `json:"releases_url"`
	DeploymentsURL           string        `json:"deployments_url"`
	CreatedAt                time.Time     `json:"created_at"`
	UpdatedAt                time.Time     `json:"updated_at"`
	PushedAt                 time.Time     `json:"pushed_at"`
	GitURL                   string        `json:"git_url"`
	SSHURL                   string        `json:"ssh_url"`
	CloneURL                 string        `json:"clone_url"`
	SvnURL                   string        `json:"svn_url"`
	Homepage                 interface{}   `json:"homepage"`
	Size                     int           `json:"size"`
	StargazersCount          int           `json:"stargazers_count"`
	WatchersCount            int           `json:"watchers_count"`
	Language                 string        `json:"language"`
	HasIssues                bool          `json:"has_issues"`
	HasProjects              bool          `json:"has_projects"`
	HasDownloads             bool          `json:"has_downloads"`
	HasWiki                  bool          `json:"has_wiki"`
	HasPages                 bool          `json:"has_pages"`
	ForksCount               int           `json:"forks_count"`
	MirrorURL                interface{}   `json:"mirror_url"`
	Archived                 bool          `json:"archived"`
	Disabled                 bool          `json:"disabled"`
	OpenIssuesCount          int           `json:"open_issues_count"`
	License                  interface{}   `json:"license"`
	AllowForking             bool          `json:"allow_forking"`
	IsTemplate               bool          `json:"is_template"`
	WebCommitSignoffRequired bool          `json:"web_commit_signoff_required"`
	Topics                   []interface{} `json:"topics"`
	Visibility               string        `json:"visibility"`
	Forks                    int           `json:"forks"`
	OpenIssues               int           `json:"open_issues"`
	Watchers                 int           `json:"watchers"`
	DefaultBranch            string        `json:"default_branch"`
}

type CommitAuthor struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type CommitCommitter struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type Tree struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

type Verification struct {
	Verified  bool        `json:"verified"`
	Reason    string      `json:"reason"`
	Signature interface{} `json:"signature"`
	Payload   interface{} `json:"payload"`
}

type Commit struct {
	Author CommitAuthor `json:"author"`
	Committer CommitCommitter `json:"committer"`
	Message string `json:"message"`
	Tree    Tree `json:"tree"`
	URL          string `json:"url"`
	CommentCount int    `json:"comment_count"`
	Verification Verification `json:"verification"`
}

type Author      struct {
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

type Committer struct {
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

type Parents []struct {
	Sha     string `json:"sha"`
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}

type Commits []struct {
	Sha    string `json:"sha"`
	NodeID string `json:"node_id"`
	Commit Commit `json:"commit"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
	Author      Author `json:"author"`
	Committer Committer `json:"committer"`
	Parents Parents `json:"parents"`
}

func requestApi(url string) ([]byte, error){
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error Request:", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error Response:", resp.Status)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ioutil.ReadAll err=%s", err.Error())
	}

	return body, nil
}

func postSlack(counts int, userName string) {
	tkn := os.Args[2]
	c := slack.New(tkn)

	var message string

	if counts == 0 {
		message = "<!channel>今日はまだコミットしていません！"
	} else {
		message = "今日のコミット数は" + strconv.Itoa(counts)
	}

	message += "\nhttps://github.com/" + userName

	_, _, err := c.PostMessage(os.Args[3], slack.MsgOptionText(message, true))
	if err != nil {
		panic(err)
	}
}

func main() {
	userName := os.Args[1]
	url := "https://api.github.com/users/" + userName + "/repos"

	body, err := requestApi(url)

	if err != nil {
		fmt.Println("Error Request API")
	}

	var repositories Repositories

	if err = json.Unmarshal(body, &repositories); err != nil {
		log.Fatalf("json.Unmarshal err=%s", err.Error())
	}
	
	counts := 0
	now := time.Now()

	for _, repo := range repositories {
		name := repo.Name

		url := "https://api.github.com/repos/" +userName + "/" + name + "/commits"

		body, err := requestApi(url)

		if err != nil {
			fmt.Println("Error Request API")
		}

		var commits Commits
	
		if err = json.Unmarshal(body, &commits); err != nil {
			log.Fatalf("json.Unmarshal err=%s", err.Error())
		}

		for j := range commits {
			date := commits[j].Commit.Author.Date
			if (date.Year() == now.Year() &&  date.Month() == now.Month() && date.Day() == now.Day()) {
				counts += 1
			} else {
				break
			}
		}
	}
	postSlack(counts, userName)
}
