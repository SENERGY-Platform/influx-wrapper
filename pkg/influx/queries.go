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

import (
	"github.com/SENERGY-Platform/influx-wrapper/pkg/api/model"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/util"
	"strconv"
	"strings"
)

func GenerateQueries(elements []model.QueriesRequestElement, timeDirection model.Direction) (query string, err error) {
	for _, element := range elements {
		if len(query) > 0 {
			query += "; "
		}
		query += "SELECT "
		for idx, column := range element.Columns {
			if idx > 0 {
				query += ", "
			}
			if column.GroupType != nil {
				if strings.HasPrefix(*column.GroupType, "difference") {
					groupParts := strings.Split(*column.GroupType, "-")
					query += "difference(" + groupParts[1] + "(\"" + column.Name + "\"))"
				} else {
					query += *column.GroupType + "(\"" + column.Name + "\")"
				}
			} else {
				query += "\"" + column.Name + "\""
			}
			if column.Math != nil {
				query += *column.Math
			}
		}

		query += " FROM \"" + element.Measurement + "\""
		if element.Filters != nil || element.Time != nil {
			query += " WHERE "
		}
		if element.Filters != nil {
			for idx, filter := range *element.Filters {
				if idx != 0 {
					query += " AND "
				}
				_, valueIsString := filter.Value.(string)
				query += "\"" + filter.Column + "\" "
				if filter.Math != nil {
					query += *filter.Math + " "
				}
				query += filter.Type
				if valueIsString {
					query += " '" + filter.Value.(string) + "'"
				} else {
					value, err := util.String(filter.Value)
					if err != nil {
						return "", err
					}
					query += " " + value
				}
			}
		}
		if element.Time != nil {
			if element.Filters != nil {
				query += " AND "
			}
			if element.Time.Last != nil {
				query += " time > now() - " + *element.Time.Last
			} else {
				query += " time > '" + *element.Time.Start + "' AND time < '" + *element.Time.End + "'"
			}
		}
		if element.GroupTime != nil {
			query += " GROUP BY time(" + *element.GroupTime + ")"
		} else {
			query += " ORDER BY time " + strings.ToUpper(string(timeDirection))
		}
		if element.Limit != nil {
			query += " LIMIT " + strconv.Itoa(*element.Limit)
		}
	}
	return
}
