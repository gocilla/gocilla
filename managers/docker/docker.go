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
	"bufio"
	_"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type DockerManager struct {
	Client *docker.Client
}

type ContainerManager struct {
	Client		*docker.Client
	Container	*docker.Container
}

type DockerConfig struct {
	Host		string
	CertPath	string
	TLSVerify	bool
}

func GetImageName(organization, repository string) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(organization), strings.ToLower(repository))
}

func GetTagName(sha string) string {
	return sha[0:7]
}

func GetTaggedImageName(organization, repository, sha string) string {
	return fmt.Sprintf("%s:%s", GetImageName(organization, repository), GetTagName(sha))
}

func NewDockerManager(dockerConfig *DockerConfig) *DockerManager {
	ca := fmt.Sprintf("%s/ca.pem", dockerConfig.CertPath)
    cert := fmt.Sprintf("%s/cert.pem", dockerConfig.CertPath)
    key := fmt.Sprintf("%s/key.pem", dockerConfig.CertPath)
	client, _ := docker.NewTLSClient(dockerConfig.Host, cert, key, ca)
	return &DockerManager{client}
}

func (dockerManager *DockerManager) BuildImage(organization, repository, sha, directory string) error {
	imageName := GetImageName(organization, repository)

    r, w := io.Pipe()
    go func(reader io.Reader) {
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
        	log.Println(scanner.Text())
        }
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(os.Stderr, "There was an error with the scanner in attached container", err)
        }
    }(r)

	//var buf bytes.Buffer
	buildImageOptions := docker.BuildImageOptions{
		Name: imageName,
		ContextDir: directory,
		OutputStream: w,
	}
	err := dockerManager.Client.BuildImage(buildImageOptions)
	if err != nil {
		log.Println("Error building the image", err)
		return err
	}

	tagImageOptions := docker.TagImageOptions{
		Repo: imageName,
		Tag: GetTagName(sha),
	}
	err = dockerManager.Client.TagImage(imageName, tagImageOptions)
	if err != nil {
		log.Println("Error tagging the image", err)
		return err
	}
	return nil
}

func (dockerManager *DockerManager) ExistsImage(organization, repository, sha string) bool {
	imageName := GetTaggedImageName(organization, repository, sha)
	image, _ := dockerManager.Client.InspectImage(imageName)
	if image == nil {
		return false
	} else {
		return true
	}
}

func (dockerManager *DockerManager) CreateAndStartContainer(organization, repository, sha, user, workingDir string, env []string) (*ContainerManager, error) {
	imageName := GetTaggedImageName(organization, repository, sha)
	log.Printf("CreateAndStartContainer for image: %s", imageName)
	log.Printf("WorkingDir: %s", workingDir)
	containerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{Image: imageName, Env: env, User: user, WorkingDir: workingDir, Memory: 1024000000},
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

func (containerManager *ContainerManager) RemoveContainer() error {
	removeContainerOptions := docker.RemoveContainerOptions{
		ID: containerManager.Container.ID,
		Force: true,
	}
	return containerManager.Client.RemoveContainer(removeContainerOptions)
}

func (containerManager *ContainerManager) ExecContainer(command string) (string, error) {
    r, w := io.Pipe()
    go func(reader io.Reader) {
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
        	log.Println(scanner.Text())
        }
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(os.Stderr, "There was an error with the scanner in attached container", err)
        }
    }(r)

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
		return "", err
	}

	//var buffer bytes.Buffer
	startExecOptions := docker.StartExecOptions{
		Detach: false,
        //OutputStream: &buffer,
        //ErrorStream:  &buffer,
        OutputStream: w,
        ErrorStream:  w,
	}
	err = containerManager.Client.StartExec(exec.ID, startExecOptions)
	if err != nil {
		log.Println("Error starting the execution of command", command)
		return "", err
	}
	inspect, err := containerManager.Client.InspectExec(exec.ID)
	if err != nil {
		log.Printf("Error inspecting the execution of command '%s'", command)
	}
	if inspect.ExitCode != 0 {
		log.Println("Invalid exit code")
		err = errors.New(fmt.Sprintf("Invalid exit code: %d", inspect.ExitCode))
	}

	//return buffer.String(), err
	return "", err
}

/*
func (dockerManager *DockerManager) ListImages() {
    r, w := io.Pipe()

    go func(reader io.Reader) {
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
        	log.Println(scanner.Text())
        }
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(os.Stderr, "There was an error with the scanner in attached container", err)
        }
    }(r)

	execOptions := docker.CreateExecOptions{
		Container:    "calandracas",
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          []string{"ls" "-al" "/tmp"},
	}
	execObj, error2 := dockerManager.Client.CreateExec(execOptions)
	log.Println(execObj)
	log.Println(error2)

	startExecOptions := docker.StartExecOptions{
		Detach: false,
        OutputStream: w,
        ErrorStream:  w,
	}
	error3 := dockerManager.Client.StartExec(execObj.ID, startExecOptions)
	log.Println(error3)
}
*/
