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

package github

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
)

const pageSize = 1000

// Config type.
type Config struct {
	Events    []string `json:"events"`
	EventsURL string   `json:"eventsUrl"`
}

// Manager type.
// Manager to use GitHub API.
type Manager struct {
	Config *Config
}

// NewManager is the constructor for a GitHug Manager.
func NewManager(config *Config) *Manager {
	return &Manager{config}
}

// Client type.
type Client struct {
	Client     *github.Client
	HTTPClient *http.Client
	Config     *Config
}

// NewClient is the constructor for a GitHub Client.
func (githubManager Manager) NewClient(httpClient *http.Client) *Client {
	client := github.NewClient(httpClient)
	return &Client{client, httpClient, githubManager.Config}
}

// GetUser to get the user profile in GitHub.
func (githubClient Client) GetUser() (user *github.User, err error) {
	user, _, err = githubClient.Client.Users.Get("")
	return
}

// GetOrganizations to retrieve the user's organizations.
func (githubClient Client) GetOrganizations() (organizations []github.Organization, err error) {
	organizations, _, err = githubClient.Client.Organizations.List("", nil)
	return
}

// GetRepositories to retrieve the user's repositories.
func (githubClient Client) GetRepositories() (repositories []github.Repository, err error) {
	listOptions := github.ListOptions{PerPage: pageSize}
	repositoryListOptions := &github.RepositoryListOptions{ListOptions: listOptions}
	repositories, _, err = githubClient.Client.Repositories.List("", repositoryListOptions)
	return
}

// CreateHook to create a hook on a repository.
func (githubClient Client) CreateHook(owner, repo string) (hookID *int, err error) {
	hookName := "web"
	hookConfig := &github.Hook{
		Name:   &hookName,
		Events: githubClient.Config.Events,
		Config: map[string]interface{}{
			"url":          githubClient.Config.EventsURL,
			"content_type": "json",
		},
	}
	h, _, error := githubClient.Client.Repositories.CreateHook(owner, repo, hookConfig)
	if error != nil {
		log.Println("Error creating hook", error)
		return nil, error
	}
	return h.ID, nil
}

// DeleteHook to remove a hook on a repository.
func (githubClient Client) DeleteHook(owner, repo string, hookID int) error {
	_, error := githubClient.Client.Repositories.DeleteHook(owner, repo, hookID)
	if error != nil {
		log.Println("Error deleting hook", error)
		return error
	}
	return nil
}

// CreateStatus creates a new status for a repository at the specified reference.
// The reference can be a SHA, a branch name, or a tag name.
func (githubClient Client) CreateStatus(owner, repo, ref, context, description, state string) error {
	status := &github.RepoStatus{Context: &context, Description: &description, State: &state}
	_, _, err := githubClient.Client.Repositories.CreateStatus(owner, repo, ref, status)
	if err != nil {
		log.Printf("Error creating status. %s", err)
	}
	return err
}

// GetFileContent to download a file from a user's repository.
func (githubClient Client) GetFileContent(owner, repo, path, ref string) ([]byte, error) {
	options := &github.RepositoryContentGetOptions{Ref: ref}
	fileContent, _, _, err := githubClient.Client.Repositories.GetContents(owner, repo, path, options)
	if err != nil {
		return nil, err
	}
	log.Printf("SHA: %s", *fileContent.SHA)
	decodedFileContent, _ := fileContent.Decode()
	return decodedFileContent, nil
}

// GetFileSHA to get the file SHA.
func (githubClient Client) GetFileSHA(owner, repo, path, ref string) (string, error) {
	log.Printf("GetFileSHA %s %s", path, ref)
	options := &github.RepositoryContentGetOptions{Ref: ref}
	fileContent, _, _, err := githubClient.Client.Repositories.GetContents(owner, repo, path, options)
	if err != nil {
		return "", err
	}
	return *fileContent.SHA, nil
}

// DownloadProjectContent to download a whole repository with a specific reference (SHA).
func (githubClient Client) DownloadProjectContent(owner, repo, ref string) (string, error) {
	options := &github.RepositoryContentGetOptions{Ref: ref}
	url, _, err := githubClient.Client.Repositories.GetArchiveLink(owner, repo, github.Tarball, options)
	if err != nil {
		log.Println("Error in DownloadProjectContent")
		return "", err
	}
	resp, err := githubClient.HTTPClient.Get(url.String())
	if err != nil {
		log.Println("Error getting the project tar.gz")
		return "", err
	}
	defer resp.Body.Close()

	dir, err := ioutil.TempDir("", "gocilla")
	if err != nil {
		log.Println("Error creating temporary directory")
		return "", err
	}
	log.Printf("Temporary directory: %s", dir)

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Println("Error getting gzip reader")
		return "", err
	}

	tarReader := tar.NewReader(gzipReader)

	var baseDir string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error next tarReader")
			return "", err
		}

		info := header.FileInfo()
		if baseDir == "" {
			if info.IsDir() {
				baseDir = header.Name
			} else {
				continue
			}
		}

		relativePath, err := filepath.Rel(baseDir, header.Name)
		if err != nil {
			log.Println("Error getting relativePath")
			return "", err
		}

		path := filepath.Join(dir, relativePath)
		log.Printf("New file: %s", path)
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				log.Println("Error making dir")
				return "", err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			log.Println("Error creating file")
			return "", err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			log.Println("Error copying file")
			return "", err
		}
	}
	return dir, nil
}
