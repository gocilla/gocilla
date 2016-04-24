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

// Trigger type.
type Trigger struct {
	ID           bson.ObjectId    `bson:"_id,omitempty" json:"id"`
	Organization string           `bson:"organization" json:"organization"`
	Repository   string           `bson:"repository" json:"repository"`
	Event        string           `bson:"event" json:"event"`
	Branch       string           `bson:"branch" json:"branch"`
	Pipeline     string           `bson:"pipeline" json:"pipeline"`
	EnvVars      []PipelineEnvVar `bson:"envVars" json:"envVars"`
}

/*
// PipelineEnvVar type.
type PipelineEnvVar struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}
*/

// FindTriggers to retrieve the triggers for a repository.
func (database *Database) FindTriggers(organization, repository string) []Trigger {
	collection := database.Session.DB("").C("triggers")
	var triggers []Trigger
	err := collection.Find(bson.M{"organization": organization, "repository": repository}).All(&triggers)
	log.Println(err)
	return triggers
}

// GetTrigger to get a trigger for a specific event on a repository.
func (database *Database) GetTrigger(organization, repository, event, branch string) (*Trigger, error) {
	collection := database.Session.DB("").C("triggers")
	var trigger Trigger
	err := collection.Find(bson.M{
		"organization": organization,
		"repository":   repository,
		"event":        event,
		"branch":       branch,
	}).One(&trigger)
	return &trigger, err
}

// CreateTrigger to insert a trigger.
func (database *Database) CreateTrigger(trigger *Trigger) error {
	collection := database.Session.DB("").C("triggers")
	trigger.ID = bson.NewObjectId()
	return collection.Insert(*trigger)
}
