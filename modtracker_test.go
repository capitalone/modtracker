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

package modtracker

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

var sampleUnmarshaler Unmarshaler

func (s *Sample) UnmarshalJSON(data []byte) error {
	var err error
	s.modified, err = sampleUnmarshaler(data, s)
	return err
}

func (s *Sample) GetModified() []string {
	return s.modified
}

var tests = []string{
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

func BenchmarkUnmarshalJSON(b *testing.B) {
	sampleUnmarshaler = UnmarshalJSON
	results := make([]Sample, len(tests))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range tests {
			var s Sample
			json.Unmarshal([]byte(v), &s)
			results[k] = s
		}
	}
}

func BenchmarkBuildJSONUnmarshaler(b *testing.B) {
	sampleUnmarshaler, _ = BuildJSONUnmarshaler((*Sample)(nil))
	results := make([]Sample, len(tests))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range tests {
			var s Sample
			json.Unmarshal([]byte(v), &s)
			results[k] = s
		}
	}
}

type Sample2 struct {
	FirstName *string
	LastName  *string
	Age       *int
	Inner     *struct {
		Address string
	}
	Pet     string
	Company string `json:"company"`
}

func BenchmarkStandardJSONUnmarshal(b *testing.B) {
	results := make([]Sample2, len(tests))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range tests {
			var s Sample2
			json.Unmarshal([]byte(v), &s)
			results[k] = s
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type TSample struct {
		FirstName  *string `json:"firstName"`
		MiddleName *string `json:"middleName"`
		LastName   *string `json:"lastName"`
		Age        *int    `json:"age"`
	}

	data := `
	 {
  "firstName": "John",
  "middleName": null,
  "lastName": "Doe",
  "age": null
}
	`
	var ts TSample
	modified, err := UnmarshalJSON([]byte(data), &ts)
	assert.Nil(t, err, "Err")
	assert.Nil(t, ts.MiddleName, "Middle name")
	assert.Nil(t, ts.Age, "Age")
	assert.Equal(t, 4, len(modified))
	assert.Equal(t, *ts.FirstName, "John")
	assert.Equal(t, *ts.LastName, "Doe")
}

func TestUnmarshalJSONAllTypes(t *testing.T) {
	type Inner struct {
		F1 string
		F2 int
		F3 *int
	}
	type TSample struct {
		FirstName  string                 `json:"firstName"`
		MiddleName *string                `json:"middleName"`
		LastName   *string                `json:"lastName"`
		Age        int                    `json:"age"`
		Age2       *int                   `json:"age2"`
		Age3       *int                   `json:"age3"`
		F1         float64                `json:"f1"`
		F2         *float64               `json:"f2"`
		F3         *float64               `json:"f3"`
		B1         bool                   `json:"b1"`
		B2         *bool                  `json:"b2"`
		B3         *bool                  `json:"b3"`
		U1         uint                   `json:"u1"`
		U2         *uint                  `json:"u2"`
		U3         *uint                  `json:"u3"`
		S1         []string               `json:"s1"`
		S2         []string               `json:"s2"`
		M1         map[string]interface{} `json:"m1"`
		M2         map[string]interface{} `json:"m2"`
		O1         Inner                  `json:"o1"`
		O2         *Inner                 `json:"o2"`
		O3         *Inner                 `json:"o3"`
	}

	data := `
	 {
  "firstName": "John",
  "middleName": null,
  "lastName": "Doe",
  "age": 10,
  "age2": null,
  "age3": 20,
  "f1": 3.5,
  "f2": null,
  "f3": 6.323,
  "b1": true,
  "b2": null,
  "b3": false,
  "u1": 32400,
  "u2": null,
  "u3": 124325,
  "s1": ["a","b","c"],
  "s2": null,
  "m1": {
  	"a": 1,
  	"b": "asdf"
  },
  "m2": null,
  "o1": {
  	"F1": "asdf",
  	"F2": 234,
  	"F3": 123
  },
  "o2": null,
  "o3": {
  	"F1": "cvcd",
  	"F2": 61,
  	"F3": 897
  }
}
	`
	var ts TSample
	modified, err := UnmarshalJSON([]byte(data), &ts)
	assert.Nil(t, err, "Err")
	assert.Nil(t, ts.MiddleName, "Middle name")
	assert.Nil(t, ts.Age2, "Age")
	assert.Nil(t, ts.F2)
	assert.Nil(t, ts.B2)
	assert.Nil(t, ts.U2)
	assert.Nil(t, ts.S2)
	assert.Nil(t, ts.M2)
	assert.Nil(t, ts.O2)
	assert.Equal(t, 22, len(modified))
	assert.Equal(t, ts.FirstName, "John")
	assert.Equal(t, *ts.LastName, "Doe")
	assert.Equal(t, ts.Age, 10)
	assert.Equal(t, *ts.Age3, 20)
	assert.Equal(t, ts.F1, 3.5)
	assert.Equal(t, *ts.F3, 6.323)
	assert.Equal(t, ts.B1, true)
	assert.Equal(t, *ts.B3, false)
	assert.Equal(t, ts.U1, uint(32400))
	assert.Equal(t, *ts.U3, uint(124325))
	assert.Equal(t, 3, len(ts.S1))
	assert.Equal(t, "a", ts.S1[0])
	assert.Equal(t, "b", ts.S1[1])
	assert.Equal(t, "c", ts.S1[2])
	assert.Equal(t, 2, len(ts.M1))
	assert.Equal(t, float64(1), ts.M1["a"])
	assert.Equal(t, "asdf", ts.M1["b"])
	assert.Equal(t, "asdf", ts.O1.F1)
	assert.Equal(t, 234, ts.O1.F2)
	assert.Equal(t, 123, *ts.O1.F3)
	assert.Equal(t, "cvcd", ts.O3.F1)
	assert.Equal(t, 61, ts.O3.F2)
	assert.Equal(t, 897, *ts.O3.F3)
}

func TestUnmarshalJSONInvalid(t *testing.T) {
	type TSample struct {
		FirstName   *string `json:"firstName"`
		MiddleName  *string `json:"middleName"`
		LastName    *string `json:"lastName"`
		Age         *int    `json:"age"`
		FavoriteNum int     `json:"fave"`
	}

	data := `
	 {
  "firstName": true,
  "middleName": 10,
  "lastName": "Doe",
  "age": 24.3,
  "fave": null
}
	`
	var ts TSample
	modified, err := UnmarshalJSON([]byte(data), &ts)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(modified))
	assert.Equal(t, "Doe", *ts.LastName)
	assert.Equal(t, 0, *ts.Age)
	assert.Equal(t, 0, ts.FavoriteNum)
	assert.Nil(t, ts.FirstName)
	assert.Nil(t, ts.MiddleName)
	fmt.Printf("%v\n", err)
	fmt.Printf("%+v\n", err)
}

func TestCustomJSONSerialilzerString(t *testing.T) {
	type TimeWrapper struct {
		T  *time.Time
		T2 time.Time
		T3 *time.Time
	}

	data := `
	{
		"T": "2009-11-10T23:00:00Z",
		"T2": "2009-11-10T23:00:00Z",
		"T3": null
	}
	`
	var ts TimeWrapper
	modified, err := UnmarshalJSON([]byte(data), &ts)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(modified))
	assert.Nil(t, ts.T3)
	var st time.Time
	json.Unmarshal([]byte(`"2009-11-10T23:00:00Z"`), &st)
	assert.Equal(t, st, *ts.T)
	assert.Equal(t, st, ts.T2)
}
