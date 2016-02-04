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

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"./apis"
	"./config"
	"./middlewares"
	"./managers/build"
	"./managers/docker"
	"./managers/github"
	"./managers/mongodb"
	"./managers/oauth2"
	"./managers/session"
)

func main() {
	config, err := config.Decode("config.json")
	if err != nil {
		log.Println("Configuration error:", err)
		return
	}

	// Mongo
	database, err := mongodb.NewDatabase(config.Mongodb)
	defer database.Close()

	// Managers
	sessionManager := session.NewSessionManager(config.Session)
	oauth2Manager := oauth2.NewOAuth2Manager(config.OAuth2, sessionManager)
	githubManager := github.NewGitHubManager(config.GitHub)
	dockerManagers := docker.NewDockerClusterManager(config.Docker)
	buildManager := build.NewBuildManager(database, oauth2Manager, githubManager, dockerManagers)

	// Middlewares
	authenticate := middlewares.Authenticate(sessionManager)

	// Apis
	buildsApi := apis.NewBuildsApi(database)
	eventsApi := apis.NewEventsApi(buildManager)
	organizationsApi := apis.NewOrganizationsApi(database, oauth2Manager, githubManager)
	triggersApi := apis.NewTriggersApi(database)
	usersApi := apis.NewUsersApi(oauth2Manager, githubManager)

	// Routing
	r := mux.NewRouter()
	r.HandleFunc("/login", oauth2Manager.Authorize).Methods("GET")
	r.HandleFunc("/login/callback", oauth2Manager.AuthorizeCallback).Methods("GET")
	r.HandleFunc("/logout", oauth2Manager.Logout).Methods("GET")
	r.HandleFunc("/api/builds", buildsApi.GetBuilds).Methods("GET")
	r.HandleFunc("/api/events", eventsApi.LaunchBuild).Methods("POST")
	r.HandleFunc("/api/organizations", authenticate(organizationsApi.GetOrganizations)).Methods("GET")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/hook",
		authenticate(organizationsApi.CreateHook)).Methods("POST")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/hook",
		authenticate(organizationsApi.DeleteHook)).Methods("DELETE")
	r.HandleFunc("/api/profile", middlewares.LoggingHandler(authenticate(usersApi.GetProfile))).Methods("GET")
	r.HandleFunc("/api/triggers", triggersApi.GetTriggers).Methods("GET")
	r.HandleFunc("/api/triggers", triggersApi.CreateTrigger).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.Handle("/", r)
	log.Println("Listening at", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
