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

type Result struct {
	userName               string
	today                  time.Time
	todayContributionCount int
	start                  time.Time
	latestDay              time.Time
	total                  int
	streak                 int
	isContinue             bool
}

func main() {
	userName := os.Getenv("GH_USER_NAME")
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)

	now := time.Now()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	result := Result{userName: userName, todayContributionCount: 0, today: today, start: today, latestDay: today, total: 0, streak: 0, isContinue: true}

	slackClient := SlackClient{slack.New(os.Getenv("SLACK_BOT_TOKEN"))}
	if err := result.countOverAYear(userName, graphqlClient); err != nil {
		slackClient.postSlackError()
		return
	}

	message := result.createMessage()
	slackClient.postSlack(message)
}

func (r *Result) countOverAYear(userName string, graphqlClient *githubv4.Client) error {
	for i := 0; r.isContinue; i++ {
		from := githubv4.DateTime{Time: r.latestDay.AddDate(0, 0, -366)}
		to := githubv4.DateTime{Time: r.latestDay.AddDate(0, 0, -1)}
		variables := map[string]interface{}{
			"name": githubv4.String(userName),
			"from": githubv4.DateTime(from),
			"to":   githubv4.DateTime(to),
		}
		query := Client{graphqlClient}.execQuery(context.Background(), variables)
		if err := r.countCommittedDays(query); err != nil {
			return err
		}
	}
	return nil
}

func (r *Result) countCommittedDays(query Query) error {
	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	for i := weeksLength - 1; i >= 0; i-- {
		daysLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays)
		for j := daysLength - 1; j >= 0; j-- {
			day := query.User.ContributionsCollection.ContributionCalendar.Weeks[i].ContributionDays[j]
			d, _ := time.Parse("2006-01-02", day.Date)
			if d.Equal(r.today) {
				r.todayContributionCount = day.ContributionCount
			}
			if d.After(r.today) || d.After(r.latestDay) || d.Equal(r.latestDay) {
				continue
			}
			if !d.Equal(r.latestDay.AddDate(0, 0, -1)) {
				return fmt.Errorf("is not consecutive")
			}
			if day.ContributionCount == 0 {
				if d.Equal(r.today) {
					continue
				} else {
					r.isContinue = false
					return nil
				}
			}
			r.latestDay = d
			r.total += day.ContributionCount
			r.streak++
		}
	}
	return nil
}

func (client Client) execQuery(ctx context.Context, variables map[string]interface{}) Query {
	var query Query
	graphqlErr := client.Query(ctx, &query, variables)
	if graphqlErr != nil {
		log.Println("query is error.", graphqlErr)
	}
	return query
}

func (r *Result) createMessage() string {
	var message string
	if r.todayContributionCount == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = fmt.Sprintf("\n今日のコミット数は%d", r.todayContributionCount)
	}
	var average float64
	if r.streak == 0 {
		average = 0
	} else {
		average = float64(r.total) / float64(r.streak)
	}
	message += fmt.Sprintf("\n連続コミット日数は%d\n合計コミット数は%d\n平均コミット数は%f\n期間は%s ~\nhttps://github.com/%s", r.streak, r.total, average, r.latestDay.Format("2006-01-02"), r.userName)
	return message
}

func (client SlackClient) postSlack(message string) {
	_, _, err := client.PostMessage(os.Getenv("SLACK_CHANNEL_ID"), slack.MsgOptionText(message, false))
	if err != nil {
		log.Println("can not post message.", err)
	}
}

func (client SlackClient) postSlackError() {
	_, _, err := client.PostMessage(os.Getenv("SLACK_CHANNEL_ID"), slack.MsgOptionText("<!channel> count-commits-js error", false))
	if err != nil {
		log.Println("can not post message.", err)
	}
}
