/*
 *    Copyright 2021 InfAI (CC SES)
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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	influxdb "github.com/SENERGY-Platform/influx-wrapper/pkg/influx"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/util"
	"github.com/julienschmidt/httprouter"
	influxLib "github.com/orourkedd/influxdb1-client"
	"log"
	"net/http"
	"sort"
	"time"
)

type format string

const (
	perQuery format = "per_query"
	table    format = "table"
)

func init() {
	endpoints = append(endpoints, QueriesEndpoint)
}

func QueriesEndpoint(router *httprouter.Router, config configuration.Config, influx *influxdb.Influx) {
	router.POST("/queries", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		start := time.Now()
		requestedFormat := format(request.URL.Query().Get("format"))
		db := request.Header.Get(userHeader)
		if db == "" {
			http.Error(writer, "Missing header "+userHeader, http.StatusBadRequest)
			return
		}

		var requestElements []influxdb.QueriesRequestElement
		err := json.NewDecoder(request.Body).Decode(&requestElements)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		for _, requestElement := range requestElements {
			if !requestElement.Valid() {
				http.Error(writer, "Invalid request body", http.StatusBadRequest)
				return
			}
		}
		query, err := influxdb.GenerateQueries(requestElements)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := influx.ExecuteQuery(db, query)
		if err != nil {
			switch err {
			case influxdb.ErrInfluxConnection, influxdb.ErrNULL:
				http.Error(writer, err.Error(), http.StatusBadGateway)
				return
			case influxdb.ErrNotFound:
				http.Error(writer, err.Error(), http.StatusNotFound)
				return
			default:
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		response, err := formatResponse(requestedFormat, requestElements, data.Results)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(response)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
		}

		if config.Debug {
			log.Println("Took " + time.Since(start).String())
		}
	})

}

func formatResponse(f format, request []influxdb.QueriesRequestElement, results []influxLib.Result) (data interface{}, err error) {
	switch f {
	case perQuery:
		return formatResponsePerQuery(results)
	case table:
		return formatResponseAsTable(request, results)
	default:
		return formatResponsePerQuery(results)
	}
}

func formatResponsePerQuery(results []influxLib.Result) (formatted [][][]interface{}, err error) {
	for _, result := range results {
		if result.Series == nil {
			// add empty column
			formatted = append(formatted, [][]interface{}{})
			continue
		}
		if len(result.Series) != 1 {
			return nil, errors.New("unexpected number of series")
		}
		// add data
		formatted = append(formatted, result.Series[0].Values)
	}
	return
}

func formatResponseAsTable(request []influxdb.QueriesRequestElement, results []influxLib.Result) (formatted [][]interface{}, err error) {
	data, err := formatResponsePerQuery(results)
	if err != nil {
		return nil, err
	}

	totalColumns := 1
	baseIndex := map[int]int{}
	for requestIndex, element := range request {
		baseIndex[requestIndex] = totalColumns
		totalColumns += len(element.Columns)
	}

	// transform all time strings to time.Time
	for seriesIndex := range data {
		for rowIndex := range data[seriesIndex] {
			data[seriesIndex][rowIndex][0], err = time.Parse(time.RFC3339, data[seriesIndex][rowIndex][0].(string))
			if err != nil {
				return nil, err
			}
		}
		sort.Slice(data[seriesIndex], func(i, j int) bool {
			return data[seriesIndex][i][0].(time.Time).After(data[seriesIndex][j][0].(time.Time))
		})
	}

	for seriesIndex := range data {
		for rowIndex := range data[seriesIndex] {
			formattedRow := make([]interface{}, totalColumns)
			formattedRow[0] = data[seriesIndex][rowIndex][0]
			for seriesColumnIndex := range request[seriesIndex].Columns {
				formattedRow[baseIndex[seriesIndex]+seriesColumnIndex] = data[seriesIndex][rowIndex][seriesColumnIndex+1]
			}
			for subSeriesIndex := range data {
				if subSeriesIndex <= seriesIndex {
					continue
				}
				subRowIndex, ok := findFirstElementIndex(data[subSeriesIndex], data[seriesIndex][rowIndex][0].(time.Time), 0, len(data[subSeriesIndex])-1)
				if ok {
					if data[subSeriesIndex][subRowIndex][0] == data[seriesIndex][rowIndex][0] {
						for subSeriesColumnIndex := range request[subSeriesIndex].Columns {
							formattedRow[baseIndex[subSeriesIndex]+subSeriesColumnIndex] = data[subSeriesIndex][subRowIndex][subSeriesColumnIndex+1]
						}
						data[subSeriesIndex] = util.RemoveElementFrom2D(data[subSeriesIndex], subRowIndex)
						// sorting required for binary search
						sort.Slice(data[subSeriesIndex], func(i, j int) bool {
							return data[subSeriesIndex][i][0].(time.Time).After(data[subSeriesIndex][j][0].(time.Time))
						})
						break
					}
				}
			}
			formatted = append(formatted, formattedRow)
		}
	}
	sort.Slice(formatted, func(i, j int) bool {
		return formatted[i][0].(time.Time).After(formatted[j][0].(time.Time))
	})
	return
}

func findFirstElementIndex(array [][]interface{}, t time.Time, left int, right int) (int, bool) {
	if left > right {
		return 0, false
	}
	mid := left + (right-left)/2
	if t.Equal(array[mid][0].(time.Time)) {
		for t.Equal(array[mid][0].(time.Time)) && mid > 0 {
			mid--
		}
		return mid, true
	}

	if array[mid][0].(time.Time).Before(t) {
		return findFirstElementIndex(array, t, left, mid-1)
	}
	return findFirstElementIndex(array, t, mid+1, right)
}
