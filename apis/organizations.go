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

// Organization type.
type Organization struct {
	Name         *string       `json:"name"`
	AvatarURL    *string       `json:"avatarURL,omitempty"`
	Repositories []*Repository `json:"repositories,omitempty"`
}

// Repository type.
type Repository struct {
	Name        *string `json:"name"`
	Description *string `json:"description,omitempty"`
	GitURL      *string `json:"gitURL,omitempty"`
	Hooked      bool    `json:"hooked,omitempty"`
}

// OrganizationsAPI type.
// API to get the organizations of a user, and to manage hooks to receive GitHub events.
type OrganizationsAPI struct {
	Database      *mongodb.Database
	OAuth2Manager *oauth2.Manager
	GitHubManager *github.Manager
}

// NewOrganizationsAPI is the constructor for OrganizationsAPI.
func NewOrganizationsAPI(database *mongodb.Database, oauth2Manager *oauth2.Manager, githubManager *github.Manager) *OrganizationsAPI {
	return &OrganizationsAPI{database, oauth2Manager, githubManager}
}

// GetOrganizations is the API resource that returns the user's organizations.
func (organizationsAPI OrganizationsAPI) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsAPI.OAuth2Manager.GetClient(r)
	githubClient := organizationsAPI.GitHubManager.NewClient(oauth2Client)

	repos, _ := githubClient.GetRepositories()
	// Create an array (final result) and a map (a temporary object to query an organization by name)
	organizations := []*Organization{}
	orgsMap := make(map[string]*Organization)
	// Iterate over the user repositories to build up the "organizations" array
	for _, repo := range repos {
		repository := &Repository{repo.Name, repo.Description, repo.GitURL, false}
		// Find organization. If not available yet, create it
		org, ok := orgsMap[*repo.Owner.Login]
		if !ok {
			// Add the new organization in the map and array
			organization := &Organization{repo.Owner.Login, repo.Owner.AvatarURL, []*Repository{repository}}
			orgsMap[*repo.Owner.Login] = organization
			organizations = append(organizations, organization)
		} else {
			org.Repositories = append(org.Repositories, repository)
		}
	}
	// Iterate over the organizations to set the hooks
	for _, organization := range organizations {
		// Get the hooks for the organization repositories
		hooks := organizationsAPI.Database.FindHooks(*organization.Name)
		for _, repository := range organization.Repositories {
			for _, hook := range hooks {
				if hook.Repository == *repository.Name {
					repository.Hooked = true
					break
				}
			}
		}
	}
	jsonOrganizations, err := json.Marshal(organizations)
	if err != nil {
		w.Write([]byte("Error marshalling the organizations"))
		return
	}
	w.Write(jsonOrganizations)
}

// CreateHook is a resource API to create a GitHub hook on a repository.
// Organization and repository are specified as parts of the request path.
func (organizationsAPI OrganizationsAPI) CreateHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsAPI.OAuth2Manager.GetClient(r)
	githubClient := organizationsAPI.GitHubManager.NewClient(oauth2Client)
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Println("Creating hook for organization", orgID, "and repository", repoID)
	hookID, err := githubClient.CreateHook(orgID, repoID)
	if err == nil {
		accessToken := organizationsAPI.OAuth2Manager.GetSessionAccessToken(r)
		organizationsAPI.Database.CreateHook(*hookID, orgID, repoID, accessToken)
	}
}

// DeleteHook is a resource API to delete  a GitHub hook on a repository.
// Organization and repository are specified as parts of the request path.
func (organizationsAPI OrganizationsAPI) DeleteHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsAPI.OAuth2Manager.GetClient(r)
	githubClient := organizationsAPI.GitHubManager.NewClient(oauth2Client)
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	repoID := vars["repoId"]
	log.Println("Deleting hook for organization", orgID, "and repository", repoID)
	hook, err := organizationsAPI.Database.GetHook(orgID, repoID)
	if err != nil {
		log.Printf("Error getting hook for organization '%s' and repository '%s'. %s", orgID, repoID, err)
		return
	}
	githubClient.DeleteHook(orgID, repoID, hook.Id)
	organizationsAPI.Database.DeleteHook(hook.Id)
}
