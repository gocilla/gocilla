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
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Build struct {
	Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Organization string `bson:"organization" json:"organization"`
	Repository string `bson:"repository" json:"repository"`
	Event string `bson:"event" json:"event"`
	Branch string `bson:"branch" json:"branch"`
	Pipeline string `bson:"pipeline" json:"pipeline"`
	Status string `bson:"status" json:"status"`
	Error string `bson:"error,omitempty" json:"error,omitempty"`
	Start *time.Time `bson:"start" json:"start"`
	End *time.Time `bson:"end,omitempty" json:"end,omitempty"`
	EnvVars []string `bson:"envVars" json:"envVars"`
	Tasks []*BuildTask `bson:"tasks" json:"tasks"`
}

type BuildTask struct {
	Name string `bson:"name" json:"name"`
	Command string `bson:"command" json:"command"`
	Status string `bson:"status" json:"status"`
	Error string `bson:"error,omitempty" json:"error,omitempty"`
	Start time.Time `bson:"start" json:"start"`
	End time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

func (database *Database) CreateBuild(build *Build) error {
	collection := database.Session.DB("").C("builds")
	build.Id = bson.NewObjectId()
	return collection.Insert(*build)
}

func (database *Database) FindBuilds() ([]Build, error) {
	collection := database.Session.DB("").C("builds")
	var builds []Build
	err := collection.Find(bson.M{}).Sort("-start").Limit(10).All(&builds)
	return builds, err
}

func (database *Database) UpdateBuild(id bson.ObjectId, status, error string, end time.Time) error {
	collection := database.Session.DB("").C("builds")
	err := collection.UpdateId(
		id,
		bson.M{"$set": bson.M{"status": status, "error": error, "end": end}})
	return err
}

func (database *Database) AddBuildTask(id bson.ObjectId, buildTask *BuildTask) error {
	collection := database.Session.DB("").C("builds")
	err := collection.UpdateId(
		id,
		bson.M{"$addToSet": bson.M{"tasks": buildTask}})
	return err
}

func (database *Database) UpdateBuildTask(id bson.ObjectId, counter int, status, error string, end time.Time) error {
	collection := database.Session.DB("").C("builds")
	err := collection.Update(
		bson.M{"_id": id, fmt.Sprintf("tasks.%d", counter): bson.M{"$exists": true}},
		bson.M{"$set": bson.M{"tasks.$.status": status, "tasks.$.error": error, "tasks.$.end": end}})
	return err
}

type BuildWriter struct {
	Build *Build
	Counter int
	Database *Database
}

func NewBuildWriter(database *Database, trigger *Trigger, envVars []string) (*BuildWriter, error) {
	now := time.Now()
	build := &Build{
		Organization: trigger.Organization,
		Repository: trigger.Repository,
		Event: trigger.Event,
		Branch: trigger.Branch,
		Pipeline: trigger.Pipeline,
		Status: "running",
		Start: &now,
		EnvVars: envVars,
		Tasks: []*BuildTask{},
	}
	buildWriter := &BuildWriter{
		Build: build,
		Counter: 0,
		Database: database,
	}
	err := database.CreateBuild(build)
	return buildWriter, err
}

func (buildWriter *BuildWriter) StartBuildTask(name, command string) error {
	buildTask := &BuildTask{
		Name: name,
		Command: command,
		Status: "running",
		Start: time.Now(),
	}
	return buildWriter.Database.AddBuildTask(buildWriter.Build.Id, buildTask)
}

func (buildWriter *BuildWriter) EndBuildTask(status, error string) error {
	err := buildWriter.Database.UpdateBuildTask(buildWriter.Build.Id, buildWriter.Counter, status, error, time.Now())
	buildWriter.Counter++
	return err
}

func (buildWriter *BuildWriter) EndBuild(status, error string) error {
	return buildWriter.Database.UpdateBuild(buildWriter.Build.Id, status, error, time.Now())
}
