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

	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/oauth2"
)

// Profile type.
type Profile struct {
	Login     *string `json:"login,omitempty"`
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatarURL,omitempty"`
	Company   *string `json:"company,omitempty"`
}

// UsersAPI type.
// API to manage GitHub users.
type UsersAPI struct {
	OAuth2Manager *oauth2.Manager
	GitHubManager *github.Manager
}

// NewUsersAPI is the constructor of UsersAPI.
func NewUsersAPI(oauth2Manager *oauth2.Manager, githubManager *github.Manager) *UsersAPI {
	return &UsersAPI{oauth2Manager, githubManager}
}

// GetProfile is the API resource that returns the user's profile.
func (usersAPI UsersAPI) GetProfile(w http.ResponseWriter, r *http.Request) {
	oauth2Client := usersAPI.OAuth2Manager.GetClient(r)
	githubClient := usersAPI.GitHubManager.NewClient(oauth2Client)
	user, err := githubClient.GetUser()
	if err != nil {
		w.Write([]byte("Error getting the user from github"))
		return
	}
	profile := Profile{user.Login, user.Name, user.AvatarURL, user.Company}
	jsonProfile, err := json.Marshal(profile)
	if err != nil {
		w.Write([]byte("Error marshalling the user profile"))
		return
	}
	w.Write(jsonProfile)
}
