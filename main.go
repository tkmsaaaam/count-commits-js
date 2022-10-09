package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v45/github"
	"github.com/shurcooL/graphql"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
	"os"
	"time"
)

func postSlack(message string) {
	c := slack.New(os.Args[2])

	_, _, err := c.PostMessage(os.Args[3], slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Println(err)
		return
	}
}

var query struct {
	User struct {
		ContributionsCollection struct {
			ContributionCalendar struct {
				TotalContributions graphql.Int
				Weeks              []struct {
					ContributionDays []struct {
						ContributionCount int
						Date              string
					}
				}
			}
		}
	} `graphql:"user(login: $name)"`
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
			if date.Format("2006-01-02") == now.Format("2006-01-02") {
				counts += 1
			} else {
				break
			}
		}
	}

	var message string

	if counts == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "今日のコミット数は" + fmt.Sprint(counts)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Args[4]},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	variables := map[string]interface{}{
		"name": graphql.String(userName),
	}
	graphqlErr := graphqlClient.Query(context.Background(), &query, variables)
	if graphqlErr != nil {
		fmt.Println(err)
	}
	weeksLen := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	var countDays = 0
	var countCommitsToday int
out:
	for i := weeksLen - 1; i >= 0; i-- {
		daysLen := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLen - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			if now.Format("2006-01-02") == day.Date {
				countCommitsToday = day.ContributionCount
				continue
			}
			if day.ContributionCount == 0 {
				break out
			} else {
				countDays++
			}
		}
	}

	if countCommitsToday != 0 {
		countDays++
	}

	message += "\n---" + "\n連続コミット日数は" + fmt.Sprint(countDays) + "\n今日のコミット数は" + fmt.Sprint(countCommitsToday) + "\n---"

	message += "\nhttps://github.com/" + userName

	postSlack(message)
}
