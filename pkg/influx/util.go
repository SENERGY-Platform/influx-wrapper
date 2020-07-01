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
	influxLib "github.com/orourkedd/influxdb1-client"
	"github.com/orourkedd/influxdb1-client/models"
	"log"
	"net"
	"strings"
)

func generateQuery(set uniqueMeasurementsColumns) (query string) {
	columns := []string{}
	measurements := []string{}
	for measurement := range set.Measurements {
		measurements = append(measurements, "\""+measurement+"\"")
	}
	for column := range set.Columns {
		columns = append(columns, "\""+column+"\"")
	}

	query += "SELECT " + strings.Join(columns, ", ") + " FROM " + strings.Join(measurements, ", ")
	return query
}

func (this *Influx) executeQuery(db string, query string) (responseP *influxLib.Response, err error) {
	if this.config.Debug {
		log.Println("Query: " + query)
	}

	responseP, err = this.client.Query(influxLib.Query{
		Command:         query,
		Database:        db,
		RetentionPolicy: "",
	})
	if err != nil {
		_, isNetError := err.(net.Error)
		if isNetError {
			log.Println(err.Error())
			return responseP, ErrInfluxConnection
		}
		return responseP, err
	}
	if responseP == nil {
		return responseP, ErrNULL
	}
	err = responseP.Error()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if this.config.Debug {
				log.Println(err.Error())
			}
			return responseP, ErrNotFound
		}
		return responseP, err
	}

	return
}

func transformMeasurementColumnPairs(pairs []MeasurementColumnPair) (unique uniqueMeasurementsColumns) {
	unique = uniqueMeasurementsColumns{
		Columns:      make(map[string]struct{}),
		Measurements: make(map[string]struct{}),
	}
	for _, pair := range pairs {
		unique.Columns[pair.ColumnName] = struct{}{}
		unique.Measurements[pair.Measurement] = struct{}{}
	}
	unique.Columns["time"] = struct{}{}
	return unique
}

func findSeriesIndex(name string, series []models.Row) (index int, err error) {
	for index, s := range series {
		if s.Name == name {
			return index, nil
		}
	}
	return 0, ErrNotFound
}

func findColumnIndex(name string, series models.Row) (index int, err error) {
	for index, column := range series.Columns {
		if column == name {
			return index, nil
		}
	}
	return 0, ErrNotFound
}
