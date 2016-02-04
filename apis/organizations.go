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

	"../managers/oauth2"
	"../managers/github"
	"../managers/mongodb"
)

type Organization struct {
	Name *string `json:"name"`
	AvatarURL *string `json:"avatarURL,omitempty"`
	Repositories []*Repository `json:"repositories,omitempty"`
}

type Repository struct {
	Name *string `json:"name"`
	Description *string `json:"description,omitempty"`
	GitURL *string `json:"gitURL,omitempty"`
	Hooked bool `json:"hooked,omitempty"`
}

type OrganizationsApi struct {
	Database *mongodb.Database
	OAuth2Manager *oauth2.OAuth2Manager
	GitHubManager *github.GitHubManager
}

func NewOrganizationsApi(database *mongodb.Database, oauth2Manager *oauth2.OAuth2Manager, githubManager *github.GitHubManager) *OrganizationsApi {
	return &OrganizationsApi{database, oauth2Manager, githubManager}
}

func (organizationsApi OrganizationsApi) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsApi.OAuth2Manager.GetClient(r)
    githubClient := organizationsApi.GitHubManager.NewGitHubClient(oauth2Client)

	/*
	// Get the personal organization
	user, _ := githubClient.GetUser()
	// Get the other organizations
	orgs, _ := githubClient.GetOrganizations()
	// Merge the personal organization and the other organizations
	organizations := make([]Organization, len(orgs) + 1)
	organizations[0] = Organization{user.Login, user.AvatarURL}
	for i := range orgs {
		organizations[i + 1] = Organization{orgs[i].Login, orgs[i].AvatarURL}
	}
	jsonOrganizations, err := json.Marshal(organizations)
	if err != nil {
		w.Write([]byte("Error marshalling the organizations"))
		return
	}
	w.Write(jsonOrganizations)
	*/

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
		hooks := organizationsApi.Database.FindHooks(*organization.Name)
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

/*
func (organizationsApi OrganizationsApi) GetRepositories(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsApi.OAuth2Manager.GetClient(r)
    githubClient := organizationsApi.GitHubManager.NewGitHubClient(oauth2Client)
	vars := mux.Vars(r)
	orgId := vars["orgId"]
	log.Println("Getting repositories for organization", orgId)
	// Get organization repositories
	repos, _ := githubClient.GetRepositories(orgId)
	// Get organization hooks
	hooks := organizationsApi.Database.FindHooks(orgId)
	log.Println(hooks)
	repositories := make([]Repository, len(repos))
	for i := range repos {
		hooked := false
		for _, hook := range hooks {
			if hook.Repository == *(repos[i].Name) {
				hooked = true
				break
			}
		}
		repositories[i] = Repository{repos[i].Name, repos[i].Description, repos[i].GitURL, hooked}
	}
	jsonRepositories, err := json.Marshal(repositories)
	if err != nil {
		w.Write([]byte("Error marshalling the repositories"))
		return
	}
	w.Write(jsonRepositories)
}
*/

func (organizationsApi OrganizationsApi) CreateHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsApi.OAuth2Manager.GetClient(r)
    githubClient := organizationsApi.GitHubManager.NewGitHubClient(oauth2Client)
	vars := mux.Vars(r)
	orgId := vars["orgId"]
	repoId := vars["repoId"]
	log.Println("Creating hook for organization", orgId, "and repository", repoId)
	hookId, err := githubClient.CreateHook(orgId, repoId)
	if err == nil {
		accessToken := organizationsApi.OAuth2Manager.GetSessionAccessToken(r)
		organizationsApi.Database.CreateHook(*hookId, orgId, repoId, accessToken)
	}
}

func (organizationsApi OrganizationsApi) DeleteHook(w http.ResponseWriter, r *http.Request) {
	oauth2Client := organizationsApi.OAuth2Manager.GetClient(r)
    githubClient := organizationsApi.GitHubManager.NewGitHubClient(oauth2Client)
	vars := mux.Vars(r)
	orgId := vars["orgId"]
	repoId := vars["repoId"]
	log.Println("Deleting hook for organization", orgId, "and repository", repoId)
	hook, err := organizationsApi.Database.GetHook(orgId, repoId)
	if err != nil {
		log.Printf("Error getting hook for organization '%s' and repository '%s'. %s", orgId, repoId, err)
		return
	}
	githubClient.DeleteHook(orgId, repoId, hook.Id)
	organizationsApi.Database.DeleteHook(hook.Id)
}
