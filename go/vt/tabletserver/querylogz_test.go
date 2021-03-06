// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tabletserver

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/youtube/vitess/go/streamlog"
	"github.com/youtube/vitess/go/vt/tabletserver/planbuilder"
	"golang.org/x/net/context"
)

func TestQuerylogzHandlerInvalidSqlQueryStats(t *testing.T) {
	req, _ := http.NewRequest("GET", "/querylogz?timeout=0&limit=10", nil)
	response := httptest.NewRecorder()
	testLogger := streamlog.New("TestLogger", 100)
	testLogger.Send("test msg")
	querylogzHandler(testLogger, response, req)
	if !strings.Contains(response.Body.String(), "error") {
		t.Fatalf("should show an error page for an non SqlQueryStats")
	}
}

func TestQuerylogzHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/querylogz?timeout=0&limit=10", nil)
	logStats := newSqlQueryStats("Execute", context.Background())
	logStats.PlanType = planbuilder.PLAN_PASS_SELECT.String()
	logStats.OriginalSql = "select name from test_table limit 1000"
	logStats.RowsAffected = 1000
	logStats.NumberOfQueries = 1
	logStats.StartTime = time.Unix(123456789, 0)
	logStats.MysqlResponseTime = 1 * time.Millisecond
	logStats.WaitingForConnection = 10 * time.Nanosecond
	logStats.CacheHits = 17
	logStats.CacheAbsent = 5
	logStats.CacheMisses = 2
	logStats.CacheInvalidations = 3
	logStats.TransactionID = 131

	testLogger := streamlog.New("TestLogger", 100)

	// fast query
	fastQueryPattern := []string{
		`<td>Execute</td>`,
		`<td></td>`,
		`<td>Nov 29 13:33:09.000000</td>`,
		`<td>Nov 29 13:33:09.001000</td>`,
		`<td>0.001</td>`,
		`<td>0.001</td>`,
		`<td>1e-08</td>`,
		`<td>PASS_SELECT</td>`,
		`<td>select name from test_table limit 1000</td>`,
		`<td>1</td>`,
		`<td>none</td>`,
		`<td>1000</td>`,
		`<td>0</td>`,
		`<td>17</td>`,
		`<td>2</td>`,
		`<td>5</td>`,
		`<td>3</td>`,
		`<td>131</td>`,
		`<td></td>`,
	}
	logStats.EndTime = logStats.StartTime.Add(1 * time.Millisecond)
	testLogger.Send(logStats)
	response := httptest.NewRecorder()
	querylogzHandler(testLogger, response, req)
	body, _ := ioutil.ReadAll(response.Body)
	checkQuerylogzHasStats(t, fastQueryPattern, logStats, body)

	// medium query
	mediumQueryPattern := []string{
		`<td>Execute</td>`,
		`<td></td>`,
		`<td>Nov 29 13:33:09.000000</td>`,
		`<td>Nov 29 13:33:09.020000</td>`,
		`<td>0.02</td>`,
		`<td>0.001</td>`,
		`<td>1e-08</td>`,
		`<td>PASS_SELECT</td>`,
		`<td>select name from test_table limit 1000</td>`,
		`<td>1</td>`,
		`<td>none</td>`,
		`<td>1000</td>`,
		`<td>0</td>`,
		`<td>17</td>`,
		`<td>2</td>`,
		`<td>5</td>`,
		`<td>3</td>`,
		`<td>131</td>`,
		`<td></td>`,
	}
	logStats.EndTime = logStats.StartTime.Add(20 * time.Millisecond)
	testLogger.Send(logStats)
	response = httptest.NewRecorder()
	querylogzHandler(testLogger, response, req)
	body, _ = ioutil.ReadAll(response.Body)
	checkQuerylogzHasStats(t, mediumQueryPattern, logStats, body)

	// slow query
	slowQueryPattern := []string{
		`<td>Execute</td>`,
		`<td></td>`,
		`<td>Nov 29 13:33:09.000000</td>`,
		`<td>Nov 29 13:33:09.500000</td>`,
		`<td>0.5</td>`,
		`<td>0.001</td>`,
		`<td>1e-08</td>`,
		`<td>PASS_SELECT</td>`,
		`<td>select name from test_table limit 1000</td>`,
		`<td>1</td>`,
		`<td>none</td>`,
		`<td>1000</td>`,
		`<td>0</td>`,
		`<td>17</td>`,
		`<td>2</td>`,
		`<td>5</td>`,
		`<td>3</td>`,
		`<td>131</td>`,
		`<td></td>`,
	}
	logStats.EndTime = logStats.StartTime.Add(500 * time.Millisecond)
	testLogger.Send(logStats)
	querylogzHandler(testLogger, response, req)
	body, _ = ioutil.ReadAll(response.Body)
	checkQuerylogzHasStats(t, slowQueryPattern, logStats, body)
}

func checkQuerylogzHasStats(t *testing.T, pattern []string, logStats *SQLQueryStats, page []byte) {
	matcher := regexp.MustCompile(strings.Join(pattern, `\s*`))
	if !matcher.Match(page) {
		t.Fatalf("querylogz page does not contain stats: %v, page: %s", logStats, string(page))
	}
}
