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

import "gopkg.in/mgo.v2"

// CreateFile to create a file in mongodb with GridFS.
func (database *Database) CreateFile(filename string) (file *mgo.GridFile, err error) {
	return database.Session.DB("").GridFS("fs").Create(filename)
}

// OpenFile to create a file in mongodb with GridFS.
func (database *Database) OpenFile(filename string) (file *mgo.GridFile, err error) {
	return database.Session.DB("").GridFS("fs").Open(filename)
}
