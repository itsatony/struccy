# struccy

[![Go Report Card](https://goreportcard.com/badge/github.com/itsatony/struccy)](https://goreportcard.com/report/github.com/itsatony/struccy)
[![GoDoc](https://godoc.org/github.com/itsatony/struccy?status.svg)](https://godoc.org/github.com/itsatony/struccy)

`struccy` is a Go package that provides utility functions for working with structs.

## Installation

To install the struccy package, run the following command:

```bash
Copy codego get github.com/itsatony/struccy
```

## Usage

### MergeStructUpdateTo

The MergeStructUpdateTo function merges the fields of a source struct into a destination struct. It takes a pointer to the source struct and a pointer to the destination struct.

```go
func MergeStructUpdateTo(sourceStruct any, destinationStruct any) error
```

**Example:**

```go
type Source struct {
 Field1 string
 Field2 int
 Field3 *string
}

type Destination struct {
 Field1 string
 Field2 int
 Field3 *string
 Field4 bool
}

source := &Source{
 Field1: "Hello",
 Field2: 42,
 Field3: nil,
}

destination := &Destination{
 Field1: "World",
 Field2: 0,
 Field3: nil,
 Field4: true,
}

err := struccy.MergeStructUpdateTo(source, destination)
if err != nil {
 fmt.Println("Error:", err)
 return
}

fmt.Printf("Merged struct: %+v\n", destination)
```

**Output:**

```bash
Merged struct: &{Field1:Hello Field2:42 Field3:<nil> Field4:true}
```

### FilterStructTo

The FilterStructTo function filters the fields of a source struct and assigns the allowed fields to a destination struct. It takes a pointer to the source struct, a pointer to the destination struct, a list of excluded fields (xsList), and a boolean flag (zeroDisallowed) indicating whether zero values are allowed for excluded fields.

```go
func FilterStructTo(sourceStruct any, filteredStruct any, xsList []string, zeroDisallowed bool) error
```

**Example:**

```go
type Source struct {
    Field1 string
    Field2 int
    Field3 *string
}

type Filtered struct {
    Field1 string
    Field2 int
    Field3 *string
}

source := &Source{
    Field1: "Hello",
    Field2: 42,
    Field3: nil,
}

filtered := &Filtered{}

xsList := []string{"Field2"}
zeroDisallowed := true

err := struccy.FilterStructTo(source, filtered, xsList, zeroDisallowed)
if err != nil {
    fmt.Println("Error:", err)
    return
}

fmt.Printf("Filtered struct: %+v\n", filtered)
```

**Output:**

```bash
Filtered struct: &{Field1:Hello Field2:0 Field3:<nil>}
```

### StructToJSONFields

StructToJSONFields takes a pointer to a struct and a slice of field names, and returns a JSON string of the struct fields filtered to the specified field names.
If any error occurs during the process, an empty string and the error are returned.

```go
func StructToJSONFields(structPtr any, fieldNames []string) (string, error)
```

**Example:**

```go
type Example struct {
    Field1 string
    Field2 int
    Field3 *string
}

example := &Example{
    Field1: "Hello",
    Field2: 42,
    Field3: nil,
}

fieldNames := []string{"Field1", "Field3"}

jsonStr, err := struccy.StructToJSONFields(example, fieldNames)

if err != nil {
    fmt.Println("Error:", err)
    return
}

fmt.Println("Filtered JSON:", jsonStr)
```

**Output:**

```bash
Filtered JSON: {"Field1":"Hello","Field3":null}
```

### StructToMapFields

StructToMapFields takes a pointer to a struct and a slice of field names, and returns a map of the struct fields filtered to the specified field names.

```go
func StructToMapFields(structPtr any, fieldNames []string) (map[string]interface{}, error)
```

**Example:**

```go
type Example struct {
    Field1 string
    Field2 int
    Field3 *string
}

example := &Example{
    Field1: "Hello",
    Field2: 42,
    Field3: nil,
}

fieldNames := []string{"Field1", "Field3"}

fieldMap, err := struccy.StructToMapFields(example, fieldNames)

if err != nil {
    fmt.Println("Error:", err)
    return
}

fmt.Println("Filtered map:", fieldMap)
```

**Output:**

```bash
Filtered map: map[Field1:Hello Field3:<nil>]
```

### StructToMap

StructToMap takes a pointer to a struct and returns a map of the struct fields.

```go
func StructToMap(structPtr any) (map[string]interface{}, error)
```

**Example:**

```go
type Example struct {
    Field1 string
    Field2 int
    Field3 *string
}

example := &Example{
    Field1: "Hello",
    Field2: 42,
    Field3: nil,
}

fieldMap, err := struccy.StructToMap(example)

if err != nil {
    fmt.Println("Error:", err)
    return
}

fmt.Println("Field map:", fieldMap)
```

**Output:**

```bash
Field map: map[Field1:Hello Field2:42 Field3:<nil>]
```

## Contributing

Contributions to the struccy package are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

The struccy package is open-source software licensed under the MIT License.

## Acknowledgements

The struccy package was inspired by the need for a simple and efficient way to handle struct merging and filtering based on field-level access control in Go applications.

## Contact

For any questions or inquiries, please contact the package maintainer at [vaudience](dev@vaudience.ai) .
