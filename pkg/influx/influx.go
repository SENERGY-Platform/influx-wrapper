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
	influxLib "github.com/orourkedd/influxdb1-client"
	"log"
	"net"
	"net/url"
	"strings"
)

type Influx struct {
	config configuration.Config
	client *influxLib.Client
}

var ErrNULL = errors.New("NULL response")
var ErrUnexpectedLength = errors.New("NULL response")
var ErrInfluxConnection = errors.New("communication with InfluxDB failed")
var ErrNotFound = errors.New("not found")

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

func (this *Influx) GetLatestValue(db string, measurement string, field string) (timeValuePair TimeValuePair, err error) {
	query := "SELECT time, " + field + " FROM \"" + measurement + "\" ORDER BY time desc LIMIT 1"
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
			return timeValuePair, ErrInfluxConnection
		}
		return timeValuePair, err
	}
	if responseP == nil {
		return timeValuePair, ErrNULL
	}
	err = responseP.Error()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			if this.config.Debug {
				log.Println(err.Error())
			}
			return timeValuePair, ErrNotFound
		}
		return timeValuePair, responseP.Error()
	}
	if len(responseP.Results) != 1 ||
		len(responseP.Results[0].Series) != 1 ||
		len(responseP.Results[0].Series[0].Values) != 1 ||
		len(responseP.Results[0].Series[0].Values[0]) != 2 {
		return timeValuePair, ErrNotFound
	}
	timeValuePair = TimeValuePair{
		Time:  (responseP.Results[0].Series[0].Values[0][0]).(string),
		Value: responseP.Results[0].Series[0].Values[0][1],
	}
	return
}
