package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
)

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
			queryStr: "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 1, \"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01T00:00:00.000+00:00\", \"contributionCount\": 1}]}]}}}}}",
			want:     Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: "2023-01-01T00:00:00.000+00:00"}}}}}}}},
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

func TestCountCommits(t *testing.T) {
	type args struct {
		query Query
	}

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
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 0,
		},
		{
			name: "countCommitsTodayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsTwo",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 3, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 3,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			want: 2,
		},
		{
			name: "noStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
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
