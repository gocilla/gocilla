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

package oauth2

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/gocilla/gocilla/managers/session"
)

// Config type.
type Config struct {
	Strategy oauth2.Config
	State    string
}

// Manager type.
// Manager to handle an OAuth2 session. OAuth2 is required to invoke the GitHub APIs on behalf the user.
type Manager struct {
	Config         *Config
	SessionManager *session.Manager
}

// NewManager is the constructor for OAuth2 Manager.
func NewManager(config *Config, sessionManager *session.Manager) *Manager {
	return &Manager{config, sessionManager}
}

// SetSessionAccessToken to store the OAuth2 access token in the session
func (oauth2Manager Manager) SetSessionAccessToken(accessToken string, w http.ResponseWriter, r *http.Request) {
	session, _ := oauth2Manager.SessionManager.GetSession(r)
	session.Values["accessToken"] = accessToken
	session.Save(r, w)
	log.Printf("Stored token %s in web session", accessToken)
}

// GetSessionAccessToken to get the OAuth2 access token from the session.
func (oauth2Manager Manager) GetSessionAccessToken(r *http.Request) string {
	session, _ := oauth2Manager.SessionManager.GetSession(r)
	if session != nil && session.Values["accessToken"] != nil {
		return session.Values["accessToken"].(string)
	}
	return ""
}

// GetClient to set up an OAuth2 HTTP client able to access GitHub APIs. The access token
// is obtained from the session.
func (oauth2Manager Manager) GetClient(r *http.Request) *http.Client {
	accessToken := oauth2Manager.GetSessionAccessToken(r)
	return oauth2Manager.GetClientFromAccessToken(accessToken)
}

// GetClientFromAccessToken to set up an OAuth2 HTTP client able to access GitHub APIs. The access token
// is passed as a parameter.
func (oauth2Manager Manager) GetClientFromAccessToken(accessToken string) *http.Client {
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "bearer",
	}
	return oauth2Manager.Config.Strategy.Client(oauth2.NoContext, token)
}

// Authorize to request for OAuth2 authorization against GitHub.
func (oauth2Manager Manager) Authorize(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got token from cookie: %s", oauth2Manager.GetSessionAccessToken(r))
	strategy := oauth2Manager.Config.Strategy
	state := oauth2Manager.Config.State
	url := strategy.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// AuthorizeCallback to handle the authorize callback from GitHub.
func (oauth2Manager Manager) AuthorizeCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("En AuthorizeCallback")
	strategy := oauth2Manager.Config.Strategy
	state := oauth2Manager.Config.State
	formState := r.FormValue("state")
	if state != formState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", state, formState)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := strategy.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Printf("Github token: %s\n", token)

	oauth2Manager.SetSessionAccessToken(token.AccessToken, w, r)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Logout to destroy the session.
func (oauth2Manager Manager) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("En Logout")
	oauth2Manager.SessionManager.DestroySession(w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
