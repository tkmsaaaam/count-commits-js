package main

import (
	"context"
	"fmt"
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
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Args[4]},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	userName := os.Args[1]
	variables := map[string]interface{}{
		"name": graphql.String(userName),
	}
	graphqlErr := graphqlClient.Query(context.Background(), &query, variables)
	if graphqlErr != nil {
		fmt.Println(graphqlErr)
	}
	weeksLen := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	var countDays = 0
	var countCommitsToday int
	now := time.Now()
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

	var message string
	if countCommitsToday == 0 {
		message = "<!channel> 今日はまだコミットしていません！"
	} else {
		message = "\n今日のコミット数は" + fmt.Sprint(countCommitsToday)
		countDays++
	}

	message += "\n---" + "\n連続コミット日数は" + fmt.Sprint(countDays) + "\nhttps://github.com/" + userName

	postSlack(message)
}
