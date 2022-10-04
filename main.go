package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v45/github"
	"github.com/slack-go/slack"
	"os"
	"time"
)

func postSlack(counts int, userName string) {
	c := slack.New(os.Args[2])

	var message string

	if counts == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "今日のコミット数は" + fmt.Sprint(counts)
	}

	message += "\nhttps://github.com/" + userName

	_, _, err := c.PostMessage(os.Args[3], slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	userName := os.Args[1]

	client := github.NewClient(nil)
	ctx := context.Background()

	repos, _, err := client.Repositories.List(ctx, userName, &github.RepositoryListOptions{Type: "public"})
	if err != nil {
		fmt.Println("Error can't get repositories")
		fmt.Println(err)
		return
	}

	counts := 0
	now := time.Now()

	for _, repo := range repos {
		repoName := *repo.Name

		repositoryCommits, _, err := client.Repositories.ListCommits(ctx, userName, repoName, &github.CommitsListOptions{})
		if err != nil {
			fmt.Println("Error can't get commits")
			fmt.Println(err)
			continue
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
