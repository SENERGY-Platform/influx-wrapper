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

package pkg

import (
	"context"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/api"
	"github.com/SENERGY-Platform/influx-wrapper/pkg/configuration"
	"sync"
)

//starts services and goroutines; returns a waiting group which is done as soon as all go routines are stopped
func Start(ctx context.Context, config configuration.Config) (wg *sync.WaitGroup, err error) {
	wg = &sync.WaitGroup{}
	err = api.Start(ctx, wg, config)
	return
}
