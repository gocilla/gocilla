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

	"github.com/gocilla/gocilla/managers/build"
	"github.com/gocilla/gocilla/managers/github"
)

// EventsAPI type.
// API to receive GitHub events (e.g. a PullRequest or a Push).
// Note that these events may launch a build if configured in gocilla.
type EventsAPI struct {
	BuildManager *build.Manager
}

// NewEventsAPI is the constructor for EventsAPI type.
func NewEventsAPI(buildManager *build.Manager) *EventsAPI {
	return &EventsAPI{buildManager}
}

// LaunchBuild is the API resource that processes the GitHub event.
func (eventsAPI EventsAPI) LaunchBuild(w http.ResponseWriter, r *http.Request) {
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
	go func() {
		err := eventsAPI.BuildManager.Build(event)
		if err != nil {
			log.Println("Error in build", err)
		}
	}()
	w.WriteHeader(200)
}
