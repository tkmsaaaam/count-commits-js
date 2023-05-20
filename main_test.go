package main

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

//go:embed testdata/CountOverAYear/todayContributionCountIsZeroAndCountDaysIsZero.json
var todayContributionCountIsZeroAndCountDaysIsZeroJson string

//go:embed testdata/CountOverAYear/todayContributionCountIsZeroAndCountDaysIsOne.json
var todayContributionCountIsZeroAndCountDaysIsOneJson string

//go:embed testdata/CountOverAYear/todayContributionCountIsZeroAndCountDaysIsTwo.json
var todayContributionCountIsZeroAndCountDaysIsTwoJson string

//go:embed testdata/CountOverAYear/days363.json
var days363 string

//go:embed testdata/CountOverAYear/days364.json
var days364 string

//go:embed testdata/CountOverAYear/weeks52.json
var weeks52 string

func TestCountOverAYear(t *testing.T) {
	type args struct {
		userName string
	}

	today := time.Now().Format("2006-01-02")

	days365 := "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"weeks\": [" + weeks52 + ",{\"contributionDays\": [{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}"

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
			queryStr:                   todayContributionCountIsZeroAndCountDaysIsZeroJson,
			wantTodayContributionCount: 0,
			wantCountDays:              0,
		},
		{
			name:                       "todayContributionCountIsZeroAndCountDaysIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayContributionCountIsZeroAndCountDaysIsOneJson,
			wantTodayContributionCount: 0,
			wantCountDays:              1,
		},
		{
			name:                       "todayContributionCountIsOneAndCountDaysIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"weeks\": [{\"contributionDays\": [{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}",
			wantTodayContributionCount: 1,
			wantCountDays:              1,
		},
		{
			name:                       "todayContributionCountIsZeroAndCountDaysIsTwo",
			args:                       args{userName: "octocat"},
			queryStr:                   todayContributionCountIsZeroAndCountDaysIsTwoJson,
			wantTodayContributionCount: 0,
			wantCountDays:              2,
		},
		{
			name:                       "todayContributionCountIsOneAndCountDaysIsTwo",
			args:                       args{userName: "octocat"},
			queryStr:                   "{\"data\": {\"user\": {\"contributionsCollection\": {\"contributionCalendar\": {\"weeks\": [{\"contributionDays\": [{\"date\": \"2023-01-01\", \"contributionCount\": 1},{\"date\": \"" + today + "\", \"contributionCount\": 1}]}]}}}}}",
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

//go:embed testdata/ExecQuery/queryIsNil.json
var queryIsNilJson string

//go:embed testdata/ExecQuery/totalContributionsIsZero.json
var totalContributionsIsZeroJson string

//go:embed testdata/ExecQuery/totalContributionsIsOne.json
var totalContributionsIsOneJson string

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
			queryStr: queryIsNilJson,
			want:     Query{},
		},
		{
			name:     "totalContributionsIsZero",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: totalContributionsIsZeroJson,
			want:     Query{User{ContributionsCollection{ContributionCalendar{}}}},
		},
		{
			name:     "totalContributionsIsOne",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: totalContributionsIsOneJson,
			want:     Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: "2023-01-01"}}}}}}}},
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
			if len(got.User.ContributionsCollection.ContributionCalendar.Weeks) != len(tt.want.User.ContributionsCollection.ContributionCalendar.Weeks) {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecQueryError(t *testing.T) {
	type args struct {
		ctx       context.Context
		variables map[string]interface{}
	}

	tests := []struct {
		name     string
		args     args
		queryStr string
		want     string
	}{
		{
			name:     "execQueryIsOk",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: queryIsNilJson,
			want:     "",
		},
		{
			name:     "execQueryIsError",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: queryIsNilJson,
			want:     "non-200 OK status code: 500 Internal Server Error body: \"Internal Server Error\"",
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, _ *http.Request) {
			if tt.name == "execQueryIsOk" {
				io.WriteString(w, tt.queryStr)
			} else {
				w.WriteHeader(500)
				io.WriteString(w, "Internal Server Error")
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			orgStdout := os.Stdout
			defer func() {
				os.Stdout = orgStdout
			}()
			r, w, _ := os.Pipe()
			os.Stdout = w
			Client{client}.execQuery(tt.args.ctx, tt.args.variables)
			w.Close()
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("failed to read buf: %v", err)
			}
			gotPrint := strings.TrimRight(buf.String(), "\n")
			if gotPrint != tt.want {
				t.Errorf("add() = %v, want %v", gotPrint, tt.want)
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
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

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
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{}}}}}},
			want: 0,
		},
		{
			name: "commitsTodayIsZero",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: today}}}}}}}}},
			want: 0,
		},
		{
			name: "countCommitsTodayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsTwo",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}},
			want: 1,
		},
		{
			name: "countCommitsTodayIsZeroAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
			want: 3,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOne",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want: 2,
		},
		{
			name: "noStreak",
			args: args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 0, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
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

//go:embed testdata/slack/ok.json
var postMessageIsOk []byte

//go:embed testdata/slack/error.json
var postMessageIsError []byte

func TestPostSlack(t *testing.T) {
	tests := []struct {
		name   string
		apiRes []byte
		want   string
	}{
		{
			name:   "iSOk",
			apiRes: postMessageIsOk,
			want:   "",
		},
		{
			name:   "isError",
			apiRes: postMessageIsError,
			want:   "too_many_attachments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := slacktest.NewTestServer(func(c slacktest.Customize) {
				c.Handle("/chat.postMessage", func(w http.ResponseWriter, _ *http.Request) {
					w.Write(tt.apiRes)
				})
			})
			ts.Start()
			client := slack.New("testToken", slack.OptionAPIURL(ts.GetAPIURL()))

			t.Helper()

			orgStdout := os.Stdout
			defer func() {
				os.Stdout = orgStdout
			}()
			r, w, _ := os.Pipe()
			os.Stdout = w
			SlackClient{client}.postSlack("message")
			w.Close()
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("failed to read buf: %v", err)
			}
			got := strings.TrimRight(buf.String(), "\n")
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}
