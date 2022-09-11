package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v45/github"
	"github.com/slack-go/slack"
	"os"
	"strconv"
	"time"
)

func postSlack(counts int, userName string) {
	c := slack.New(os.Args[2])

	var message string

	if counts == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "今日のコミット数は" + strconv.Itoa(counts)
	}

	message += "\nhttps://github.com/" + userName

	_, _, err := c.PostMessage(os.Args[3], slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func main() {
	userName := os.Args[1]

	client := github.NewClient(nil)

	opt := &github.RepositoryListOptions{Type: "public"}
	repos, _, err := client.Repositories.List(context.Background(), userName, opt)

	if err != nil {
		fmt.Println("Error can't get repositories")
		fmt.Println(err)
		panic(err)
	}

	counts := 0
	now := time.Now()

	for _, repo := range repos {
		repoName := *repo.Name

		opt := &github.CommitsListOptions{}
		repositoryCommits, _, err := client.Repositories.ListCommits(context.Background(), userName, repoName, opt)

		if err != nil {
			fmt.Println("Error can't get commits")
			fmt.Println(err)
			panic(err)
		}

		for _, repositoryCommit := range repositoryCommits {
			date := repositoryCommit.Commit.Author.Date
			if date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day() {
				counts += 1
			} else {
				break
			}
		}
	}
	postSlack(counts, userName)
}
