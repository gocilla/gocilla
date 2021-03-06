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

import (
	"log"

	"gopkg.in/mgo.v2"
)

// Config type.
type Config struct {
	URL string
}

// Database type.
type Database struct {
	Session *mgo.Session
}

// NewDatabase is the constructor for Database.
func NewDatabase(mongodbConfig *Config) (*Database, error) {
	log.Println("Connecting to ", mongodbConfig.URL)
	session, err := mgo.Dial(mongodbConfig.URL)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	return &Database{session}, nil
}

// Close to close the mongodb session.
func (database *Database) Close() {
	database.Session.Close()
}
