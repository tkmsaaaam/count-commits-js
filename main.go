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
	var countCommitsToday int
	var streak = 365
	userName := os.Getenv("GH_USER_NAME")
	for i := 0; i < 10; i++ {
		if streak < 365 {
			break
		}
		from := DateTime{time.Now().AddDate(-(i + 1), 0, 1)}
		to := DateTime{time.Now().AddDate(-i, 0, 0)}
		variables := map[string]interface{}{
			"name": githubv4.String(userName),
			"from": githubv4.DateTime(from),
			"to":   githubv4.DateTime(to),
		}
		query := Client{graphqlClient}.execQuery(context.Background(), variables)
		today, count := countCommits(query)
		if countCommitsToday == 0 {
			countCommitsToday = today
		}
		streak = count
		countDays += streak
	}
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
