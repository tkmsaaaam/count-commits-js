package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

func TestCountOverAYear(t *testing.T) {
	type args struct {
		userName string
	}

	today := time.Now().Format("2006-01-02")

	var week = "{\"contributionDays\": ["
	var week3days = "{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1},{\"date\": \"2023-01-01\", \"contributionCount\": 1},{\"date\": \"2023-01-01\", \"contributionCount\": 1}]}"
	date := "{\"date\": \"2023-01-01\", \"contributionCount\": 1}"
	for i := 0; i < 7; i++ {
		week += date
		if i != 6 {
			week += ","
		}
	}
	week += "]}"
	var weeks = "\"weeks\": ["
	weeks += week3days
	weeks += ","
	for i := 0; i < 51; i++ {
		weeks += week
		weeks += ","
	}
	weeks += week3days
	weeks += "]"

	var weeks52 = "\"weeks\": ["
	for i := 0; i < 52; i++ {
		weeks52 += week
		if i != 51 {
			weeks52 += ","
		}
	}
	days365 := "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, " + weeks52 + ",{\"contributionDays\": [{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}"
	weeks52 += "]"

	days363 := "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, " + weeks + "}}}}}"
	days364 := "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, " + weeks52 + "}}}}}"

	tests := []struct {
		name                       string
		args                       args
		queryStr                   string
		wantTodayContributionCount int
		wantCountDays              int
	}{
		{
			name:                       "todayContributionCountIsZeroAndCountDaysIsZero",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 0}]}]}}}}}",
			wantTodayContributionCount: 0,
			wantCountDays:              0,
		},
		{
			name:                       "todayContributionCountIsZeroAndCountDaysIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1}]}]}}}}}",
			wantTodayContributionCount: 0,
			wantCountDays:              1,
		},
		{
			name:                       "todayContributionCountIsOneAndCountDaysIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}",
			wantTodayContributionCount: 1,
			wantCountDays:              1,
		},
		{
			name:                       "todayContributionCountIsZeroAndCountDaysIsTwo",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1},{\"date\": \"2023-01-02\", \"contributionCount\": 1}]}]}}}}}",
			wantTodayContributionCount: 0,
			wantCountDays:              2,
		},
		{
			name:                       "todayContributionCountIsOneAndCountDaysIsTwo",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1},{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}",
			wantTodayContributionCount: 1,
			wantCountDays:              2,
		},
		{
			name:                       "todayContributionCountIsZeroAndOverAYear",
			args:                       args{userName: "octocat"},
			queryStr:                   days364,
			wantTodayContributionCount: 0,
			wantCountDays:              727,
		},
		{
			name:                       "todayContributionCountIsOneAndOverAYear",
			args:                       args{userName: "octocat"},
			queryStr:                   days365,
			wantTodayContributionCount: 1,
			wantCountDays:              728,
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			reqQuery, _ := io.ReadAll(req.Body)
			if strings.Index(string(reqQuery), today) != -1 {
				io.WriteString(w, tt.queryStr)
			} else {
				io.WriteString(w, days363)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			gotTodayContributionCount, gotCountDays := countOverAYear(tt.args.userName, client)
			if gotTodayContributionCount != tt.wantTodayContributionCount {
				t.Errorf("add() = %v, want %v", gotTodayContributionCount, tt.wantTodayContributionCount)
			}
			if gotCountDays != tt.wantCountDays {
				t.Errorf("add() = %v, want %v", gotCountDays, tt.wantCountDays)
			}
		})
	}
}

func TestExecQuery(t *testing.T) {
	type args struct {
		ctx       context.Context
		variables map[string]interface{}
	}

	tests := []struct {
		name     string
		args     args
		queryStr string
		want     Query
	}{
		{
			name:     "queryIsNil",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "{\"data\": {}}",
			want:     Query{},
		},
		{
			name:     "totalContributionsIsZero",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 0}}}}}",
			want:     Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0}}}},
		},
		{
			name:     "totalContributionsIsOne",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1}]}]}}}}}",
			want:     Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: "2023-01-01"}}}}}}}},
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, tt.queryStr)
		})
		t.Run(tt.name, func(t *testing.T) {
			got := Client{client}.execQuery(tt.args.ctx, tt.args.variables)
			if got.User.ContributionsCollection.ContributionCalendar.TotalContributions != tt.want.User.ContributionsCollection.ContributionCalendar.TotalContributions {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

type localRoundTripper struct {
	handler http.Handler
}

func (localRoundTripper localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ressponseRecorder := httptest.NewRecorder()
	localRoundTripper.handler.ServeHTTP(ressponseRecorder, req)
	return ressponseRecorder.Result(), nil
}

func TestIsContinue(t *testing.T) {
	type args struct {
		i                      int
		todayContributionCount int
		streak                 int
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AllArgsAreZero",
			args: args{i: 0, todayContributionCount: 0, streak: 0},
			want: true,
		},
		{
			name: "TodayContributionCountIsZeroStreakIs364",
			args: args{i: 1, todayContributionCount: 0, streak: 364},
			want: true,
		},
		{
			name: "iIsOneTodayContributionCountIsZeroStreakIs729",
			args: args{i: 1, todayContributionCount: 0, streak: 729},
			want: true,
		},
		{
			name: "iIsOneTodayContributionCountIsOneStreakIs365",
			args: args{i: 1, todayContributionCount: 1, streak: 365},
			want: true,
		},
		{
			name: "iIsTwoTodayContributionCountIsOneStreakIs365",
			args: args{i: 2, todayContributionCount: 1, streak: 365},
			want: false,
		},
		{
			name: "iIsOneTodayContributionCountIsOneStreakIs730",
			args: args{i: 1, todayContributionCount: 1, streak: 730},
			want: true,
		},
		{
			name: "iIsOneTodayContributionCountIsOneStreakIs364",
			args: args{i: 1, todayContributionCount: 1, streak: 364},
			want: false,
		},
		{
			name: "iIsOneTodayContributionCountIsZeroStreakIs363",
			args: args{i: 1, todayContributionCount: 0, streak: 363},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isContinue(tt.args.i, tt.args.todayContributionCount, tt.args.streak)
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountCommits(t *testing.T) {
	type args struct {
		query Query
	}

	today := time.Now().Format("2006-01-02")

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "queryIsNil",
			args: args{query: Query{}},
			want: 0,
		},
		{
			name: "weeksLengthIsZero",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0, Weeks: []Week{}}}}}},
			want: 0,
		},
		{
			name: "commitsTodayIsZero",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: today}}}}}}}}},
			want: 0,
		},
		{
			name: "countCommitsTodayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsTwo",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 3, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: today}}}}}}}}},
			want: 3,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "noStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: today}}}}}}}}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countCommittedDays(tt.args.query)
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateMessage(t *testing.T) {
	type args struct {
		countCommitsToday int
		countDays         int
		userName          string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "notCommitedAndNoStreak",
			args: args{countCommitsToday: 0, countDays: 0, userName: "octocat"},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は0\nhttps://github.com/octocat",
		},
		{
			name: "notCommitedAnd",
			args: args{countCommitsToday: 0, countDays: 1, userName: "octocat"},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は1\nhttps://github.com/octocat",
		},
		{
			name: "commitedNoStreak",
			args: args{countCommitsToday: 1, countDays: 0, userName: "octocat"},
			want: "\n今日のコミット数は1\n連続コミット日数は0\nhttps://github.com/octocat",
		},
		{
			name: "commited",
			args: args{countCommitsToday: 1, countDays: 1, userName: "octocat"},
			want: "\n今日のコミット数は1\n連続コミット日数は1\nhttps://github.com/octocat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMessage(tt.args.countCommitsToday, tt.args.countDays, tt.args.userName); got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}
