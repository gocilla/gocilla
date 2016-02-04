// Copyright 2016 Telefónica Investigación y Desarrollo, S.A.U
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
// http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apis

import (
	"log"
	"net/http"

	"../managers/build"
	"../managers/github"
)

type EventsApi struct {
	BuildManager	*build.BuildManager
}

func NewEventsApi(buildManager *build.BuildManager) *EventsApi {
	return &EventsApi{buildManager}
}

func (eventsApi EventsApi) LaunchBuild(w http.ResponseWriter, r *http.Request) {
	event, err := github.ParseEvent(r)
	if err != nil {
		log.Println("Error decoding build payload.", err)
        w.WriteHeader(500)
        return
	}
	if event == nil {
		log.Println("Ignoring the event")
		w.WriteHeader(200)
		return
	}
	go eventsApi.BuildManager.Build(event)
    w.WriteHeader(200)
}
