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
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	influxLib "github.com/orourkedd/influxdb1-client"
	"log"
	"net"
	"net/url"
	"strings"
)

func NewInflux(config configuration.Config) (influx *Influx, err error) {
	influxUrl, err := url.Parse(config.InfluxDbUrl)
	if err != nil {
		return influx, err
	}
	influxConfig := influxLib.Config{
		URL:      *influxUrl,
		Username: config.InfluxDbUser,
		Password: config.InfluxDbPw,
	}
	client, err := influxLib.NewClient(influxConfig)
	if err != nil {
		return influx, err
	}
	return &Influx{config: config, client: client}, nil
}

func (this *Influx) GetLatestValue(db string, pair MeasurementColumnPair) (timeValuePair TimeValuePair, err error) {
	timeValuePairs, err := this.GetLatestValues(db, []MeasurementColumnPair{pair})
	if err != nil {
		return timeValuePair, err
	}
	return timeValuePairs[0], err
}

func (this *Influx) GetLatestValues(db string, pairs []MeasurementColumnPair) (timeValuePairs []TimeValuePair, err error) {
	set := transformMeasurementColumnPairs(pairs)

	query := generateQuery(set) + " ORDER BY \"time\" DESC LIMIT 1"
	if this.config.Debug {
		log.Println("Query: " + query)
	}

	responseP, err := this.client.Query(influxLib.Query{
		Command:         query,
		Database:        db,
		RetentionPolicy: "",
	})
	if err != nil {
		_, isNetError := err.(net.Error)
		if isNetError {
			log.Println(err.Error())
			return timeValuePairs, ErrInfluxConnection
		}
		return timeValuePairs, err
	}
	if responseP == nil {
		return timeValuePairs, ErrNULL
	}
	err = responseP.Error()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if this.config.Debug {
				log.Println(err.Error())
			}
			return timeValuePairs, ErrNotFound
		}
		return timeValuePairs, responseP.Error()
	}

	if len(responseP.Results) != 1 || len(responseP.Results[0].Series) != len(set.Measurements) {
		return timeValuePairs, ErrNotFound
	}

	for i := range responseP.Results[0].Series {
		if len(responseP.Results[0].Series[i].Values) != 1 || len(responseP.Results[0].Series[i].Values[0]) != len(set.Columns) {
			return timeValuePairs, ErrNotFound
		}
	}

	for _, pair := range pairs {
		seriesIndex, err := findSeriesIndex(pair.Measurement, responseP.Results[0].Series)
		if err != nil {
			return timeValuePairs, err
		}
		columnIndex, err := findColumnIndex(pair.ColumnName, responseP.Results[0].Series[seriesIndex])
		if err != nil {
			return timeValuePairs, err
		}
		timeValuePairs = append(timeValuePairs, TimeValuePair{
			Time:  responseP.Results[0].Series[seriesIndex].Values[0][0].(string),
			Value: responseP.Results[0].Series[seriesIndex].Values[0][columnIndex],
		})
	}

	return
}
