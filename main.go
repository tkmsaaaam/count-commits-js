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
	_, _, err := slack.New(os.Getenv("TOKEN_FOR_BOT")).PostMessage(os.Getenv("CHANNEL_ID"), slack.MsgOptionText(message, false))
	if err != nil {
		fmt.Println(err)
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
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)
	userName := os.Getenv("USER_NAME")
	variables := map[string]interface{}{
		"name": graphql.String(userName),
	}
	graphqlErr := graphqlClient.Query(context.Background(), &query, variables)
	if graphqlErr != nil {
		fmt.Println(graphqlErr)
	}

	weeksLength := len(query.User.ContributionsCollection.ContributionCalendar.Weeks)
	var countCommitsToday int
	now := time.Now()
	var countDays int
out:
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
				break out
			} else {
				countDays++
			}
		}
	}

	message := createMessage(countCommitsToday, countDays, userName)
	postSlack(message)
}
