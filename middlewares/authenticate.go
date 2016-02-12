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

package middlewares

import (
	"log"
	"net/http"

	"github.com/gocilla/gocilla/managers/session"
)

type AuthenticateFunc func(http.HandlerFunc) http.HandlerFunc

func Authenticate(sessionManager *session.SessionManager) AuthenticateFunc {
	return func(fn http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("Getting web session")
			session, _ := sessionManager.GetSession(r)
			if session.Values["accessToken"] == "" {
				log.Println("User is not authenticated")
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			} else {
				log.Println("User is already authenticated")
				fn(w, r)
			}
		}
	}
}
