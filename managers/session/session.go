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
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionConfig struct {
	Name string
	Key  string
}

type SessionManager struct {
	Config *SessionConfig
	Store  *sessions.CookieStore
}

func NewSessionManager(config *SessionConfig) *SessionManager {
	store := sessions.NewCookieStore([]byte(config.Key))
	return &SessionManager{config, store}
}

func (sessionManager SessionManager) GetSession(r *http.Request) (session *sessions.Session, err error) {
	session, err = sessionManager.Store.Get(r, sessionManager.Config.Name)
	return
}

func (sessionManager SessionManager) DestroySession(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionManager.Store.Get(r, sessionManager.Config.Name)
	session.Options = &sessions.Options{MaxAge: -1}
	log.Println("En DestroySession")
	session.Save(r, w)
}
