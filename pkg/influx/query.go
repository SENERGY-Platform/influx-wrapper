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
	"log"
	"net"
	"strings"
)

func generateQuery(set uniqueMeasurementsColumns) (query string) {
	columns := []string{}
	measurements := []string{}
	for measurement := range set.Measurements {
		if measurement != "" {
			measurements = append(measurements, "\""+measurement+"\"")
		}
	}
	for columnName, mathOperations := range set.Columns {
		if columnName != "" {
			for mathOperation := range mathOperations {
				part := "\"" + columnName + "\""
				if mathOperation != "" {
					part += mathOperation + " AS \"" + columnName + mathOperation + "\""
				}
				columns = append(columns, part)
			}
		}
	}

	query += "SELECT " + strings.Join(columns, ", ") + " FROM " + strings.Join(measurements, ", ")
	return query
}

func (this *Influx) ExecuteQuery(db string, query string) (responseP *influxLib.Response, err error) {
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
