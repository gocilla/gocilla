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

package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gocilla/gocilla/managers/docker"
	"github.com/gocilla/gocilla/managers/github"
	"github.com/gocilla/gocilla/managers/mongodb"
	"github.com/gocilla/gocilla/managers/oauth2"
	"github.com/gocilla/gocilla/managers/session"
)

// Config type.
type Config struct {
	Port    uint16
	OAuth2  *oauth2.Config
	GitHub  *github.Config
	Session *session.Config
	Mongodb *mongodb.Config
	Docker  *docker.ClusterConfig
}

// Decode the JSON configuration stored in a file path.
func Decode(path string) (config Config, err error) {
	buf, error := ioutil.ReadFile(path)
	if error != nil {
		err = error
		return
	}
	err = json.Unmarshal(buf, &config)
	return
}
