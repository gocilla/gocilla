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
	"net/http"

	"../managers/oauth2"
	"../managers/github"
)

type Profile struct {
	Login *string `json:"login,omitempty"`
	Name *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatarURL,omitempty"`
	Company *string `json:"company,omitempty"`
}

type UsersApi struct {
	OAuth2Manager *oauth2.OAuth2Manager
	GitHubManager *github.GitHubManager
}

func NewUsersApi(oauth2Manager *oauth2.OAuth2Manager, githubManager *github.GitHubManager) *UsersApi {
	return &UsersApi{oauth2Manager, githubManager}
}

func (usersApi UsersApi) GetProfile(w http.ResponseWriter, r *http.Request) {
	oauth2Client := usersApi.OAuth2Manager.GetClient(r)
	githubClient := usersApi.GitHubManager.NewGitHubClient(oauth2Client)
	user, _ := githubClient.GetUser()
	profile := Profile{user.Login, user.Name, user.AvatarURL, user.Company}
	jsonProfile, err := json.Marshal(profile)
	if err != nil {
		w.Write([]byte("Error marshalling the user profile"))
		return
	}
	w.Write(jsonProfile)
}
