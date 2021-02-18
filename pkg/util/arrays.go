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

package util

func ElementInArray(element interface{}, array []interface{}) bool {
	for _, comparison := range array {
		if comparison == element {
			return true
		}
	}
	return false
}

/*
Removes an element form an array. If the array was ordered before, it will loose that order.
*/
func RemoveElementFrom2D(array [][]interface{}, index int) [][]interface{} {
	array[len(array)-1], array[index] = array[index], array[len(array)-1]
	return array[:len(array)-1]
}
