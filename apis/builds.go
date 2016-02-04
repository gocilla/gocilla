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

	"../managers/mongodb"
)

type BuildsApi struct {
	Database *mongodb.Database
}

func NewBuildsApi(database *mongodb.Database) *BuildsApi {
	return &BuildsApi{database}
}

func (buildsApi BuildsApi) GetBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := buildsApi.Database.FindBuilds()
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