/*
 *    Copyright 2020 InfAI (CC SES)
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package influx

import (
	"errors"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/tests/services"
	influxLib "github.com/orourkedd/influxdb1-client"
	"github.com/orourkedd/influxdb1-client/models"
	"strings"
	"testing"
)

func TestUtil(t *testing.T) {
	influxClientMock := services.NewClientMock()

	influxClient := Influx{
		config: &configuration.ConfigStruct{
			Debug: true,
		},
		client: &influxClientMock,
	}

	t.Run("generateQuery", func(t *testing.T) {
		t.Run("empty set", func(t *testing.T) {
			q := generateQuery(uniqueMeasurementsColumns{})
			expect := "SELECT  FROM "
			if q != expect {
				t.Error("expect", expect, "actual", q)
			}
		})
		t.Run("empty strings", func(t *testing.T) {
			measurements := make(map[string]struct{})
			measurements[""] = struct{}{}
			columns := make(map[string]struct{})
			columns[""] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Measurements: measurements,
				Columns:      columns,
			})
			expect := "SELECT  FROM "
			if q != expect {
				t.Error("expect", expect, "actual", q)
			}
		})
		t.Run("empty measurements", func(t *testing.T) {
			columns := make(map[string]struct{})
			columns["c1"] = struct{}{}
			columns["c2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Columns: columns,
			})
			expect := "SELECT \"c1\", \"c2\" FROM "
			expectAlt := "SELECT \"c2\", \"c1\" FROM "
			if q != expect && q != expectAlt {
				t.Error("\nexpect\n", expect, "\nor\n", expectAlt, "\nactual\n", q)
			}
		})
		t.Run("empty columns", func(t *testing.T) {
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Measurements: measurements,
			})
			expect := "SELECT  FROM \"m1\", \"m2\""
			expectAlt := "SELECT  FROM \"m2\", \"m1\""
			if q != expect && q != expectAlt {
				t.Error("expect\n", expect, "\nor\n", expectAlt, "\nactual\n", q)
			}
		})
		t.Run("normal set", func(t *testing.T) {
			columns := make(map[string]struct{})
			columns["c1"] = struct{}{}
			columns["c2"] = struct{}{}
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			q := generateQuery(uniqueMeasurementsColumns{
				Columns:      columns,
				Measurements: measurements,
			})
			validResults := []string{}
			validResults = append(validResults,
				"SELECT \"c1\", \"c2\" FROM \"m1\", \"m2\"",
				"SELECT \"c2\", \"c1\" FROM \"m1\", \"m2\"",
				"SELECT \"c1\", \"c2\" FROM \"m2\", \"m1\"",
				"SELECT \"c2\", \"c1\" FROM \"m2\", \"m1\"")

			foundValid := false

			for _, validResult := range validResults {
				if q == validResult {
					foundValid = true
				}
			}
			if !foundValid {
				t.Error("expect any of\n", strings.Join(validResults, "\n"), "\nactual\n", q)
			}
		})
	})

	t.Run("executeQuery", func(t *testing.T) {
		t.Run("net error", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, netError{
				error: errors.New("net error"),
			})
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrInfluxConnection {
				t.Fail()
			}
		})
		t.Run("other err", func(t *testing.T) {
			testErr := errors.New("other err")
			influxClientMock.SetQueryResponse(nil, testErr)
			_, err := influxClient.executeQuery("test", "test")
			if err != testErr {
				t.Fail()
			}
		})
		t.Run("response nil", func(t *testing.T) {
			influxClientMock.SetQueryResponse(nil, nil)
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrNULL {
				t.Fail()
			}
		})
		t.Run("response not found", func(t *testing.T) {
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Err: errors.New("DB test not found"),
			}, nil)
			_, err := influxClient.executeQuery("test", "test")
			if err != ErrNotFound {
				t.Fail()
			}
		})
		t.Run("response other found", func(t *testing.T) {
			testErr := errors.New("unknown error message")
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Err: testErr,
			}, nil)
			_, err := influxClient.executeQuery("test", "test")
			if err != testErr {
				t.Fail()
			}
		})
		t.Run("response normal", func(t *testing.T) {
			expect := &influxLib.Response{
				Results: []influxLib.Result{
					{
						Series: []models.Row{
							{
								Name: "test",
							},
						},
					},
				},
			}
			influxClientMock.SetQueryResponse(expect, nil)
			actual, err := influxClient.executeQuery("test", "test")
			if err != nil {
				t.Fail()
			}
			if actual != expect {
				t.Fail()
			}
		})
	})

	t.Run("transformMeasurementColumnPairs", func(t *testing.T) {
		t.Run("empty pairs", func(t *testing.T) {
			actual := transformMeasurementColumnPairs([]RequestElement{})

			columns := make(map[string]struct{})
			columns["time"] = struct{}{}
			expect := uniqueMeasurementsColumns{Columns: columns}

			if !uniqueMeasurementsColumnsEquals(actual, expect) {
				t.Fail()
			}
		})
		t.Run("invalid pair", func(t *testing.T) {
			actual := transformMeasurementColumnPairs([]RequestElement{
				{Measurement: ""},
			})

			columns := make(map[string]struct{})
			columns["time"] = struct{}{}
			columns[""] = struct{}{}
			measurements := make(map[string]struct{})
			measurements[""] = struct{}{}
			expect := uniqueMeasurementsColumns{Columns: columns, Measurements: measurements}

			if !uniqueMeasurementsColumnsEquals(actual, expect) {
				t.Fail()
			}
		})
		t.Run("single pair", func(t *testing.T) {
			actual := transformMeasurementColumnPairs([]RequestElement{
				{
					Measurement: "m1",
					ColumnName:  "c1",
				},
			})

			columns := make(map[string]struct{})
			columns["time"] = struct{}{}
			columns["c1"] = struct{}{}
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			expect := uniqueMeasurementsColumns{Columns: columns, Measurements: measurements}

			if !uniqueMeasurementsColumnsEquals(actual, expect) {
				t.Fail()
			}
		})
		t.Run("multiple pairs", func(t *testing.T) {
			actual := transformMeasurementColumnPairs([]RequestElement{
				{
					Measurement: "m1",
					ColumnName:  "c1",
				},
				{
					Measurement: "m2",
					ColumnName:  "c2",
				},
			})

			columns := make(map[string]struct{})
			columns["time"] = struct{}{}
			columns["c1"] = struct{}{}
			columns["c2"] = struct{}{}
			measurements := make(map[string]struct{})
			measurements["m1"] = struct{}{}
			measurements["m2"] = struct{}{}
			expect := uniqueMeasurementsColumns{Columns: columns, Measurements: measurements}

			if !uniqueMeasurementsColumnsEquals(actual, expect) {
				t.Fail()
			}
		})
	})

	t.Run("findSeriesIndex", func(t *testing.T) {
		series := []models.Row{
			{
				Name: "test",
			},
			{
				Name: "test2",
			},
		}
		t.Run("empty name", func(t *testing.T) {
			n, err := findSeriesIndex("", series)
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("empty series", func(t *testing.T) {
			n, err := findSeriesIndex("test", []models.Row{})
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("not found", func(t *testing.T) {
			n, err := findSeriesIndex("no", series)
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("found", func(t *testing.T) {
			n, err := findSeriesIndex("test2", series)
			if err != nil {
				t.Fail()
			}
			if n != 1 {
				t.Fail()
			}
		})
	})

	t.Run("findColumnIndex", func(t *testing.T) {
		series := models.Row{
			Columns: []string{
				"c1", "c2",
			},
		}
		t.Run("empty name", func(t *testing.T) {
			n, err := findColumnIndex("", series)
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("empty series", func(t *testing.T) {
			n, err := findColumnIndex("test", models.Row{})
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("not found", func(t *testing.T) {
			n, err := findColumnIndex("no", series)
			if err != ErrNotFound {
				t.Fail()
			}
			if n != 0 {
				t.Fail()
			}
		})
		t.Run("found", func(t *testing.T) {
			n, err := findColumnIndex("c2", series)
			if err != nil {
				t.Fail()
			}
			if n != 1 {
				t.Fail()
			}
		})
	})
}

type netError struct {
	error
}

func (n netError) Error() string {
	return n.error.Error()
}

func (n netError) Timeout() bool {
	return true
}
func (n netError) Temporary() bool {
	return true
}

func uniqueMeasurementsColumnsEquals(u1 uniqueMeasurementsColumns, u2 uniqueMeasurementsColumns) bool {
	return mapStringStructEquals(u1.Columns, u2.Columns) && mapStringStructEquals(u1.Measurements, u2.Measurements)
}

func mapStringStructEquals(m1 map[string]struct{}, m2 map[string]struct{}) bool {
	for key := range m1 {
		_, ok := m2[key]
		if !ok {
			return false
		}
	}
	for key := range m2 {
		_, ok := m1[key]
		if !ok {
			return false
		}
	}
	return true
}
