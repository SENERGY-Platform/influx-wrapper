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
)

type Influx struct {
	config configuration.Config
	client Client
}

type TimeValuePair struct {
	Time  *string     `json:"time"`
	Value interface{} `json:"value"`
}

type RequestElement struct {
	Measurement string  `json:"measurement"`
	ColumnName  string  `json:"columnName"`
	Math        *string `json:"math"`
}

type uniqueMeasurementsColumns struct {
	Measurements map[string]struct{}            //Keys of the map are all requested measurements
	Columns      map[string]map[string]struct{} //Keys of the outer map are all requested columns.
	// Keys of the inner maps are all requested math operations. Empty string is no math operation
}

type Client interface {
	Query(query influxLib.Query) (*influxLib.Response, error)
}

var ErrInfluxConnection = errors.New("communication with InfluxDB failed")
var ErrNotFound = errors.New("not found")
var ErrNULL = errors.New("NULL response")
