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

// Build type.
type Build struct {
	ID           bson.ObjectId     `bson:"_id,omitempty" json:"id"`
	Organization string            `bson:"organization" json:"organization"`
	Repository   string            `bson:"repository" json:"repository"`
	Event        string            `bson:"event" json:"event"`
	Branch       string            `bson:"branch" json:"branch"`
	Pipeline     string            `bson:"pipeline" json:"pipeline"`
	Status       string            `bson:"status" json:"status"`
	Error        string            `bson:"error,omitempty" json:"error,omitempty"`
	Start        *time.Time        `bson:"start" json:"start"`
	End          *time.Time        `bson:"end,omitempty" json:"end,omitempty"`
	EnvVars      map[string]string `bson:"envVars" json:"envVars"`
	Tasks        []*BuildTask      `bson:"tasks" json:"tasks"`
}

// BuildTask type.
type BuildTask struct {
	Name    string    `bson:"name" json:"name"`
	Command string    `bson:"command" json:"command"`
	Status  string    `bson:"status" json:"status"`
	Error   string    `bson:"error,omitempty" json:"error,omitempty"`
	Start   time.Time `bson:"start" json:"start"`
	End     time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

// CreateBuild to insert a new build.
func (database *Database) CreateBuild(build *Build) error {
	collection := database.Session.DB("").C("builds")
	build.ID = bson.NewObjectId()
	return collection.Insert(*build)
}

// FindBuilds to list the latest 10 builds.
func (database *Database) FindBuilds() ([]Build, error) {
	collection := database.Session.DB("").C("builds")
	var builds []Build
	err := collection.Find(bson.M{}).Sort("-start").Limit(10).All(&builds)
	return builds, err
}

// FindRepositoryBuilds to list the latest 50 builds of a repository.
func (database *Database) FindRepositoryBuilds(organization, repository string) ([]Build, error) {
	collection := database.Session.DB("").C("builds")
	var builds []Build
	err := collection.Find(bson.M{"organization": organization, "repository": repository}).Sort("-start").Limit(50).All(&builds)
	return builds, err
}

// UpdateBuild to update the status of a build.
func (database *Database) UpdateBuild(id bson.ObjectId, status, error string, end time.Time) error {
	collection := database.Session.DB("").C("builds")
	err := collection.UpdateId(
		id,
		bson.M{"$set": bson.M{"status": status, "error": error, "end": end}})
	return err
}

// AddBuildTask to insert a task in a build.
func (database *Database) AddBuildTask(id bson.ObjectId, buildTask *BuildTask) error {
	collection := database.Session.DB("").C("builds")
	err := collection.UpdateId(
		id,
		bson.M{"$addToSet": bson.M{"tasks": buildTask}})
	return err
}

// UpdateBuildTask to update a task in a build.
func (database *Database) UpdateBuildTask(id bson.ObjectId, counter int, status, error string, end time.Time) error {
	collection := database.Session.DB("").C("builds")
	err := collection.Update(
		bson.M{"_id": id, fmt.Sprintf("tasks.%d", counter): bson.M{"$exists": true}},
		bson.M{"$set": bson.M{"tasks.$.status": status, "tasks.$.error": error, "tasks.$.end": end}})
	return err
}

// BuildWriter type.
type BuildWriter struct {
	Build    *Build
	Counter  int
	Database *Database
}

// NewBuildWriter is a constructor.
func NewBuildWriter(database *Database, organization, repository, event, branch, pipeline string,
	envVars map[string]string) (*BuildWriter, error) {
	now := time.Now()
	build := &Build{
		Organization: organization,
		Repository:   repository,
		Event:        event,
		Branch:       branch,
		Pipeline:     pipeline,
		Status:       "running",
		Start:        &now,
		EnvVars:      envVars,
		Tasks:        []*BuildTask{},
	}
	buildWriter := &BuildWriter{
		Build:    build,
		Counter:  0,
		Database: database,
	}
	err := database.CreateBuild(build)
	return buildWriter, err
}

// StartBuildTask to insert a task, with "running" status, in a build.
func (buildWriter *BuildWriter) StartBuildTask(name, command string) error {
	buildTask := &BuildTask{
		Name:    name,
		Command: command,
		Status:  "running",
		Start:   time.Now(),
	}
	return buildWriter.Database.AddBuildTask(buildWriter.Build.ID, buildTask)
}

// EndBuildTask to update a task, with completed status, in a build.
func (buildWriter *BuildWriter) EndBuildTask(status, error string) error {
	err := buildWriter.Database.UpdateBuildTask(buildWriter.Build.ID, buildWriter.Counter, status, error, time.Now())
	buildWriter.Counter++
	return err
}

// EndBuild to update a build with completion status.
func (buildWriter *BuildWriter) EndBuild(status, error string) error {
	return buildWriter.Database.UpdateBuild(buildWriter.Build.ID, status, error, time.Now())
}
