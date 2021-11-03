# Due to changes in the priorities, this project is currently not being supported. The project is archived as of 11/3/21 and will be available in a read-only state. Please note, since archival, the project is not maintained or reviewed. #

# modtracker
JSON unmarshaling in Go that includes detection of the fields that were modified during unmarshaling

If you are doing a CRUD update via a REST API using JSON, you probably only want to update the database
 fields that were actually sent by the client. The standard Go JSON Unmarshaler doesn't provide this information.
 
 The modtracker package provides a way to discover this information. You can use it in a few ways. First, 
 you can call modtracker.UnmarshalJSON directly:
 
 ```go
 type Sample struct {
 	FirstName *string
 	LastName  *string
 	Age       *int
 }

func main() {
	var s Sample
	data := `
	    {
			"FirstName": "Homer",
    		"LastName": "Simpson",
    		"Age": 37
    	}
    `
	fmt.Println("converting", data)
	modified, err := modtracker.Unmarshal([]byte(data), &s)
	if err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(s,modified)
	}
}
```
 
You can wrap the call to modtracker.Unmarshal inside of an UnmarshalJSON method:
  
```go
 type Sample struct {
 	FirstName *string
 	LastName  *string
 	Age       *int
 	modified  []string
 }

func (s *Sample) UnmarshalJSON(data []byte) error {
	var err error
	s.modified, err = modtracker.Unmarshal(data, s)
	return err
}

func (s *Sample) GetModified() []string {
	return s.modified
}

func main() {
	var s Sample
	data := `
	    {
			"FirstName": "Homer",
    		"LastName": "Simpson",
    		"Age": 37
    	}
    `
	fmt.Println("converting", data)
	err := json.Unmarshal([]byte(data), &s)
	if err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(s)
	}
}
```

You can avoid some of the cost of reflection by pre-caclulating the fields, tags, and types using the 
BuildJSONUnmarshaler function:

```go
var sampleUnmarshaler modtracker.Unmarshaler

func init() {
	var err error
	sampleUnmarshaler, err = modtracker.BuildJSONUnmarshaler((*Sample)(nil))
	if err != nil {
		panic(err)
	}
}

type Sample struct {
 	FirstName *string
 	LastName  *string
 	Age       *int
 	modified  []string
 }

func (s *Sample) UnmarshalJSON(data []byte) error {
	var err error
	s.modified, err = sampleUnmarshaler(data, s)
	return err
}

func (s *Sample) GetModified() []string {
	return s.modified
}

func main() {
	var s Sample
	data := `
	    {
			"FirstName": "Homer",
    		"LastName": "Simpson",
    		"Age": 37
    	}
    `
	fmt.Println("converting", data)
	err := json.Unmarshal([]byte(data), &s)
	if err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(s)
	}
}
```
 
The modtracker unmarshalers respect json struct tags and work with both pointer and value fields. Fields of function type
and channel type are ignored.

Contributors:

We welcome your interest in Capital One’s Open Source Projects (the “Project”). Any Contributor to the project must accept and sign a CLA indicating agreement to the license terms. Except for the license granted in this CLA to Capital One and to recipients of software distributed by Capital One, you reserve all right, title, and interest in and to your contributions; this CLA does not impact your rights to use your own contributions for any other purpose.

[Link to CLA](https://docs.google.com/forms/d/e/1FAIpQLSfwtl1s6KmpLhCY6CjiY8nFZshDwf_wrmNYx1ahpsNFXXmHKw/viewform)

This project adheres to the [Open Source Code of Conduct](https://developer.capitalone.com/single/code-of-conduct/). By participating, you are expected to honor this code.
