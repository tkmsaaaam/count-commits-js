package main

import (
	"context"
	"fmt"
	"github.com/shurcooL/githubv4"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
	"os"
	"time"
)

type ContributionDay struct {
	ContributionCount int
	Date              string
}

type Week struct {
	ContributionDays []ContributionDay
}

type ContributionCalendar struct {
	TotalContributions githubv4.Int
	Weeks              []Week
}

type ContributionsCollection struct {
	ContributionCalendar ContributionCalendar
}

type User struct {
	ContributionsCollection ContributionsCollection
}

type Query struct {
	User User `graphql:"user(login: $name)"`
}

type Client struct {
	*githubv4.Client
}

func main() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)
	userName := os.Getenv("GH_USER_NAME")
	variables := map[string]interface{}{
		"name": githubv4.String(userName),
	}
	query := Client{graphqlClient}.execQuery(context.Background(), variables)
	countCommitsToday, countDays := countCommits(query)

	message := createMessage(countCommitsToday, countDays, userName)
	postSlack(message)
}

func countCommits(query Query) (int, int) {
	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	var countCommitsToday int
	now := time.Now()
	var countDays int
	for i := weeksLength - 1; i >= 0; i-- {
		daysLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLength - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			if now.Format("2006-01-02") == day.Date {
				countCommitsToday = day.ContributionCount
				if countCommitsToday != 0 {
					countDays++
				}
				continue
			}
			if day.ContributionCount == 0 {
				return countCommitsToday, countDays
			} else {
				countDays++
			}
		}
	}
	return countCommitsToday, countDays
}

func (client Client) execQuery(ctx context.Context, variables map[string]interface{}) Query {
	var query Query
	graphqlErr := client.Query(ctx, &query, variables)
	if graphqlErr != nil {
		fmt.Println(graphqlErr)
	}
	return query
}

func createMessage(countCommitsToday int, countDays int, userName string) string {
	var message string
	if countCommitsToday == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "\n今日のコミット数は" + fmt.Sprint(countCommitsToday)
	}
	message += "\n連続コミット日数は" + fmt.Sprint(countDays) + "\nhttps://github.com/" + userName
	return message
}

func postSlack(message string) {
	_, _, err := slack.New(os.Getenv("SLACK_BOT_TOKEN")).PostMessage(os.Getenv("SLACK_CHANNEL_ID"), slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Println(err)
	}
}
