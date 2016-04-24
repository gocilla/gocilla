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

	"github.com/gorilla/mux"

	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gocilla/gocilla/managers/oauth2"
)

// Repository type.
type Repository struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description,omitempty"`
	GitURL      *string          `json:"gitURL,omitempty"`
	Hooked      bool             `json:"hooked,omitempty"`
	Builds      *[]mongodb.Build `json:"builds"`
}

// RepositoryAPI type.
// API to manage a repository (including the hooks to receive GitHub events).
type RepositoryAPI struct {
	Database      *mongodb.Database
	OAuth2Manager *oauth2.Manager
	GitHubManager *github.Manager
}

// NewRepositoryAPI is the constructor for RepositoryAPI.
func NewRepositoryAPI(database *mongodb.Database, oauth2Manager *oauth2.Manager, githubManager *github.Manager) *RepositoryAPI {
	return &RepositoryAPI{database, oauth2Manager, githubManager}
}

// GetRepository is the API resource that returns the settings of the repository.
func (repositoryAPI RepositoryAPI) GetRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Printf("Getting settings for repository: %s/%s", orgID, repoID)

	repository, err := repositoryAPI.Database.GetRepository(orgID, repoID)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error getting repository from database"))
		return
	}

	jsonRepository, err := json.Marshal(repository)
	if err != nil {
		w.Write([]byte("Error marshalling the repository"))
		return
	}
	w.Write(jsonRepository)
}

// UpdateRepository is the API resource that updates the settings of the repository.
func (repositoryAPI RepositoryAPI) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Printf("Updating settings for repository: %s/%s", orgID, repoID)

	var repository mongodb.Repository
	if err := json.NewDecoder(r.Body).Decode(&repository); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Error decoding JSON repository"))
		return
	}
	if err := repositoryAPI.Database.UpdateRepository(&repository); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Error updating the repository"))
		return
	}
	w.WriteHeader(200)
}

// GetBuilds is the API resource that returns the builds of the repository.
func (repositoryAPI RepositoryAPI) GetBuilds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Printf("Getting builds for repository: %s/%s", orgID, repoID)

	builds, err := repositoryAPI.Database.FindRepositoryBuilds(orgID, repoID)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Error getting repository builds from database"))
		return
	}

	jsonBuilds, err := json.Marshal(builds)
	if err != nil {
		w.Write([]byte("Error marshalling the builds"))
		return
	}
	w.Write(jsonBuilds)
}

// CreateHook is a resource API to create a GitHub hook on a repository.
// Organization and repository are specified as parts of the request path.
func (repositoryAPI RepositoryAPI) CreateHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := repositoryAPI.OAuth2Manager.GetClient(r)
	githubClient := repositoryAPI.GitHubManager.NewClient(oauth2Client)
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Println("Creating hook for organization", orgID, "and repository", repoID)
	hookID, err := githubClient.CreateHook(orgID, repoID)
	if err == nil {
		accessToken := repositoryAPI.OAuth2Manager.GetSessionAccessToken(r)
		repositoryAPI.Database.CreateHook(*hookID, orgID, repoID, accessToken)
	}
}

// DeleteHook is a resource API to delete  a GitHub hook on a repository.
// Organization and repository are specified as parts of the request path.
func (repositoryAPI RepositoryAPI) DeleteHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := repositoryAPI.OAuth2Manager.GetClient(r)
	githubClient := repositoryAPI.GitHubManager.NewClient(oauth2Client)
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Println("Deleting hook for organization", orgID, "and repository", repoID)
	hook, err := repositoryAPI.Database.GetHook(orgID, repoID)
	if err != nil {
		log.Printf("Error getting hook for organization '%s' and repository '%s'. %s", orgID, repoID, err)
		return
	}
	githubClient.DeleteHook(orgID, repoID, hook.ID)
	repositoryAPI.Database.DeleteHook(hook.ID)
}
