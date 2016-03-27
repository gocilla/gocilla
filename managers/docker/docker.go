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

package docker

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

// Manager type.
// Manager to create and destroy the docker containers that execute the builds.
type Manager struct {
	Client *docker.Client
}

// ContainerManager type.
type ContainerManager struct {
	Client    *docker.Client
	Container *docker.Container
}

// Config type.
type Config struct {
	Host      string
	CertPath  string
	TLSVerify bool
}

// GetImageName get the docker image corresponding to a repository.
// The image name is composed as {organization}/{repository} in lower case.
func GetImageName(organization, repository string) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(organization), strings.ToLower(repository))
}

// GetTagName gets the image tag corresponding to the first 8 characters of the repository SHA.
func GetTagName(sha string) string {
	return sha[0:7]
}

// GetTaggedImageName gets the tagged image name.
//    Note that the tag corresponds to the SHA of Dockerfile, not to the repository SHA.
//    The reason is to only generate a docker image when there is a change in Dockerfile.
func GetTaggedImageName(organization, repository, sha string) string {
	return fmt.Sprintf("%s:%s", GetImageName(organization, repository), GetTagName(sha))
}

// GetEnv get an array of environment variables where each array element corresponds
// to an envVars object with format: {key}={value}.
func GetEnv(envVars map[string]string) []string {
	var env []string
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// NewManager is the constructor for Manager.
func NewManager(dockerConfig *Config) *Manager {
	ca := fmt.Sprintf("%s/ca.pem", dockerConfig.CertPath)
	cert := fmt.Sprintf("%s/cert.pem", dockerConfig.CertPath)
	key := fmt.Sprintf("%s/key.pem", dockerConfig.CertPath)
	client, _ := docker.NewTLSClient(dockerConfig.Host, cert, key, ca)
	return &Manager{client}
}

// BuildImage to build a docker image.
func (dockerManager *Manager) BuildImage(organization, repository, sha, directory string, w io.Writer) error {
	imageName := GetImageName(organization, repository)

	buildImageOptions := docker.BuildImageOptions{
		Name:         imageName,
		ContextDir:   directory,
		OutputStream: w,
	}
	err := dockerManager.Client.BuildImage(buildImageOptions)
	if err != nil {
		log.Println("Error building the image", err)
		return err
	}

	tagImageOptions := docker.TagImageOptions{
		Repo: imageName,
		Tag:  GetTagName(sha),
	}
	err = dockerManager.Client.TagImage(imageName, tagImageOptions)
	if err != nil {
		log.Println("Error tagging the image", err)
		return err
	}
	return nil
}

// ExistsImage to check if the image already exists (using GetTaggedImageName method).
func (dockerManager *Manager) ExistsImage(organization, repository, sha string) bool {
	imageName := GetTaggedImageName(organization, repository, sha)
	image, _ := dockerManager.Client.InspectImage(imageName)
	return image != nil
}

// CreateAndStartContainer creates and starts a docker container.
func (dockerManager *Manager) CreateAndStartContainer(organization, repository, sha, user, workingDir string, envVars map[string]string) (*ContainerManager, error) {
	imageName := GetTaggedImageName(organization, repository, sha)
	log.Printf("CreateAndStartContainer for image: %s", imageName)
	log.Printf("WorkingDir: %s", workingDir)
	containerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{Image: imageName, Env: GetEnv(envVars), User: user, WorkingDir: workingDir, Memory: 1024000000},
	}
	container, err := dockerManager.Client.CreateContainer(containerOptions)
	if err != nil {
		log.Println("Error creating container with image", imageName)
		return nil, err
	}
	err = dockerManager.Client.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		log.Println("Error starting container with image", imageName)
		return nil, err
	}
	return &ContainerManager{dockerManager.Client, container}, nil
}

// RemoveContainer removes a docker container.
func (containerManager *ContainerManager) RemoveContainer() error {
	removeContainerOptions := docker.RemoveContainerOptions{
		ID:    containerManager.Container.ID,
		Force: true,
	}
	return containerManager.Client.RemoveContainer(removeContainerOptions)
}

// ExecContainer executes a command on a running docker container.
func (containerManager *ContainerManager) ExecContainer(command string, w io.Writer) error {
	execOptions := docker.CreateExecOptions{
		Container:    containerManager.Container.ID,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          []string{"sh", "-c", command},
	}
	exec, err := containerManager.Client.CreateExec(execOptions)
	if err != nil {
		log.Println("Error creating the execution of command", command)
		return err
	}

	startExecOptions := docker.StartExecOptions{
		Detach:       false,
		OutputStream: w,
		ErrorStream:  w,
	}
	err = containerManager.Client.StartExec(exec.ID, startExecOptions)
	if err != nil {
		log.Println("Error starting the execution of command", command)
		return err
	}
	inspect, err := containerManager.Client.InspectExec(exec.ID)
	if err != nil {
		log.Printf("Error inspecting the execution of command '%s'", command)
	}
	if inspect.ExitCode != 0 {
		log.Println("Invalid exit code")
		err = fmt.Errorf("Invalid exit code: %d", inspect.ExitCode)
	}

	return err
}
