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
}
