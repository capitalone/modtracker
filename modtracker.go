//Copyright 2016 Capital One Services, LLC
//
// SPDX-License-Identifier: Apache-2.0
// SPDX-Copyright: Copyright (c) Capital One Services, LLC
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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strings"
)

// Modifiable is implemented by struct types that contain a list of their fields that were populated from JSON.
// If a value for a field, even null, was provided in the JSON, the name of the field appears in the slice of strings.
type Modifiable interface {
	GetModified() []string
}

// An Unmarshaler takes in JSON in the first parameter, a pointer to a struct in the second parameter, populates the
// struct with the JSON and returns the modified fields as a slice of strings. In case of error, the struct might be
// partially populated. If there is an error, the modified field slice will be nil.
type Unmarshaler func([]byte, interface{}) ([]string, error)

// UnmarshalJSON provides the default implementation of the Unmarshaler type. It will rediscover the fields in the structure
// each time it is called; to improve performance, use BuildJSONUnmarshaler to create an Unmarshaler instance with the
// struct fields pre-calculated.
func UnmarshalJSON(data []byte, s interface{}) ([]string, error) {
	fm, err := buildJSONFieldMap(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failure during UnmarshalJSON")
	}

	return unmarshalJSONInner(fm, data, s)
}

// BuildJSONUnmarshaler generates a custom implementation of the Unmarshaler type for the type of the provided struct.
// The preferred way to use BuildJSONUnmarshaler is to create a package-level variable and assign it in init with a
// nil instance of the type:
//
// 	type Sample struct {
//		FirstName *string
//		LastName  *string
//		Age       *int
//		Inner     *struct {
//			Address string
//		}
//		Pet      string
//		Company  string `json:"company"`
//		modified []string
//	}
//
//	var sampleUnmarshaler modtracker.Unmarshaler
//
//	func init() {
//		var err error
//		sampleUnmarshaler, err = modtracker.BuildJSONUnmarshaler((*Sample)(nil))
//		if err !=  nil {
//			panic(err)
//		}
//	}
//
//	func (s *Sample) UnmarshalJSON(data []byte) error {
//		modified, err := sampleUnmarshaler(data, s)
//		if err != nil {
//			return err
//		}
//		s.modified = modified
//		return nil
//	}
//
func BuildJSONUnmarshaler(s interface{}) (func([]byte, interface{}) ([]string, error), error) {
	fm, err := buildJSONFieldMap(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failure during UnmarshalJSON")
	}

	return func(data []byte, s interface{}) ([]string, error) {
		return unmarshalJSONInner(fm, data, s)
	}, nil
}

type errorList []error

func (el errorList) innerErr(verb rune, plusFlag bool) string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%d Errors found:\n", len(el)))
	for _, v := range el {
		switch verb {
		case 'v':
			if plusFlag {
				b.WriteString(fmt.Sprintf("%+v\n", v))
			} else {
				b.WriteString(fmt.Sprintf("%v\n", v))
			}
		case 's':
			b.WriteString(fmt.Sprintf("%s\n", v))
		}
	}
	return b.String()
}

func (el errorList) Error() string {
	return el.innerErr('s', false)
}

func (el errorList) Format(s fmt.State, verb rune) {
	msg := el.innerErr(verb, s.Flag('+'))
	io.WriteString(s, msg)
}

func validateType(nt reflect.Type, typeKind reflect.Kind, n string, validKind reflect.Kind, jsonType string) error {
	if typeKind != validKind {
		return errors.Errorf("Invalid type in JSON, expected %s for field %s, got %s", nt, n, jsonType)
	}
	return nil
}

var (
	unmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
)

func unmarshalJSONInner(fm fieldMap, data []byte, s interface{}) ([]string, error) {
	modified := make([]string, 0, len(fm.names))
	var el errorList
	se := reflect.ValueOf(s).Elem()
	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		var fv reflect.Value
		fValue := fm.values[idx]
		t := fValue.t
		n := fValue.name
		fv = reflect.New(fValue.internalType)
		switch vt {
		case jsonparser.String:
			if fValue.unmarshaler {
				b := make([]byte, len(value)+2)
				b[0] = 34
				b[len(b)-1] = 34
				copy(b[1:], value)
				err = json.Unmarshal(b, fv.Interface())
				if err != nil {
					el = append(el, errors.Wrap(err, "JSON unmarshaling"))
					return
				}
			} else {
				err := validateType(fValue.internalType, fValue.internalKind, n, reflect.String, "String")
				if err != nil {
					el = append(el, err)
					return
				}
				s, _ := jsonparser.ParseString(value)
				fv.Elem().SetString(s)
			}
		case jsonparser.Number:
			switch {
			case fValue.intType:
				i, _ := jsonparser.ParseInt(value)
				fv.Elem().SetInt(i)
			case fValue.uintType:
				i, _ := jsonparser.ParseInt(value)
				fv.Elem().SetUint(uint64(i))
			case fValue.floatType:
				f, _ := jsonparser.ParseFloat(value)
				fv.Elem().SetFloat(f)
			default:
				el = append(el, errors.Errorf("Invalid type in JSON, expected %s for field %s, got Number", fValue.internalType, n))
				return
			}
		case jsonparser.Object, jsonparser.Array:
			err = json.Unmarshal(value, fv.Interface())
			if err != nil {
				el = append(el, errors.Wrap(err, "JSON unmarshaling"))
				return
			}
		case jsonparser.Boolean:
			err := validateType(fValue.internalType, fValue.internalKind, n, reflect.Bool, "Boolean")
			if err != nil {
				el = append(el, err)
				return
			}
			b, _ := jsonparser.ParseBoolean(value)
			fv.Elem().SetBool(b)
		case jsonparser.Null:
			if fValue.pointerType {
				fv = reflect.Zero(t)
			} else {
				el = append(el, errors.Errorf("Invalid type in JSON, cannot assign null to field %s", n))
				return
			}
		default:
			el = append(el, (errors.Errorf("Unexpected jsonparser value type %d", vt)))
			return
		}
		target := se.FieldByName(n)
		switch fValue.kind {
		case reflect.Ptr:
			target.Set(fv)
		case reflect.Slice, reflect.Map:
			if vt == jsonparser.Null {
				target.Set(fv)
			} else {
				target.Set(fv.Elem())
			}
		default:
			target.Set(fv.Elem())
		}
		modified = append(modified, n)
	}, fm.names...)

	if el == nil {
		return modified, nil
	}
	return nil, el
}

type fieldMap struct {
	names  [][]string
	values []fieldValue
}

type fieldValue struct {
	kind         reflect.Kind
	internalType reflect.Type
	internalKind reflect.Kind
	t            reflect.Type //type in struct
	name         string       //name in struct
	pointerType  bool
	unmarshaler  bool
	intType      bool
	uintType     bool
	floatType    bool
}

func buildJSONFieldMap(s interface{}) (fieldMap, error) {
	st := reflect.TypeOf(s)
	if st.Kind() != reflect.Ptr {
		return fieldMap{}, errors.New("Only works on pointers to structs")
	}
	stInner := st.Elem()
	if stInner.Kind() != reflect.Struct {
		return fieldMap{}, errors.New("Only works on pointers to structs")
	}
	out := fieldMap{}
	out.names = make([][]string, stInner.NumField())
	out.values = make([]fieldValue, stInner.NumField())
	for i := 0; i < stInner.NumField(); i++ {
		sf := stInner.Field(i)
		//skip over any chan fields or func fields
		if sf.Type.Kind() == reflect.Func || sf.Type.Kind() == reflect.Chan {
			continue
		}
		var fieldName string
		if name := sf.Tag.Get("json"); len(name) > 0 {
			fieldName = strings.Split(name, ",")[0]
		}
		if fieldName == "-" {
			continue
		}
		if fieldName == "" {
			fieldName = sf.Name
		}
		t := sf.Type
		k := t.Kind()
		it := t
		if k == reflect.Ptr {
			it = t.Elem()
		}
		itk := it.Kind()
		um := (t.Implements(unmarshalerType) || reflect.PtrTo(t).Implements(unmarshalerType))
		pt := t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Ptr
		intType := false
		uintType := false
		floatType := false
		switch itk {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intType = true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintType = true
		case reflect.Float32, reflect.Float64:
			floatType = true
		}

		out.names[i] = []string{fieldName}

		out.values[i] = fieldValue{
			t:            t,
			name:         sf.Name,
			kind:         k,
			internalType: it,
			unmarshaler:  um,
			internalKind: itk,
			pointerType:  pt,
			intType:      intType,
			uintType:     uintType,
			floatType:    floatType,
		}
	}
	return out, nil
}
