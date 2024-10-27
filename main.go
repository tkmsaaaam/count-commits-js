package main

import (
	"context"
	"fmt"
	"log"
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

type SlackClient struct {
	*slack.Client
}

func main() {
	userName := os.Getenv("GH_USER_NAME")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)

	todayContributionCount, countDays, total := countOverAYear(userName, graphqlClient)

	message := createMessage(todayContributionCount, countDays, total, userName)
	SlackClient{slack.New(os.Getenv("SLACK_BOT_TOKEN"))}.postSlack(message)
}

var offset = 0
var lastCount int

func countOverAYear(userName string, graphqlClient *githubv4.Client) (int, int, int) {
	var countDays int
	var todayContributionCount int
	var total int
	for i := 0; isContinue(i); i++ {
		const daysLength = 365
		from := githubv4.DateTime{Time: time.Now().AddDate(0, 0, offset-daysLength)}
		to := githubv4.DateTime{Time: time.Now().AddDate(0, 0, offset)}
		variables := map[string]interface{}{
			"name": githubv4.String(userName),
			"from": githubv4.DateTime(from),
			"to":   githubv4.DateTime(to),
		}
		offset = offset - daysLength
		query := Client{graphqlClient}.execQuery(context.Background(), variables)
		if i == 0 {
			w := len(query.User.ContributionsCollection.ContributionCalendar.Weeks) - 1
			c := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[w].ContributionDays) - 1
			today := query.User.ContributionsCollection.ContributionCalendar.Weeks[w].ContributionDays[c]
			if today.Date == time.Now().Format("2006-01-02") {
				todayContributionCount = today.ContributionCount
			}
		}
		count, sum := countCommittedDays(query)
		countDays += count
		total += sum
	}

	return todayContributionCount, countDays, total
}

func isContinue(i int) bool {
	if i == 0 {
		return true
	}
	if lastCount == 0 {
		return false
	}
	return true
}

func countCommittedDays(query Query) (int, int) {
	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	now := time.Now()
	var countDays int
	var sumCommits int
	for i := weeksLength - 1; i >= 0; i-- {
		daysLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLength - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			lastCount = day.ContributionCount
			if day.ContributionCount == 0 {
				d, _ := time.Parse("2006-01-02", day.Date)
				if now.Format("2006-01-02") == day.Date {
					continue
				} else if d.After(now) {
					continue
				} else {
					return countDays, sumCommits
				}
			}
			sumCommits += day.ContributionCount
			countDays++
		}
	}
	return countDays, sumCommits
}

func (client Client) execQuery(ctx context.Context, variables map[string]interface{}) Query {
	var query Query
	graphqlErr := client.Query(ctx, &query, variables)
	if graphqlErr != nil {
		log.Println("query is error.", graphqlErr)
	}
	return query
}

func createMessage(todayContributionCount, countDays, total int, userName string) string {
	var message string
	if todayContributionCount == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = fmt.Sprintf("\n今日のコミット数は%d", todayContributionCount)
	}
	var average float64
	if countDays == 0 {
		average = 0
	} else {
		average = float64(total) / float64(countDays)
	}
	message += fmt.Sprintf("\n連続コミット日数は%d\n合計コミット数は%d\n平均コミット数は%f\nhttps://github.com/%s", countDays, total, average, userName)
	return message
}

func (client SlackClient) postSlack(message string) {
	_, _, err := client.PostMessage(os.Getenv("SLACK_CHANNEL_ID"), slack.MsgOptionText(message, false))
	if err != nil {
		log.Println("can not post message.", err)
	}
}
