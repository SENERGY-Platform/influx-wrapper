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

package influx

import "errors"

func (this *Influx) GetTags(db string, measurement string) (tagMap map[string][]string, err error) {
	response, err := this.executeQuery(db, "SHOW TAG VALUES FROM \""+measurement+"\" WITH KEY =~ /.*/ ")
	if err != nil {
		return nil, err
	}
	tagMap = make(map[string][]string)
	if len(response.Results) == 0 {
		return tagMap, nil
	}
	if len(response.Results) > 1 {
		return tagMap, errors.New("unexpected response length (more than one result)")
	}
	if len(response.Results[0].Series) == 0 {
		return tagMap, nil
	}
	if len(response.Results[0].Series) > 1 {
		return tagMap, errors.New("unexpected response length (more than one series)")
	}
	for _, row := range response.Results[0].Series[0].Values {
		if len(row) != 2 {
			return nil, errors.New("unexpected response length (not 2 values per row)")
		}
		tagKey := row[0].(string)
		tagValue := row[1].(string)
		tagArray, ok := tagMap[tagKey]
		if !ok {
			tagArray = []string{}
		}
		tagMap[tagKey] = append(tagArray, tagValue)
	}
	return tagMap, nil
}
