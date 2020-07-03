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
	"testing"
)

func TestInflux(t *testing.T) {
	t.Run("NewInflux", func(t *testing.T) {
		t.Run("invalid url", func(t *testing.T) {
			config := configuration.ConfigStruct{
				InfluxDbUrl: "http://ur:l/:80",
			}
			_, err := NewInflux(&config)
			if err == nil {
				t.Fail()
			}
		})
		t.Run("no error", func(t *testing.T) {
			config := configuration.ConfigStruct{
				InfluxDbUrl:  "http://url:80",
				InfluxDbUser: "user",
				InfluxDbPw:   "pw",
			}
			influx, err := NewInflux(&config)
			if err != nil {
				t.Error(err.Error())
			}
			if influx == nil || influx.config == nil || influx.config != &config {
				t.Fail()
			}
		})
	})

	t.Run("GetLatestValues", func(t *testing.T) {
		influxClientMock := services.NewClientMock()
		influxClient := Influx{
			config: &configuration.ConfigStruct{
				Debug: true,
			},
			client: &influxClientMock,
		}

		t1 := "2000-01-01T00:00:00.000Z"
		t2 := "2000-01-02T00:00:00.000Z"
		db := "db"
		v1 := 1
		v2 := 2

		t.Run("single", func(t *testing.T) {
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Results: []influxLib.Result{
					{
						Series: []models.Row{
							{
								Name:    "m1",
								Columns: []string{"time", "c1"},
								Values: [][]interface{}{
									{t1, v1},
								},
							},
						},
					},
				},
			}, nil)
			measurementColumnPair := RequestElement{
				Measurement: "m1",
				ColumnName:  "c1",
			}
			t.Run("normal", func(t *testing.T) {
				actual, err := influxClient.GetLatestValue(db, measurementColumnPair)
				expect := TimeValuePair{
					Time:  &t1,
					Value: 1,
				}
				if err != nil {
					t.Error(err.Error())
					return
				}
				if !timeValuePairEquals(actual, expect) {
					t.Fail()
				}
			})
			t.Run("error", func(t *testing.T) {
				testErr := errors.New("random err")
				influxClientMock.SetQueryResponse(&influxLib.Response{}, testErr)
				_, err := influxClient.GetLatestValue(db, measurementColumnPair)
				if err != testErr {
					t.Fail()
				}
			})
			t.Run("no result", func(t *testing.T) {
				influxClientMock.SetQueryResponse(&influxLib.Response{}, nil)
				_, err := influxClient.GetLatestValue(db, measurementColumnPair)
				if err != ErrNULL {
					t.Fail()
				}
			})
			t.Run("no values", func(t *testing.T) {
				influxClientMock.SetQueryResponse(&influxLib.Response{
					Results: []influxLib.Result{
						{
							Series: []models.Row{
								{
									Name:    "m1",
									Columns: []string{"time", "c1"},
									Values:  [][]interface{}{},
								},
							},
						},
					},
				}, nil)
				_, err := influxClient.GetLatestValue(db, measurementColumnPair)
				if err != ErrNULL {
					t.Fail()
				}
			})
		})
		t.Run("multi", func(t *testing.T) {
			influxClientMock.SetQueryResponse(&influxLib.Response{
				Results: []influxLib.Result{
					{
						Series: []models.Row{
							{
								Name:    "m1",
								Columns: []string{"time", "c1", "c2"},
								Values: [][]interface{}{
									{t1, v1, nil},
								},
							},
							{
								Name:    "m2",
								Columns: []string{"time", "c1", "c2"},
								Values: [][]interface{}{
									{t2, nil, v2},
								},
							},
						},
					},
				},
			}, nil)
			measurementColumnPairs := []RequestElement{
				{
					Measurement: "m1",
					ColumnName:  "c1",
				},
				{
					Measurement: "m2",
					ColumnName:  "c2",
				},
			}
			t.Run("normal", func(t *testing.T) {
				expected := []TimeValuePair{
					{
						Time:  &t1,
						Value: v1,
					},
					{
						Time:  &t2,
						Value: v2,
					},
				}
				actual, err := influxClient.GetLatestValues(db, measurementColumnPairs)
				if err != nil {
					t.Fail()
					return
				}
				if !timeValuePairListEquals(expected, actual) {
					t.Fail()
				}
			})
			t.Run("with math", func(t *testing.T) {
				influxClientMock.SetQueryResponse(&influxLib.Response{
					Results: []influxLib.Result{
						{
							Series: []models.Row{
								{
									Name:    "m1",
									Columns: []string{"time", "c1+5", "c1-5"},
									Values: [][]interface{}{
										{t1, v1 + 5, v1 - 5},
									},
								},
							},
						},
					},
				}, nil)
				expected := []TimeValuePair{
					{
						Time:  &t1,
						Value: v1 + 5,
					},
					{
						Time:  &t1,
						Value: v1 - 5,
					},
				}
				math1 := "+5"
				math2 := "-5"
				actual, err := influxClient.GetLatestValues(db, []RequestElement{
					{
						Measurement: "m1",
						ColumnName:  "c1",
						Math:        &math1,
					},
					{
						Measurement: "m1",
						ColumnName:  "c1",
						Math:        &math2,
					},
				})
				if err != nil {
					t.Fail()
					return
				}
				if !timeValuePairListEquals(expected, actual) {
					t.Fail()
				}
			})
			t.Run("series missing", func(t *testing.T) {
				influxClientMock.SetQueryResponse(&influxLib.Response{
					Results: []influxLib.Result{
						{
							Series: []models.Row{
								{
									Name:    "m1",
									Columns: []string{"time", "c1", "c2"},
									Values: [][]interface{}{
										{t1, v1, nil},
									},
								},
							},
						},
					},
				}, nil)
				expected := []TimeValuePair{
					{
						Time:  &t1,
						Value: v1,
					},
					{
						Time:  nil,
						Value: nil,
					},
				}
				actual, err := influxClient.GetLatestValues(db, measurementColumnPairs)
				if err != nil {
					t.Fail()
					return
				}
				if !timeValuePairListEquals(expected, actual) {
					t.Fail()
				}
			})
			t.Run("column missing", func(t *testing.T) {
				influxClientMock.SetQueryResponse(&influxLib.Response{
					Results: []influxLib.Result{
						{
							Series: []models.Row{
								{
									Name:    "m1",
									Columns: []string{"time", "c1", "c2"},
									Values: [][]interface{}{
										{t1, v1, nil},
									},
								},
								{
									Name:    "m2",
									Columns: []string{"time", "c1", "c2"},
									Values: [][]interface{}{
										{t1, v1, nil},
									},
								},
							},
						},
					},
				}, nil)
				expected := []TimeValuePair{
					{
						Time:  &t1,
						Value: v1,
					},
					{
						Time:  &t1,
						Value: nil,
					},
				}
				actual, err := influxClient.GetLatestValues(db, measurementColumnPairs)
				if err != nil {
					t.Fail()
					return
				}
				if !timeValuePairListEquals(expected, actual) {
					t.Fail()
				}
			})
		})
	})
}

func timeValuePairEquals(p1 TimeValuePair, p2 TimeValuePair) bool {
	if p1.Value != p2.Value {
		return false
	}
	if p1.Time == nil {
		if p2.Time != nil {
			return false
		}
		return true
	}
	if *p1.Time != *p2.Time {
		return false
	}
	return true
}

func timeValuePairListEquals(l1 []TimeValuePair, l2 []TimeValuePair) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i := range l1 {
		if !timeValuePairEquals(l1[i], l2[i]) {
			return false
		}
	}
	return true
}
