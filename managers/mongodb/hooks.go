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

package mongodb

import (
	"log"

	"gopkg.in/mgo.v2/bson"
)

// Hook type.
type Hook struct {
	ID           int    `bson:"_id"`
	Organization string `bson:"organization"`
	Repository   string `bson:"repository"`
	AccessToken  string `bson:"accessToken"`
}

// FindHooks to retrieve the list of hooks available for an organization.
func (database *Database) FindHooks(organization string) []Hook {
	collection := database.Session.DB("").C("hooks")
	var hooks []Hook
	err := collection.Find(bson.M{"organization": organization}).All(&hooks)
	log.Println(err)
	return hooks
}

// GetHook to get a hook for a repository.
func (database *Database) GetHook(organization string, repository string) (Hook, error) {
	collection := database.Session.DB("").C("hooks")
	var hook Hook
	err := collection.Find(bson.M{"organization": organization, "repository": repository}).One(&hook)
	return hook, err
}

// CreateHook to create a hook for a repository.
func (database *Database) CreateHook(id int, organization, repository, accessToken string) {
	collection := database.Session.DB("").C("hooks")
	doc := Hook{id, organization, repository, accessToken}
	err := collection.Insert(doc)
	log.Println(err)
}

// DeleteHook to remove a hook for a repository.
func (database *Database) DeleteHook(id int) {
	collection := database.Session.DB("").C("hooks")
	err := collection.RemoveId(id)
	log.Println(err)
}
