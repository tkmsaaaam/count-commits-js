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
	ContributionsCollection ContributionsCollection `graphql:"contributionsCollection(from: $from to: $to)"`
}

type Query struct {
	User User `graphql:"user(login: $name)"`
}

type Client struct {
	*githubv4.Client
}

type DateTime struct {
	time.Time
}

func main() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)

	countOverAYear(graphqlClient)
}

func countOverAYear(graphqlClient *githubv4.Client) {
	var countDays int
	var todayContributionCount int
	var streak int
	userName := os.Getenv("GH_USER_NAME")
	for i := 0; i == 0 || streak >= 364; i++ {
		from := DateTime{time.Now().AddDate(-(i + 1), 0, 1)}
		to := DateTime{time.Now().AddDate(-i, 0, 0)}
		variables := map[string]interface{}{
			"name": githubv4.String(userName),
			"from": githubv4.DateTime(from),
			"to":   githubv4.DateTime(to),
		}
		query := Client{graphqlClient}.execQuery(context.Background(), variables)
		if i == 0 {
			w := len(query.User.ContributionsCollection.ContributionCalendar.Weeks) - 1
			c := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[w].ContributionDays) - 1
			today := query.User.ContributionsCollection.ContributionCalendar.Weeks[w].ContributionDays[c]
			if today.Date == time.Now().Format("2006-01-02") {
				todayContributionCount = today.ContributionCount
			}
		}
		count := countCommits(query)
		streak = count
		countDays += streak
	}
	message := createMessage(todayContributionCount, countDays, userName)
	postSlack(message)
}

func countCommits(query Query) int {
	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	now := time.Now()
	var countDays int
	for i := weeksLength - 1; i >= 0; i-- {
		daysLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLength - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			if day.ContributionCount == 0 {
				if now.Format("2006-01-02") != day.Date {
					return countDays
				}
			} else {
				countDays++
			}
		}
	}
	return countDays
}

func (client Client) execQuery(ctx context.Context, variables map[string]interface{}) Query {
	var query Query
	graphqlErr := client.Query(ctx, &query, variables)
	if graphqlErr != nil {
		fmt.Println(graphqlErr)
	}
	return query
}

func createMessage(todayContributionCount int, countDays int, userName string) string {
	var message string
	if todayContributionCount == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "\n今日のコミット数は" + fmt.Sprint(todayContributionCount)
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
