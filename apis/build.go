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
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gorilla/mux"
)

// BuildAPI type.
// API to manage a build launched in the platform.
type BuildAPI struct {
	Database *mongodb.Database
}

// NewBuildAPI is the constructor for BuildAPI type.
func NewBuildAPI(database *mongodb.Database) *BuildAPI {
	return &BuildAPI{database}
}

// GetLog is an API resource to get the logs corresponding to a build.
func (buildAPI BuildAPI) GetLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildLogFile := fmt.Sprintf("/%s/%s/%s", vars["orgId"], vars["repoId"], vars["buildId"])
	log.Printf("Getting log for build: %s", buildLogFile)

	buildLog, err := buildAPI.Database.OpenFile(buildLogFile)
	if err != nil {
		log.Printf("Error getting log file: %s. %s", buildLogFile, err)
		if err.Error() == "not found" {
			w.WriteHeader(404)
			w.Write([]byte("Not found log: " + buildLogFile))
			return
		}
		w.WriteHeader(500)
		w.Write([]byte("Error getting log: " + buildLogFile))
		return
	}
	defer buildLog.Close()
	io.Copy(w, buildLog)
}
