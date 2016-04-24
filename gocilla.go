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
	"os"

	"github.com/gorilla/mux"

	"github.com/gocilla/gocilla/apis"
	"github.com/gocilla/gocilla/config"
	"github.com/gocilla/gocilla/managers/build"
	"github.com/gocilla/gocilla/managers/docker"
	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gocilla/gocilla/managers/oauth2"
	"github.com/gocilla/gocilla/managers/session"
	"github.com/gocilla/gocilla/middlewares"
)

func main() {
	// Read the configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}
	log.Println("Configuration path:", configPath)
	config, err := config.Decode(configPath)
	if err != nil {
		log.Println("Configuration error:", err)
		return
	}

	// Mongo
	database, _ := mongodb.NewDatabase(config.Mongodb)
	defer database.Close()

	// Managers
	sessionManager := session.NewManager(config.Session)
	oauth2Manager := oauth2.NewManager(config.OAuth2, sessionManager)
	githubManager := github.NewManager(config.GitHub)
	dockerManagers := docker.NewManagers(config.Docker)
	buildManager := build.NewManager(database, oauth2Manager, githubManager, dockerManagers)

	// Middlewares
	authenticate := middlewares.Authenticate(sessionManager)
	logging := middlewares.LoggingHandler

	// Apis
	eventsAPI := apis.NewEventsAPI(buildManager)
	organizationsAPI := apis.NewOrganizationsAPI(database, oauth2Manager, githubManager)
	repositoryAPI := apis.NewRepositoryAPI(database, oauth2Manager, githubManager)
	buildAPI := apis.NewBuildAPI(database)
	triggersAPI := apis.NewTriggersAPI(database)
	usersAPI := apis.NewUsersAPI(oauth2Manager, githubManager)

	// Routing
	r := mux.NewRouter()
	r.HandleFunc("/login", logging(oauth2Manager.Authorize)).Methods("GET")
	r.HandleFunc("/login/callback", logging(oauth2Manager.AuthorizeCallback)).Methods("GET")
	r.HandleFunc("/logout", logging(oauth2Manager.Logout)).Methods("GET")
	r.HandleFunc("/api/events", logging(eventsAPI.LaunchBuild)).Methods("POST")
	r.HandleFunc("/api/organizations", logging(authenticate(organizationsAPI.GetOrganizations))).Methods("GET")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}",
		logging(authenticate(repositoryAPI.GetRepository))).Methods("GET")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}",
		logging(authenticate(repositoryAPI.UpdateRepository))).Methods("POST")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/builds",
		logging(authenticate(repositoryAPI.GetBuilds))).Methods("GET")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/builds/{buildId}/logs",
		logging(authenticate(buildAPI.GetLog))).Methods("GET")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/hook",
		logging(authenticate(repositoryAPI.CreateHook))).Methods("POST")
	r.HandleFunc("/api/organizations/{orgId}/repositories/{repoId}/hook",
		logging(authenticate(repositoryAPI.DeleteHook))).Methods("DELETE")
	r.HandleFunc("/api/profile", logging(authenticate(usersAPI.GetProfile))).Methods("GET")
	r.HandleFunc("/api/triggers", logging(triggersAPI.GetTriggers)).Methods("GET")
	r.HandleFunc("/api/triggers", logging(triggersAPI.CreateTrigger)).Methods("POST")
	// Static content
	r.PathPrefix("/public").Handler(http.FileServer(http.Dir("./")))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	})
	http.Handle("/", r)
	log.Println("Listening at", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
