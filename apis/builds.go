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

// BuildsAPI type.
// API to manage the builds launched in the platform.
type BuildsAPI struct {
	Database *mongodb.Database
}

// NewBuildsAPI is the constructor for BuildsAPI type.
func NewBuildsAPI(database *mongodb.Database) *BuildsAPI {
	return &BuildsAPI{database}
}

// GetBuilds is an API resource to get the builds launched by the platform.
func (buildsAPI BuildsAPI) GetBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := buildsAPI.Database.FindBuilds()
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error getting builds from database"))
		return
	}
	jsonBuilds, err := json.Marshal(builds)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error marshalling the builds"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBuilds)
}
