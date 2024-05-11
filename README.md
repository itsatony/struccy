# struccy

[![Go Report Card](https://goreportcard.com/badge/github.com/itsatony/struccy)](https://goreportcard.com/report/github.com/itsatony/struccy)
[![GoDoc](https://godoc.org/github.com/itsatony/struccy?status.svg)](https://godoc.org/github.com/itsatony/struccy)

`struccy` is a Go package that provides utility functions for working with structs.

## Installation

To install the struccy package, run the following command:

```bash
go get github.com/itsatony/struccy
```

## JSON Tagging for `xsread` and `xswrite`

The `struccy` package introduces two special JSON tags, `xsread` and `xswrite`, which allow you to control the visibility of struct fields based on roles or scenarios. These tags provide a flexible way to include or exclude fields during struct filtering and merging operations.

### `xsread` Tag

The `xsread` tag is used to specify the roles or scenarios for which a field should be included when reading or filtering a struct. It accepts a comma-separated list of strings representing the allowed roles or scenarios.

Example:

```go
type User struct {
 Name     string `json:"name"`
 Email    string `json:"email" xsread:"admin,user"`
 Password string `json:"password" xsread:"admin"`
}
```

In this example, the `Email` field is tagged with `xsread:"admin,user"`, indicating that it should be included when filtering the struct for roles "admin" or "user". The `Password` field is tagged with `xsread:"admin"`, indicating that it should be included only when filtering for the "admin" role.

### `xswrite` Tag

The `xswrite` tag is used to specify the roles or scenarios for which a field should be included when writing or merging a struct. It follows the same format as the `xsread` tag, accepting a comma-separated list of strings representing the allowed roles or scenarios.

Example:

```go
type User struct {
 Name     string `json:"name"`
 Email    string `json:"email" xswrite:"admin,user"`
 Password string `json:"password" xswrite:"admin"`
}
```

In this example, the `Email` field is tagged with `xswrite:"admin,user"`, indicating that it should be included when merging the struct for roles "admin" or "user". The `Password` field is tagged with `xswrite:"admin"`, indicating that it should be included only when merging for the "admin" role.

### Special Characters

The `xsread` and `xswrite` tags support two special characters:

- `*`: Represents all roles or scenarios. When used, it includes the field for any role or scenario.
- `!`: Represents a negation. When used before a role or scenario, it excludes the field for that specific role or scenario.

Example:

```go
type User struct {
 Name     string `json:"name"`
 Email    string `json:"email" xsread:"*"`
 Password string `json:"password" xsread:"*,!public"`
}
```

In this example, the `Email` field is tagged with `xsread:"*"`, indicating that it should be included for all roles or scenarios. The `Password` field is tagged with `xsread:"*,!public"`, indicating that it should be included for all roles or scenarios except for the "public" role.

### Usage in `FilterStructTo` and `MergeStructUpdateTo`

The `FilterStructTo` and `MergeStructUpdateTo` functions in the `struccy` package utilize the `xsread` and `xswrite` tags to determine which fields should be included based on the provided roles or scenarios.

The `isFieldAccessAllowed` helper function is used internally to check if a field should be included based on the specified tags and the given roles or scenarios.

Example usage of `FilterStructTo`:

```go
user := User{
 Name:     "John Doe",
 Email:    "john@example.com",
 Password: "secret",
}

var filteredUser User
err := struccy.FilterStructTo(&user, &filteredUser, []string{"user"}, true)
if err != nil {
 log.Fatal(err)
}
fmt.Printf("%+v\n", filteredUser)
// Output: {Name:John Doe Email:john@example.com Password:}
```

In this example, the `FilterStructTo` function is called with the `user` struct, and the roles or scenarios are specified as `[]string{"user"}`. Based on the `xsread` tags, the `Name` and `Email` fields are included in the filtered struct, while the `Password` field is excluded.

Example usage of `MergeStructUpdateTo`:

```go
existingUser := User{
 Name:  "John Doe",
 Email: "john@example.com",
}

updatedUser := User{
 Email:    "johndoe@example.com",
 Password: "newpassword",
}

err := struccy.MergeStructUpdateTo(&updatedUser, &existingUser, []string{"admin"})
if err != nil {
 log.Fatal(err)
}
fmt.Printf("%+v\n", existingUser)
// Output: {Name:John Doe Email:johndoe@example.com Password:newpassword}
```

In this example, the `MergeStructUpdateTo` function is called with the `updatedUser` struct, and the roles or scenarios are specified as `[]string{"admin"}`. Based on the `xswrite` tags, the `Email` and `Password` fields are merged into the `existingUser` struct.

By leveraging the `xsread` and `xswrite` tags, you can control the visibility of struct fields based on roles or scenarios, providing a flexible way to filter and merge structs according to your application's requirements.

## General Usage

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
