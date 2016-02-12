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

package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Config type.
type Config struct {
	Name string
	Key  string
}

// Manager type.
// Manager to handle HTTP sessions via cookies.
type Manager struct {
	Config *Config
	Store  *sessions.CookieStore
}

// NewManager is the constructor for session Manager.
func NewManager(config *Config) *Manager {
	store := sessions.NewCookieStore([]byte(config.Key))
	return &Manager{config, store}
}

// GetSession to get a HTTP session.
func (sessionManager Manager) GetSession(r *http.Request) (session *sessions.Session, err error) {
	session, err = sessionManager.Store.Get(r, sessionManager.Config.Name)
	return
}

// DestroySession to destroy a HTTP session.
func (sessionManager Manager) DestroySession(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionManager.Store.Get(r, sessionManager.Config.Name)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
}
