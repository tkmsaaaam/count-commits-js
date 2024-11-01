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
	allOneJson, _ := testData.ReadFile("testdata/CountOverAYear/allOne.json")
	overAYear := make([][]byte, 0)
	overAYear = append(overAYear, allOneJson)
	overAYear = append(overAYear, todayAndYesterDayAreOneJson)

	ed := time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local)
	ed2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name                       string
		args                       args
		queryStr                   [][]byte
		ed                         *time.Time
		wantTodayContributionCount int
		wantCountDays              int
		wantED                     time.Time
	}{
		{
			name:                       "todayAndYesterdayAreZero",
			args:                       args{userName: "octocat"},
			queryStr:                   todayAndYesterDayArezero,
			ed:                         &ed,
			wantTodayContributionCount: 0,
			wantCountDays:              0,
			wantED:                     time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
		},
		{
			name:                       "todayIsZeroYesterdayIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayIsZeroYesterdayIsOne,
			ed:                         &ed,
			wantTodayContributionCount: 0,
			wantCountDays:              362,
			wantED:                     time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
		},
		{
			name:                       "todayIsOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayIsOne,
			ed:                         &ed2,
			wantTodayContributionCount: 1,
			wantCountDays:              1,
			wantED:                     time.Date(2023, 1, 2, 0, 0, 0, 0, time.Local),
		},
		{
			name:                       "todayAndYesterDayAreOne",
			args:                       args{userName: "octocat"},
			queryStr:                   todayAndYesterDayAreOne,
			wantTodayContributionCount: 1,
			ed:                         &ed,
			wantCountDays:              363,
			wantED:                     time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
		},
		{
			name:                       "overAYear",
			args:                       args{userName: "octocat"},
			queryStr:                   overAYear,
			wantTodayContributionCount: 1,
			ed:                         &ed,
			wantCountDays:              727,
			wantED:                     time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local),
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
			gotTodayContributionCount, gotCountDays, _, ed := countOverAYear(tt.args.userName, client, time.Date(2023, 1, 3, 12, 0, 0, 0, time.Local), tt.ed)
			if gotTodayContributionCount != tt.wantTodayContributionCount {
				t.Errorf("add() = %v, want %v", gotTodayContributionCount, tt.wantTodayContributionCount)
			}
			if gotCountDays != tt.wantCountDays {
				t.Errorf("add() = %v, want %v", gotCountDays, tt.wantCountDays)
			}
			if *ed != tt.wantED {
				t.Errorf("add() = %v, want %v", ed, tt.wantED)
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

func TestIsContinue(t *testing.T) {
	type args struct {
		i         int
		lastCount int
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "IIsZero",
			args: args{i: 0, lastCount: 0},
			want: true,
		},
		{
			name: "LastCountIsZero",
			args: args{i: 1, lastCount: 0},
			want: false,
		},
		{
			name: "LastCountIsOne",
			args: args{i: 1, lastCount: 1},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isContinue(tt.args.i, tt.args.lastCount)
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountCommits(t *testing.T) {
	type args struct {
		query Query
		time  *time.Time
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	ed := time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local)
	ed2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name   string
		args   args
		want   int
		wantEd *time.Time
	}{
		{
			name:   "queryIsNil",
			args:   args{query: Query{}, time: nil},
			want:   0,
			wantEd: nil,
		},
		{
			name:   "weeksLengthIsZero",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{}}}}}, time: nil},
			want:   0,
			wantEd: nil,
		},
		{
			name:   "commitsTodayIsZero",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 0, Date: today}}}}}}}}, time: nil},
			want:   0,
			wantEd: nil,
		},
		{
			name:   "countCommitsTodayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}, time: &ed},
			want:   1,
			wantEd: &ed,
		},
		{
			name:   "countCommitsTodayIsTwo",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 2, Date: today}}}}}}}}, time: &ed2},
			want:   1,
			wantEd: &ed2,
		},
		{
			name:   "countCommitsTodayIsZeroAndYesterdayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}, time: &ed2},
			want:   1,
			wantEd: &ed2,
		},
		{
			name:   "countCommitsTodayIsZeroAndStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 0, Date: today}}}}}}}}, time: &ed2},
			want:   2,
			wantEd: &ed2,
		},
		{
			name:   "countCommitsTodayIsOneAndStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}, time: &ed2},
			want:   3,
			wantEd: &ed2,
		},
		{
			name:   "countCommitsTodayIsOneAndYesterdayIsOne",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}, time: &ed2},
			want:   2,
			wantEd: &ed2,
		},
		{
			name:   "countCommitsTodayIsOneAndYesterdayIsOneInLastWeek",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: yesterday}}}, {ContributionDays: []ContributionDay{{ContributionCount: 1, Date: today}}}}}}}}, time: &ed2},
			want:   2,
			wantEd: &ed2,
		},
		{
			name:   "noStreak",
			args:   args{query: Query{User{ContributionsCollection{ContributionCalendar{Weeks: []Week{{ContributionDays: []ContributionDay{{ContributionCount: 1, Date: twoDaysAgo}, {ContributionCount: 0, Date: yesterday}, {ContributionCount: 1, Date: today}}}}}}}}, time: &ed2},
			want:   1,
			wantEd: &ed2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, end := countCommittedDays(tt.args.query, now, tt.args.time)
			if got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
			if end != tt.wantEd {
				t.Errorf("add() = %v, want %v", end, tt.wantEd)
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
		start             *time.Time
	}

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "notCommitedAndNoStreak",
			args: args{countCommitsToday: 0, countDays: 0, total: 0, userName: "octocat", start: &start},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は0\n合計コミット数は0\n平均コミット数は0.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "notCommitedAnd",
			args: args{countCommitsToday: 0, countDays: 1, total: 1, userName: "octocat", start: &start},
			want: "<!channel> 今日はまだコミットしていません！\n連続コミット日数は1\n合計コミット数は1\n平均コミット数は1.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "commitedNoStreak",
			args: args{countCommitsToday: 1, countDays: 0, total: 1, userName: "octocat", start: &start},
			want: "\n今日のコミット数は1\n連続コミット日数は0\n合計コミット数は1\n平均コミット数は0.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
		{
			name: "commited",
			args: args{countCommitsToday: 1, countDays: 1, total: 1, userName: "octocat", start: &start},
			want: "\n今日のコミット数は1\n連続コミット日数は1\n合計コミット数は1\n平均コミット数は1.000000\n期間は2023-01-01 ~\nhttps://github.com/octocat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createMessage(tt.args.countCommitsToday, tt.args.countDays, tt.args.total, tt.args.userName, tt.args.start); got != tt.want {
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
