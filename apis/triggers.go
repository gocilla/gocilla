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
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocilla/gocilla/managers/mongodb"
)

// TriggersAPI type.
// API to manage the triggers on a GitHub repository.
type TriggersAPI struct {
	Database *mongodb.Database
}

// NewTriggersAPI is the constructor for TriggersAPI.
func NewTriggersAPI(database *mongodb.Database) *TriggersAPI {
	return &TriggersAPI{database}
}

// GetTriggers is the API resource that returns the triggers registered on a repository.
func (triggersAPI TriggersAPI) GetTriggers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	organization := q["organization"][0]
	repository := q["repository"][0]
	log.Println("Find triggers for organization", organization, "and repository", repository)
	triggers := triggersAPI.Database.FindTriggers(organization, repository)
	jsonTriggers, err := json.Marshal(triggers)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error marshalling the triggers"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTriggers)
}

// CreateTrigger is the API resource that creates a new trigger on a repository.
func (triggersAPI TriggersAPI) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	var trigger mongodb.Trigger
	if err := json.NewDecoder(r.Body).Decode(&trigger); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Error decoding JSON trigger"))
		return
	}
	if err := triggersAPI.Database.CreateTrigger(&trigger); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Error creating the trigger"))
		return
	}
	w.Header().Set("Location", "/api/triggers/"+string(trigger.ID))
	w.WriteHeader(201)
}
