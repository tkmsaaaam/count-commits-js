package main

import (
	"testing"
	"time"
)

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
			name:                  "countCommitsTodayIsOneAndYesterdayIsOne",
			args:                  args{query: Query{User{ContributionsCollection{ContributionCalendar{TotalContributions: 1, Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, {ContributionCount: 1, Date: time.Now().Format("2006-01-02")}}}}}}}}},
			wantCountCommitsToday: 1,
			wantCountDays:         2,
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
