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

import "fmt"

// Repository type.
type Repository struct {
	OrgID   string           `bson:"organization" json:"orgId"`
	RepoID  string           `bson:"repository" json:"repoId"`
	EnvVars []PipelineEnvVar `bson:"envVars" json:"envVars"`
}

// PipelineEnvVar type.
type PipelineEnvVar struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}

// GetRepository to get a repository (settings).
func (database *Database) GetRepository(orgID, repoID string) (*Repository, error) {
	var repository Repository
	ID := fmt.Sprintf("%s/%s", orgID, repoID)
	collection := database.Session.DB("").C("repositories")
	err := collection.FindId(ID).One(&repository)
	if err != nil && err.Error() == "not found" {
		return &Repository{OrgID: orgID, RepoID: repoID}, nil
	}
	return &repository, err
}

// UpdateRepository to update a repository (settings).
func (database *Database) UpdateRepository(repository *Repository) error {
	ID := fmt.Sprintf("%s/%s", repository.OrgID, repository.RepoID)
	collection := database.Session.DB("").C("repositories")
	_, err := collection.UpsertId(ID, repository)
	return err
}
