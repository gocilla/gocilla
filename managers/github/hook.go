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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
)

// ExtendedWebHookPayload type.
type ExtendedWebHookPayload struct {
	BaseRef *string        `json:"base_ref,omitempty"`
	Repo    *FixRepository `json:"repository,omitempty"`
	github.WebHookPayload
}

// FixRepository type.
//     See issue: https://github.com/google/go-github/issues/131
type FixRepository struct {
	Owner    *github.User `json:"owner,omitempty"`
	Name     *string      `json:"name,omitempty"`
	CloneURL *string      `json:"clone_url,omitempty"`
	SSHURL   *string      `json:"ssh_url,omitempty"`
}

// Event type.
type Event struct {
	Type         string
	Branch       string
	Organization string
	Repository   string
	GitURL       string
	SHA          string
	Push         *EventPush
	Pull         *EventPull
}

// EventPull type.
type EventPull struct {
	Number int
}

// EventPush type.
type EventPush struct {
}

// ParsePushEvent to parse a GitHub push event.
// It differentiates when the push corresponds to a tag.
func ParsePushEvent(r *http.Request) (*Event, error) {
	var payload ExtendedWebHookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// Ignore events related to removal of a branch or a tag
	if *payload.Deleted {
		return nil, nil
	}

	event := &Event{
		Organization: *payload.Repo.Owner.Name,
		Repository:   *payload.Repo.Name,
		GitURL:       *payload.Repo.SSHURL,
		SHA:          *payload.HeadCommit.ID,
		Push:         &EventPush{},
	}
	if strings.HasPrefix(*payload.Ref, "refs/tags/") {
		event.Type = "tag"
		if payload.BaseRef != nil {
			event.Branch = (*payload.BaseRef)[len("refs/heads/"):]
		}
	} else {
		event.Type = "push"
		event.Branch = (*payload.Ref)[len("refs/heads/"):]
	}
	return event, nil
}

// ParsePullEvent to parse a pull request event.
func ParsePullEvent(r *http.Request) (*Event, error) {
	var payload github.PullRequestEvent
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// Ignore the event when a pull request is closed
	if *payload.Action == "closed" {
		return nil, nil
	}

	event := &Event{
		Type:         "pull",
		Branch:       *payload.PullRequest.Base.Ref,
		Organization: *payload.PullRequest.Head.Repo.Owner.Login,
		Repository:   *payload.PullRequest.Head.Repo.Name,
		GitURL:       *payload.PullRequest.Head.Repo.GitURL,
		SHA:          fmt.Sprintf("pull/%d/head", *payload.Number),
		Pull:         &EventPull{Number: *payload.Number},
	}
	return event, nil
}

// ParseEvent to parse a GitHub event.
func ParseEvent(r *http.Request) (*Event, error) {
	githubEvent := r.Header.Get("X-GitHub-Event")
	log.Printf("X-GitHub-Event: %s", githubEvent)
	if githubEvent == "push" {
		return ParsePushEvent(r)
	} else if githubEvent == "pull_request" {
		return ParsePullEvent(r)
	} else {
		log.Printf("Invalid X-GitHub-Event header: %s", githubEvent)
		return nil, nil
	}
}
