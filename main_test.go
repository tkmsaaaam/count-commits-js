package main

import (
	"bytes"
	"context"
	"embed"
	"log"
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

//go:embed testdata/*
var testData embed.FS

func TestCountOverAYear(t *testing.T) {
	type args struct {
		userName string
	}

	todayAndYesterDayArezeroJson, _ := testData.ReadFile("testdata/CountOverAYear/todayAndYesterdayAreZero.json")
	todayAndYesterDayArezero := make([][]byte, 0)
	todayAndYesterDayArezero = append(todayAndYesterDayArezero, todayAndYesterDayArezeroJson)
	todayIsOneJson, _ := testData.ReadFile("testdata/CountOverAYear/todayIsOne.json")
	todayIsOne := make([][]byte, 0)
	todayIsOne = append(todayIsOne, todayIsOneJson)
	todayIsZeroYesterdayIsOneJson, _ := testData.ReadFile("testdata/CountOverAYear/todayIsZeroYesterdayIsOne.json")
	todayIsZeroYesterdayIsOne := make([][]byte, 0)
	todayIsZeroYesterdayIsOne = append(todayIsZeroYesterdayIsOne, todayIsZeroYesterdayIsOneJson)
	todayAndYesterDayAreOneJson, _ := testData.ReadFile("testdata/CountOverAYear/todayAndYesterdayAreOne.json")
	todayAndYesterDayAreOne := make([][]byte, 0)
	todayAndYesterDayAreOne = append(todayAndYesterDayAreOne, todayAndYesterDayAreOneJson)

	tests := []struct {
		name                       string
		args                       args
		queryStr                   [][]byte
		wantTodayContributionCount int
		wantCountDays              int
		wantED                     time.Time
	}{
		{
			name:                       "todayAndYesterdayAreZero",
			args:                       args{userName: "octocat"},
			queryStr:                   todayAndYesterDayArezero,
			wantTodayContributionCount: 0,
			wantCountDays:              0,
			wantED:                     time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:                       "todayIsZeroYesterdayIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayIsZeroYesterdayIsOne,
			wantTodayContributionCount: 1,
			wantCountDays:              363,
			wantED:                     time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:                       "todayIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayIsOne,
			wantTodayContributionCount: 1,
			wantCountDays:              1,
			wantED:                     time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:                       "todayAndYesterDayAreOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayAndYesterDayAreOne,
			wantTodayContributionCount: 1,
			wantCountDays:              364,
			wantED:                     time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		var i = 0
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
			w.Write(tt.queryStr[i])
			i++
		})
		t.Run(tt.name, func(t *testing.T) {
			log.Println(tt.name)
			r := &Result{userName: tt.args.userName, today: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), start: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC), latestDay: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), isContinue: true}
			err := r.countOverAYear(client)
			if err != nil {
				t.Errorf("countOverAYear() err = %v", err)
			}
			if r.todayContributionCount != tt.wantTodayContributionCount {
				t.Errorf("countOverAYear() todayContributionCount = %v, want %v", r.todayContributionCount, tt.wantTodayContributionCount)
			}
			if r.streak != tt.wantCountDays {
				t.Errorf("countOverAYear() streak = %v, want %v", r.streak, tt.wantCountDays)
			}
			if r.start != tt.wantED {
				t.Errorf("countOverAYear() start = %v, want %v", r.start, tt.wantED)
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
			queryStr: "testdata/ExecQuery/queryIsNil.json",
			want:     Query{},
		},
		{
			name:     "totalContributionsIsZero",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "testdata/ExecQuery/totalContributionsIsZero.json",
			want:     Query{User{ContributionsCollection{ContributionCalendar{}}}},
		},
		{
			name:     "totalContributionsIsOne",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "testdata/ExecQuery/totalContributionsIsOne.json",
			want:     Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: "2023-01-01"}}}}}}}},
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, _ *http.Request) {
			res, _ := testData.ReadFile(tt.queryStr)
			w.Write(res)
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
			queryStr: "testdata/ExecQuery/queryIsNil.json",
			want:     "",
		},
		{
			name:     "execQueryIsError",
			args:     args{ctx: context.Background(), variables: map[string]interface{}{"name": githubv4.String("octocat")}},
			queryStr: "testdata/ExecQuery/queryIsNil.json",
			want:     "query is error. non-200 OK status code: 500 Internal Server Error body: \"Internal Server Error\"",
		},
	}
	for _, tt := range tests {
		mux := http.NewServeMux()
		client := githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
		mux.HandleFunc("/graphql", func(w http.ResponseWriter, _ *http.Request) {
			if tt.name == "execQueryIsOk" {
				res, _ := testData.ReadFile(tt.queryStr)
				w.Write(res)
			} else {
				w.WriteHeader(500)
				w.Write([]byte("Internal Server Error"))
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			var buf bytes.Buffer
			log.SetOutput(&buf)
			defaultFlags := log.Flags()
			log.SetFlags(0)
			defer func() {
				log.SetOutput(os.Stderr)
				log.SetFlags(defaultFlags)
				buf.Reset()
			}()
			Client{client}.execQuery(tt.args.ctx, tt.args.variables)
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

func TestCountCommits(t *testing.T) {
	type args struct {
		query Query
	}

	now := time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)
	today := now.Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	tests := []struct {
		name   string
		args   args
		want   int
		wantEd time.Time
	}{
		{
			name:   "queryIsNil",
			args:   args{query: Query{}},
			want:   0,
			wantEd: now,
		},
		{
			name:   "weeksLengthIsZero",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{}}}}}},
			want:   0,
			wantEd: now,
		},
		{
			name:   "commitsTodayIsZero",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: today}}}}}}}}},
			want:   0,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsTwo",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsZeroAndYesterdayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}},
			want:   0,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsZeroAndStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}},
			want:   0,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsOneAndStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsOneAndYesterdayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
		{
			name:   "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
		{
			name:   "noStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 0, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}},
			want:   1,
			wantEd: now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{today: now, latestDay: now.AddDate(0, 0, 1), start: now, isContinue: true}
			err := r.countCommittedDays(tt.args.query)
			if err != nil {
				t.Errorf("countCommittedDays() = %v", err)
			}
			if r.streak != tt.want {
				t.Errorf("add() = %v, want %v", r.streak, tt.want)
			}
			if r.start != tt.wantEd {
				t.Errorf("add() = %v, want %v", r.start, tt.wantEd)
			}
		})
	}
}

func TestCreateMessage(t *testing.T) {
	type args struct {
		countCommitsToday int
		countDays         int
		total             int
		userName          string
	}

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "notCommitedAndNoStreak",
			args: args{countCommitsToday: 0, countDays: 0, total: 0, userName: "octocat"},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は0\n合計コミット数は0\n平均コミット数は0.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "notCommitedAnd",
			args: args{countCommitsToday: 0, countDays: 1, total: 1, userName: "octocat"},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は1\n合計コミット数は1\n平均コミット数は1.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "commitedNoStreak",
			args: args{countCommitsToday: 1, countDays: 0, total: 1, userName: "octocat"},
			want: "\n今日のコミット数は1\n連続コミット日数は0\n合計コミット数は1\n平均コミット数は0.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "commited",
			args: args{countCommitsToday: 1, countDays: 1, total: 1, userName: "octocat"},
			want: "\n今日のコミット数は1\n連続コミット日数は1\n合計コミット数は1\n平均コミット数は1.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{todayContributionCount: tt.args.countCommitsToday, total: tt.args.total, streak: tt.args.countDays, latestDay: start, userName: tt.args.userName, isContinue: true}
			if got := r.createMessage(); got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostSlack(t *testing.T) {
	tests := []struct {
		name   string
		apiRes string
		want   string
	}{
		{
			name:   "iSOk",
			apiRes: "testdata/slack/ok.json",
			want:   "",
		},
		{
			name:   "isError",
			apiRes: "testdata/slack/error.json",
			want:   "can not post message. too_many_attachments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := slacktest.NewTestServer(func(c slacktest.Customize) {
				c.Handle("/chat.postMessage", func(w http.ResponseWriter, _ *http.Request) {
					res, _ := testData.ReadFile(tt.apiRes)
					w.Write(res)
				})
			})
			ts.Start()
			client := slack.New("testToken", slack.OptionAPIURL(ts.GetAPIURL()))

			t.Helper()

			var buf bytes.Buffer
			log.SetOutput(&buf)
			defaultFlags := log.Flags()
			log.SetFlags(0)
			defer func() {
				log.SetOutput(os.Stderr)
				log.SetFlags(defaultFlags)
				buf.Reset()
			}()
			SlackClient{client}.postSlack("message")
			got := strings.TrimRight(buf.String(), "\n")
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}
