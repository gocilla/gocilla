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
	"os"
	"path/filepath"
	"net/http"

	"github.com/google/go-github/github"
)

type GitHubConfig struct {
	Events 		[]string `json:"events"`
	EventsURL 	string   `json:"eventsUrl"`
}

type GitHubManager struct {
	Config *GitHubConfig
}

func NewGitHubManager(config *GitHubConfig) *GitHubManager {
	return &GitHubManager{config}
}

func (githubManager GitHubManager) NewGitHubClient(httpClient *http.Client) *GitHubClient {
	client := github.NewClient(httpClient)
	return &GitHubClient{client, httpClient, githubManager.Config}
}

type GitHubClient struct {
	Client *github.Client
	HttpClient *http.Client
	Config *GitHubConfig
}

func (githubClient GitHubClient) GetUser() (user *github.User, err error) {
	user, _, err = githubClient.Client.Users.Get("")
	return
}

func (githubClient GitHubClient) GetOrganizations() (organizations []github.Organization, err error) {
	organizations, _, err = githubClient.Client.Organizations.List("", nil)
	return
}

func (githubClient GitHubClient) GetRepositories() (repositories []github.Repository, err error) {
	repositories, _, err = githubClient.Client.Repositories.List("", nil)
	return
}
/*
func (githubClient GitHubClient) GetRepositories(organization string) (repositories []github.Repository, err error) {
	repositories, _, err = githubClient.Client.Repositories.List(organization, nil)
	return
}*/

func (githubClient GitHubClient) CreateHook(owner, repo string) (hookId *int, err error) {
	hookName := "web"
	hookConfig := &github.Hook{
		Name: &hookName,
		Events: githubClient.Config.Events,
		Config: map[string]interface{}{
			"url": githubClient.Config.EventsURL,
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

func (githubClient GitHubClient) DeleteHook(owner, repo string, hookId int) error {
	_, error := githubClient.Client.Repositories.DeleteHook(owner, repo, hookId)
	if error != nil {
		log.Println("Error deleting hook", error)
		return error
	}
	return nil
}

func (githubClient GitHubClient) GetFileContent(owner, repo, path, ref string) ([]byte, error) {
	options := &github.RepositoryContentGetOptions{ref}
	fileContent, _, _, err := githubClient.Client.Repositories.GetContents(owner, repo, path, options)
	if err != nil {
		return nil, err
	}
	log.Printf("SHA: %s", *fileContent.SHA)
	decodedFileContent, _ := fileContent.Decode()
	return decodedFileContent, nil
}

func (githubClient GitHubClient) GetFileSHA(owner, repo, path, ref string) (string, error) {
	log.Printf("GetFileSHA %s %s", path, ref)
	options := &github.RepositoryContentGetOptions{ref}
	fileContent, _, _, err := githubClient.Client.Repositories.GetContents(owner, repo, path, options)
	if err != nil {
		return "", err
	}
	return *fileContent.SHA, nil
}

func (githubClient GitHubClient) DownloadProjectContent(owner, repo, ref string) (string, error) {
	options := &github.RepositoryContentGetOptions{ref}
	url, _, err := githubClient.Client.Repositories.GetArchiveLink(owner, repo, github.Tarball, options)
	if err != nil {
		log.Println("Error in DownloadProjectContent")
		return "", err
	}
	resp, err := githubClient.HttpClient.Get(url.String())
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

/*
func (githubClient GitHubClient) DownloadDirectoryContent(owner, repo, path, ref string) (io.ReadCloser, error) {
	options := &github.RepositoryContentGetOptions{ref}
	return githubClient.Client.Repositories.GetContents(owner, repo, path, options)
}

*/