package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

type ContributionDay struct {
	ContributionCount int
	Date              string
}

type Week struct {
	ContributionDays []ContributionDay
}

type ContributionCalendar struct {
	Weeks []Week
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
	userName := os.Getenv("GH_USER_NAME")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)

	todayContributionCount, countDays := countOverAYear(userName, graphqlClient)

	message := createMessage(todayContributionCount, countDays, userName)
	postSlack(message)
}

func countOverAYear(userName string, graphqlClient *githubv4.Client) (int, int) {
	var countDays int
	var todayContributionCount int
	var streak int
	for i := 0; isContinue(i, todayContributionCount, streak); i++ {
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
		count := countCommittedDays(query)
		streak = count
		countDays += streak
	}

	return todayContributionCount, countDays
}

func isContinue(i int, todayContributionCount int, streak int) bool {
	if i == 0 {
		return true
	}
	j := streak % 365
	if todayContributionCount == 0 && j == 365-1 && streak >= (i*365)-1 {
		return true
	}
	if todayContributionCount > 0 && j == 0 && streak >= i*365 {
		return true
	}
	return false
}

func countCommittedDays(query Query) int {
	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	now := time.Now()
	var countDays int
	for i := weeksLength - 1; i >= 0; i-- {
		daysLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLength - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			if day.ContributionCount == 0 {
				if now.Format("2006-01-02") == day.Date {
					continue
				} else {
					return countDays
				}
			}
			countDays++
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
