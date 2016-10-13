//Copyright 2016 Capital One Services, LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and limitations under the License. 

package main

import (
	"encoding/json"
	"fmt"
	"github.com/capitalone/modtracker"
)

type Sample struct {
	FirstName *string
	LastName  *string
	Age       *int
	Inner     *struct {
		Address string
	}
	Pet      string
	Company  string `json:"company"`
	modified []string
}

var sampleUnmarshaler modtracker.Unmarshaler

func init() {
	var err error
	sampleUnmarshaler, err = modtracker.BuildJSONUnmarshaler((*Sample)(nil))
	if err != nil {
		panic(err)
	}
}

func (s *Sample) UnmarshalJSON(data []byte) error {
	var err error
	s.modified, err = sampleUnmarshaler(data, s)
	return err
}

func (s *Sample) GetModified() []string {
	return s.modified
}

func (s Sample) String() string {
	fn := ""
	if s.FirstName != nil {
		fn = *s.FirstName
	}
	ln := ""
	if s.LastName != nil {
		ln = *s.LastName
	}
	a := 0
	if s.Age != nil {
		a = *s.Age
	}
	return fmt.Sprintf("%s %s %d %v %s %s modified:%v", fn, ln, a, s.Inner, s.Pet, s.Company, s.modified)
}

func printIt(data string) {
	var s Sample
	fmt.Println("converting", data)
	err := json.Unmarshal([]byte(data), &s)
	if err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(s)
	}
}

func main() {
	tests := []string{
		`
	{
		"FirstName": "Homer",
		"LastName": "Simpson",
		"Age": 37
	}`,
		`
	{
		"FirstName": "Homer",
		"Age": 37
	}`,
		`
	{
		"FirstName": null,
		"Age": 37
	}`, `
	{
		"Age": 37
	}`,
		`
	{
		"Age": 37,
		"Inner": {
			"Address": "742 Evergreen Terr."
		}
	}`,
		`
	{
		"Age": 37,
		"Pet": "Spider-Pig"
	}`,
		`
	{
		"Age": 37,
		"Pet": "Spider-Pig",
		"company": "Springfield Nuclear Power Plant"
	}`,
	}

	for _, v := range tests {
		printIt(v)
	}
}
