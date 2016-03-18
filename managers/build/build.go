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
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/gocilla/gocilla/managers/docker"
	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gocilla/gocilla/managers/oauth2"
)

// Manager type.
// Manager to perform a build (after a trigger). It requires other managers:
//   - GitHubManager to access to GitHub to download or clone the repository via API.
//   - OAuth2Manager to help GitHubManager with OAuth2 access.
//   - DockerManagers to launch a container to perform the build on a docker cluster.
type Manager struct {
	Database       *mongodb.Database
	OAuth2Manager  *oauth2.Manager
	GitHubManager  *github.Manager
	DockerManagers docker.Managers
}

// Spec type.
type Spec struct {
	Docker    DockerSpec
	Jobs      map[string]string
	Pipelines []PipelineSpec
	Triggers  []TriggerSpec
}

// DockerSpec type.
type DockerSpec struct {
	File       string
	User       string
	WorkingDir string `json:"workingDir" yaml:"workingDir"`
}

// PipelineSpec type.
type PipelineSpec struct {
	Name string
	Jobs []string
}

// TriggerSpec type.
type TriggerSpec struct {
	Name     string
	Event    string
	Branch   string
	Pipeline string
	EnvVars  map[string]string
}

// NewManager is the constructor of Manager.
func NewManager(database *mongodb.Database, oauth2Manager *oauth2.Manager, githubManager *github.Manager, dockerManagers docker.Managers) *Manager {
	return &Manager{database, oauth2Manager, githubManager, dockerManagers}
}

// Build the project.
// It uses the GitHub event to know which repository and git SHA to be build.
func (buildManager *Manager) Build(event *github.Event) error {
	log.Printf("Starting build process for event: %+v", event)

	hook, err := buildManager.Database.GetHook(event.Organization, event.Repository)
	if err != nil {
		log.Printf("Error getting hook. %s", err)
		return err
	}
	log.Printf("Using hook: %+v", hook)

	githubClient := buildManager.GitHubManager.NewClient(buildManager.OAuth2Manager.GetClientFromAccessToken(hook.AccessToken))
	buildSpec, err := buildManager.GetSpec(githubClient, event)
	if err != nil {
		log.Printf("Error getting the project specification. %s", err)
		return err
	}

	trigger := buildManager.GetTrigger(buildSpec, event)
	if trigger == nil {
		return fmt.Errorf("No trigger matching the event '%s' and branch '%s'", event.Type, event.Branch)
	}

	pipeline := buildManager.GetPipeline(buildSpec, trigger)
	if pipeline == nil {
		return fmt.Errorf("No pipeline matching the trigger pipeline: %s", trigger.Pipeline)
	}
	log.Printf("Pipeline to be executed: %s", trigger.Pipeline)

	dockerManager, dockerSHA, err := buildManager.PrepareDockerImage(githubClient, event, buildSpec)
	if err != nil {
		log.Printf("Error preparing the docker image. %s", err)
		return err
	}

	containerManager := NewContainerManager(buildManager.Database, dockerManager, buildSpec, pipeline, trigger, event, dockerSHA)
	if err := containerManager.ExecutePipeline(); err != nil {
		log.Printf("Error executing the pipeline. %s", err)
		return err
	}
	return nil
}

// GetSpec to retrieve .gocilla.yml from GitHub repository
func (buildManager *Manager) GetSpec(githubClient *github.Client, event *github.Event) (*Spec, error) {
	content, err := githubClient.GetFileContent(event.Organization, event.Repository, ".gocilla.yml", event.SHA)
	if err != nil {
		return nil, err
	}
	var buildSpec Spec
	err = yaml.Unmarshal(content, &buildSpec)
	return &buildSpec, err
}

// GetTrigger to get the trigger matching the GitHub event from the build spec.
func (buildManager *Manager) GetTrigger(buildSpec *Spec, event *github.Event) *TriggerSpec {
	for _, triggerSpec := range buildSpec.Triggers {
		if triggerSpec.Event == event.Type && triggerSpec.Branch == event.Branch {
			return &triggerSpec
		}
	}
	return nil
}

// GetPipeline to get the pipeline to be executed according to the trigger that matches the GitHub event.
func (buildManager *Manager) GetPipeline(buildSpec *Spec, triggerSpec *TriggerSpec) *PipelineSpec {
	for _, pipelineSpec := range buildSpec.Pipelines {
		if pipelineSpec.Name == triggerSpec.Pipeline {
			return &pipelineSpec
		}
	}
	return nil
}

// PrepareDockerImage to set up the docker image.
func (buildManager *Manager) PrepareDockerImage(githubClient *github.Client, event *github.Event, buildSpec *Spec) (*docker.Manager, string, error) {
	dockerSHA, err := githubClient.GetFileSHA(event.Organization, event.Repository, buildSpec.Docker.File, event.SHA)
	if err != nil {
		return nil, dockerSHA, err
	}
	log.Printf("Dockerfile '%s' with SHA '%s'", buildSpec.Docker.File, dockerSHA)

	dockerManager := buildManager.DockerManagers.Get()
	if !dockerManager.ExistsImage(event.Organization, event.Repository, dockerSHA) {
		dir, err := githubClient.DownloadProjectContent(event.Organization, event.Repository, event.SHA)
		if err != nil || dir == "" {
			log.Println("Error downloading the project")
			return nil, dockerSHA, err
		}
		defer os.RemoveAll(dir)

		dockerfileDir := filepath.Dir(filepath.Join(dir, buildSpec.Docker.File))
		log.Printf("Directory to build the docker image: %s", dockerfileDir)
		err = dockerManager.BuildImage(event.Organization, event.Repository, dockerSHA, dockerfileDir)
		if err != nil {
			log.Println("Error building docker image", err)
			return nil, dockerSHA, err
		}
		log.Println("Dockerfile built successfully")
	} else {
		log.Println("Image already existed")
	}
	return dockerManager, dockerSHA, nil
}
