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
	"net/url"
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
	responseP, err := this.executeQuery(db, query)
	if err != nil {
		return timeValuePairs, err
	}

	if len(responseP.Results) != 1 {
		return timeValuePairs, ErrNULL
	}

	for i := range responseP.Results[0].Series {
		if len(responseP.Results[0].Series[i].Values) != 1 || len(responseP.Results[0].Series[i].Values[0]) != len(set.Columns) {
			return timeValuePairs, ErrNULL
		}
	}

	for _, pair := range pairs {
		seriesIndex, err := findSeriesIndex(pair.Measurement, responseP.Results[0].Series)
		if err != nil {
			if err == ErrNotFound {
				timeValuePairs = append(timeValuePairs, TimeValuePair{
					Time:  nil,
					Value: nil,
				})
				continue
			}
			return timeValuePairs, err
		}

		columnIndex, err := findColumnIndex(pair.ColumnName, responseP.Results[0].Series[seriesIndex])
		if err != nil {
			if err == ErrNotFound {
				time := responseP.Results[0].Series[seriesIndex].Values[0][0].(string)
				timeValuePairs = append(timeValuePairs, TimeValuePair{
					Time:  &time,
					Value: nil,
				})
				continue
			}
			return timeValuePairs, err
		}

		time := responseP.Results[0].Series[seriesIndex].Values[0][0].(string)
		timeValuePairs = append(timeValuePairs, TimeValuePair{
			Time:  &time,
			Value: responseP.Results[0].Series[seriesIndex].Values[0][columnIndex],
		})
	}

	return
}
