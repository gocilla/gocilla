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
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/gocilla/gocilla/managers/docker"
	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gocilla/gocilla/managers/oauth2"
)

type BuildManager struct {
	Database       *mongodb.Database
	OAuth2Manager  *oauth2.OAuth2Manager
	GitHubManager  *github.GitHubManager
	DockerManagers docker.DockerManagers
}

type BuildSpec struct {
	Docker    BuildDockerSpec
	Jobs      map[string]string
	Pipelines []BuildPipelineSpec
}

type BuildDockerSpec struct {
	File       string
	User       string
	WorkingDir string `json:"workingDir" yaml:"workingDir"`
}

type BuildPipelineSpec struct {
	Name string
	Jobs []string
}

func NewBuildManager(database *mongodb.Database, oauth2Manager *oauth2.OAuth2Manager, githubManager *github.GitHubManager, dockerManagers docker.DockerManagers) *BuildManager {
	return &BuildManager{database, oauth2Manager, githubManager, dockerManagers}
}

func (buildManager *BuildManager) Build(event *github.Event) error {
	log.Printf("Starting build process for event: %+v", event)

	trigger, err := buildManager.Database.GetTrigger(event.Organization, event.Repository, event.Type, event.Branch)
	if err != nil {
		log.Printf("Error getting trigger. %s", err)
		return err
	}
	log.Printf("Trigger: %+v", trigger)

	hook, err := buildManager.Database.GetHook(event.Organization, event.Repository)
	if err != nil {
		log.Printf("Error getting hook. %s", err)
		return err
	}
	log.Printf("Using hook: %+v", hook)

	githubClient := buildManager.GitHubManager.NewGitHubClient(buildManager.OAuth2Manager.GetClientFromAccessToken(hook.AccessToken))
	buildSpec, err := buildManager.GetBuildSpec(githubClient, event)
	if err != nil {
		log.Printf("Error getting the pipeline. %s", err)
		return err
	}

	pipeline := buildManager.GetPipeline(buildSpec, trigger)
	if pipeline == nil {
		log.Printf("No pipeline matching the trigger pipeline: %s", trigger.Pipeline)
		return err
	}

	dockerManager, dockerSHA, err := buildManager.PrepareDockerImage(githubClient, event, buildSpec)
	if err != nil {
		log.Printf("Error preparing the docker image. %s", err)
		return err
	}

	containerBuildManager := NewContainerBuildManager(buildManager.Database, dockerManager, buildSpec, pipeline, trigger, event, dockerSHA)
	if err := containerBuildManager.ExecutePipeline(); err != nil {
		log.Printf("Error executing the pipeline. %s", err)
		return err
	}
	return nil
}

func (buildManager *BuildManager) GetBuildSpec(githubClient *github.GitHubClient, event *github.Event) (*BuildSpec, error) {
	content, err := githubClient.GetFileContent(event.Organization, event.Repository, ".gocilla.yml", event.SHA)
	if err != nil {
		return nil, err
	}
	var buildSpec BuildSpec
	err = yaml.Unmarshal(content, &buildSpec)
	return &buildSpec, err
}

func (buildManager *BuildManager) GetPipeline(buildSpec *BuildSpec, trigger *mongodb.Trigger) *BuildPipelineSpec {
	for _, buildPipelineSpec := range buildSpec.Pipelines {
		if buildPipelineSpec.Name == trigger.Pipeline {
			return &buildPipelineSpec
		}
	}
	return nil
}

func (buildManager *BuildManager) PrepareDockerImage(githubClient *github.GitHubClient, event *github.Event, buildSpec *BuildSpec) (*docker.DockerManager, string, error) {
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
