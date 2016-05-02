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

package build

import (
	"fmt"
	"io"

	"gopkg.in/mgo.v2"

	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
)

// Register type.
// Manager to register a build and its operations.
type Register struct {
	Database       *mongodb.Database
	GithubClient   *github.Client
	Event          *github.Event
	Trigger        *TriggerSpec
	BuildWriter    *mongodb.BuildWriter
	BuildLogFile   *mgo.GridFile
	BuildLogWriter io.Writer
}

// NewRegister is the constructor for Register.
func NewRegister(database *mongodb.Database, githubClient *github.Client, event *github.Event, trigger *TriggerSpec) (register *Register, err error) {
	register = &Register{
		Database:     database,
		GithubClient: githubClient,
		Event:        event,
		Trigger:      trigger,
	}

	// Create the build writer in mongodb (with info about the executed steps)
	register.BuildWriter, err = mongodb.NewBuildWriter(
		database,
		event.Organization, event.Repository, event.Type, event.Branch,
		trigger.Pipeline, trigger.EnvVars)
	if err != nil {
		err = fmt.Errorf("Error creating build writer. %s", err)
		return
	}
	buildID := register.BuildWriter.Build.ID.Hex()

	// Write logs to a mongodb gridfs file and to console
	buildLogFileName := fmt.Sprintf("/%s/%s/%s", event.Organization, event.Repository, buildID)
	register.BuildLogFile, err = database.CreateFile(buildLogFileName)
	if err != nil {
		err = fmt.Errorf("Error creating build mongo log file: %s. %s", buildLogFileName, err)
		register.End(err)
		return
	}
	register.BuildLogWriter = io.MultiWriter(register.BuildLogFile)
	return
}

// End logs the end of a pipeline build and closes the shared resources.
func (register *Register) End(err error) {
	if register.BuildWriter != nil {
		status, error := statusFromError(err)
		register.BuildWriter.EndBuild(status, error)
	}
	if register.BuildLogFile != nil {
		register.BuildLogFile.Close()
	}
}

// StartTask logs the start of a pipeline task.
func (register *Register) StartTask(task, command string) {
	logString := fmt.Sprintf("Starting task '%s' with command '%s'\n", task, command)
	io.WriteString(register.BuildLogWriter, logString)

	if register.BuildWriter != nil {
		register.BuildWriter.StartBuildTask(task, command)
	}
	if register.Event.Type == github.EventTypePull && register.GithubClient != nil {
		register.GithubClient.CreateStatus(
			register.Event.Organization, register.Event.Repository, register.Event.Pull.HeadSHA,
			task, command, "pending")
	}
}

// EndTask logs the start of a pipeline task.
func (register *Register) EndTask(task, command string, err error) {
	status, error := statusFromError(err)

	logString := fmt.Sprintf("Ended task '%s' with status '%s'. %s\n", task, status, error)
	io.WriteString(register.BuildLogWriter, logString)

	if register.BuildWriter != nil {
		register.BuildWriter.EndBuildTask(status, error)
	}
	if register.Event.Type == github.EventTypePull && register.GithubClient != nil {
		description := command
		if err != nil {
			description = error
		}
		register.GithubClient.CreateStatus(
			register.Event.Organization, register.Event.Repository, register.Event.Pull.HeadSHA,
			task, description, status)
	}
}

func statusFromError(err error) (status, error string) {
	if err == nil {
		return "success", ""
	}
	return "error", err.Error()
}
