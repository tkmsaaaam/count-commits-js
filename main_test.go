package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/shurcooL/graphql"
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
			name:     "urlNil",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": graphql.String("octocat")}},
			queryStr: "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"totalContributions\": 0}}}}}",
			want:     Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0}}}},
		},
	}
	client := graphql.NewClient("/graphql", http.DefaultClient)
	for _, tt := range tests {
		http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
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

func TestCountCommits(t *testing.T) {
	type args struct {
		query Query
	}

	tests := []struct {
		name                  string
		args                  args
		wantCountCommitsToday int
		wantCountDays         int
	}{
		{
			name:                  "queryIsNil",
			args:                  args{query: Query{}},
			wantCountCommitsToday: 0,
			wantCountDays:         0,
		},
		{
			name:                  "weeksLengthIsZero",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0, Weeks: []Week{}}}}}},
			wantCountCommitsToday: 0,
			wantCountDays:         0,
		},
		{
			name:                  "commitsTodayIsZero",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 0, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 0,
			wantCountDays:         0,
		},
		{
			name:                  "countCommitsTodayIsOne",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         1,
		},
		{
			name:                  "countCommitsTodayIsTwo",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 2,
			wantCountDays:         1,
		},
		{
			name:                  "countCommitsTodayIsZeroAndYesterdayIsOne",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 0,
			wantCountDays:         1,
		},
		{
			name:                  "countCommitsTodayIsZeroAndStreak",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 0,
			wantCountDays:         2,
		},
		{
			name:                  "countCommitsTodayIsOneAndStreak",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 3, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         3,
		},
		{
			name:                  "countCommitsTodayIsOneAndYesterdayIsOne",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         2,
		},
		{
			name:                  "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         2,
		},
		{
			name:                  "noStreak",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 2, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -2).Format("2006-01-02")}, {ContributionCount: 0, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCountCommtsToday, gotCountDays := countCommits(tt.args.query)
			if gotCountCommtsToday != tt.wantCountCommitsToday {
				t.Errorf("add() = %v, want %v", gotCountCommtsToday, tt.wantCountCommitsToday)
			}
			if gotCountDays != tt.wantCountDays {
				t.Errorf("add() = %v, want %v", gotCountDays, tt.wantCountDays)
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
