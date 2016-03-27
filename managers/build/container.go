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
	"log"

	"github.com/gocilla/gocilla/managers/docker"
	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
)

// ContainerManager type.
// Manager to execute a pipeline in a docker container.
type ContainerManager struct {
	database      *mongodb.Database
	dockerManager *docker.Manager
	buildSpec     *Spec
	pipeline      *PipelineSpec
	trigger       *TriggerSpec
	event         *github.Event
	dockerSHA     string
	buildLog      io.Writer
}

// NewContainerManager is the constructor for ContainerManager.
func NewContainerManager(database *mongodb.Database, dockerManager *docker.Manager, buildSpec *Spec,
	pipeline *PipelineSpec, trigger *TriggerSpec, event *github.Event, dockerSHA string, buildLog io.Writer) *ContainerManager {
	return &ContainerManager{database, dockerManager, buildSpec, pipeline, trigger, event, dockerSHA, buildLog}
}

// ExecutePipeline executes the pipeline corresponding to the build triggered.
func (containerBuildManager *ContainerManager) ExecutePipeline() error {
	event := containerBuildManager.event
	dockerSHA := containerBuildManager.dockerSHA

	buildWriter, err := mongodb.NewBuildWriter(
		containerBuildManager.database,
		containerBuildManager.event.Organization,
		containerBuildManager.event.Repository,
		containerBuildManager.event.Type,
		containerBuildManager.event.Branch,
		containerBuildManager.trigger.Pipeline,
		containerBuildManager.trigger.EnvVars)
	if err != nil {
		log.Println("Error creating build writer:", err)
		return err
	}

	user := containerBuildManager.buildSpec.Docker.User
	workingDir := containerBuildManager.buildSpec.Docker.WorkingDir
	log.Println("En CreateAndStartContainer")
	containerManager, err := containerBuildManager.dockerManager.CreateAndStartContainer(
		event.Organization, event.Repository, dockerSHA, user, workingDir,
		containerBuildManager.trigger.EnvVars)
	if err != nil {
		log.Println("Error creating and starting the container:", err)
		buildWriter.EndBuild("error", "Error creating and starting the container")
		return err
	}
	defer containerManager.RemoveContainer()

	log.Println("En GitProjectClone")
	if err := containerBuildManager.GitProjectClone(containerManager, event); err != nil {
		buildWriter.EndBuild("error", "Error cloning the project")
		return err
	}

	log.Println("En ExecutePipelineJobs")
	if err := containerBuildManager.ExecutePipelineJobs(containerManager, buildWriter); err != nil {
		buildWriter.EndBuild("error", "Error executing the pipeline")
		return err
	}
	log.Printf("Completed successfully execution of pipeline '%s'", containerBuildManager.pipeline.Name)
	buildWriter.EndBuild("success", "")
	return nil
}

// GitProjectClone clones a GitHub project in the container.
func (containerBuildManager *ContainerManager) GitProjectClone(containerManager *docker.ContainerManager, event *github.Event) error {
	commands := []string{
		fmt.Sprintf("git clone %s .", event.GitURL),
	}
	if event.Type == "pull" {
		commands = append(commands, fmt.Sprintf("git fetch origin %s:pr/merge", event.SHA))
		commands = append(commands, "git checkout pr/merge")
	} else {
		commands = append(commands, fmt.Sprintf("git checkout %s", event.SHA))
	}
	for _, command := range commands {
		log.Printf("Executing command: %s", command)
		err := containerManager.ExecContainer(command, containerBuildManager.buildLog)
		if err != nil {
			log.Println("Error executing command", err)
			return err
		}
	}
	return nil
}

// ExecutePipelineJobs executes the list of jobs of the pipeline.
func (containerBuildManager *ContainerManager) ExecutePipelineJobs(containerManager *docker.ContainerManager, buildWriter *mongodb.BuildWriter) error {
	for _, job := range containerBuildManager.pipeline.Jobs {
		command := containerBuildManager.buildSpec.Jobs[job]
		buildWriter.StartBuildTask(job, command)
		log.Printf("Executing job '%s' with command: %s", job, command)
		err := containerManager.ExecContainer(command, containerBuildManager.buildLog)
		if err != nil {
			log.Println("Error executing command", err)
			buildWriter.EndBuildTask("error", "Error executing command")
			return err
		}
		buildWriter.EndBuildTask("success", "")
	}
	return nil
}
