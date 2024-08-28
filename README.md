# struccy

[![Go Report Card](https://goreportcard.com/badge/github.com/itsatony/struccy)](https://goreportcard.com/report/github.com/itsatony/struccy)
[![GoDoc](https://godoc.org/github.com/itsatony/struccy?status.svg)](https://godoc.org/github.com/itsatony/struccy)

`struccy` is a Go package that provides utility functions for working with structs, focusing on struct manipulation, filtering, and merging capabilities, while handling role-based field access and type-safe operations.

## Installation

```bash
go get github.com/itsatony/struccy
```

## Features

- **Struct Field Access Control**: Utilize `readxs` and `writexs` tags to manage read and write access to struct fields based on roles.
- **Type-safe Merging and Filtering**: Safely merge and filter data between structs with full support for type conversions.
- **Conversion Utilities**: Convert structs to maps or JSON strings, applying field-level access rules.

## Field Access Control

`struccy` uses special struct tags (`readxs` and `writexs`) to control access to struct fields based on roles or scenarios:

- **`readxs`**: Controls visibility during read operations like filtering or converting to JSON/map.
- **`writexs`**: Controls modification access during struct updates or merges.

### Tag Features

- **Wildcard `*`**: Grants access to all roles.
- **Negation `!`**: Explicitly denies access to specified roles.

### Struct Field Access Example

```go
type Profile struct {
    Name    string `readxs:"*"`
    Email   string `readxs:"admin,user"`
    Address string `readxs:"!public"`
}
```

- `Name` is accessible to all roles.
- `Email` is accessible only to `admin` and `user` roles.
- `Address` is hidden from `public`.

## Struct Manipulation Functions

### UpdateStructFields

Updates fields of a struct based on non-zero values from another struct, applying role-based field access.

**Parameters**:

- `entity`: Target struct to update.
- `incomingEntity`: Struct containing update values.
- `roles`: Roles applicable to the operation.
- `ignoreZeroValues`: Flag to ignore zero values during update.
- `ignoreEmptyStrings`: Flag to ignore empty strings during update.

**Returns**:

- Updated fields map.
- Error if update fails due to access restrictions or type mismatches.

### SetField

Sets a value to a struct field with role-based access and type conversion.

**Parameters**:

- `entity`: Struct to update.
- `fieldName`: Field name to update.
- `value`: Value to set.
- `skipZeroVals`: Skip zero values during update.
- `roles`: User roles for access validation.

**Returns**:

- Error if the field can't be set due to access restrictions or type incompatibilities.

### IsAllowedToSetField

Checks if a field can be set based on the user's roles.

**Parameters**:

- `entity`: Struct containing the field.
- `fieldName`: Field name to check.
- `roles`: Roles to evaluate.

**Returns**:

- `true` if the field can be set, `false` otherwise.

## Usage Examples

### Merging Structs with Field Access

```go
adminUser := User{
    Email: "admin@example.com",
    Role:  "admin",
}

incomingUpdates := map[string]any{
    "Email": "newadmin@example.com",
    "Role":  "user",  // Assuming 'Role' field is protected and not writable by 'admin'
}

updatedUser, err := MergeMapStringFieldsToStruct(&adminUser, incomingUpdates, []string{"admin"})
if err != nil {
    log.Println("Failed to merge:", err)
}
fmt.Printf("Updated User: %+v\n", updatedUser)
```

### Filtering Structs to JSON with Role-based Access

```go
user := Profile{
    Name:    "John Doe",
    Email:   "john@example.com",
    Address: "Secret Location",
}

jsonOutput, err := StructToJSONFieldsWithReadXS(&user, []string{"user"})
if err != nil {
    log.Println("Error generating JSON:", err)
}
fmt.Println("JSON Output:", jsonOutput)
```

## Struct to Map/JSON Conversion

`struccy` provides functions to convert structs to maps or JSON strings, respecting `readxs` and `writexs` tags. This allows for dynamic data handling in applications that require role-based data visibility.

### Convenience Functions

The `struccy` package provides a set of convenience functions to work with struct fields and their read/write access rules (xsList). These functions allow you to retrieve field names, convert structs to maps, and convert structs to JSON strings based on specified access rules.

### GetFieldNames

The `GetFieldNames` function returns a slice of field names for the given struct pointer. It uses reflection to iterate over the fields of the struct and collect their names.

```go
func GetFieldNames(structPtr any) ([]string, error)
```

#### Example

```go
type MyStruct struct {
 Field1 string
 Field2 int
 Field3 bool
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

fieldNames, err := struccy.GetFieldNames(s)
if err != nil {
 log.Fatal(err)
}

fmt.Println(fieldNames)
// Output: [Field1 Field2 Field3]
```

### GetFieldNamesWithReadXS

The `GetFieldNamesWithReadXS` function returns a slice of field names for the given struct pointer, filtered by the specified read access rules (xsList). It uses reflection to iterate over the fields of the struct and collect the names of fields that have read access allowed based on the provided xsList.

```go
func GetFieldNamesWithReadXS(structPtr any, xsList []string) ([]string, error)
```

#### Example for GetFieldNamesWithReadXS

```go
type MyStruct struct {
 Field1 string `readxs:"admin,user"`
 Field2 int    `readxs:"admin"`
 Field3 bool   `readxs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

fieldNames, err := struccy.GetFieldNamesWithReadXS(s, []string{"user"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(fieldNames)
// Output: [Field1 Field3]
```

### GetFieldNamesWithWriteXS

The `GetFieldNamesWithWriteXS` function returns a slice of field names for the given struct pointer, filtered by the specified write access rules (xsList). It uses reflection to iterate over the fields of the struct and collect the names of fields that have write access allowed based on the provided xsList.

```go
func GetFieldNamesWithWriteXS(structPtr any, xsList []string) ([]string, error)
```

#### Example for GetFieldNamesWithWriteXS

```go
type MyStruct struct {
 Field1 string `writexs:"admin,user"`
 Field2 int    `writexs:"admin"`
 Field3 bool   `writexs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

fieldNames, err := struccy.GetFieldNamesWithWriteXS(s, []string{"admin"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(fieldNames)
// Output: [Field1 Field2]
```

### StructToMapFieldsWithReadXS

The `StructToMapFieldsWithReadXS` function converts the specified struct pointer to a map, including only the fields with read access allowed based on the provided xsList. It uses reflection to iterate over the fields of the struct and collect the field names and values that have read access allowed.

```go
func StructToMapFieldsWithReadXS(structPtr any, xsList []string) (map[string]any, error)
```

#### Example for StructToMapFieldsWithReadXS

```go
type MyStruct struct {
 Field1 string `readxs:"admin,user"`
 Field2 int    `readxs:"admin"`
 Field3 bool   `readxs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

fieldMap, err := struccy.StructToMapFieldsWithReadXS(s, []string{"user"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(fieldMap)
// Output: map[Field1:value1 Field3:true]
```

### StructToMapFieldsWithWriteXS

The `StructToMapFieldsWithWriteXS` function converts the specified struct pointer to a map, including only the fields with write access allowed based on the provided xsList. It uses reflection to iterate over the fields of the struct and collect the field names and values that have write access allowed.

```go
func StructToMapFieldsWithWriteXS(structPtr any, xsList []string) (map[string]any, error)
```

#### Example for StructToMapFieldsWithWriteXS

```go
type MyStruct struct {
 Field1 string `writexs:"admin,user"`
 Field2 int    `writexs:"admin"`
 Field3 bool   `writexs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

fieldMap, err := struccy.StructToMapFieldsWithWriteXS(s, []string{"admin"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(fieldMap)
// Output: map[Field1:value1 Field2:42]
```

### StructToJSONFieldsWithReadXS

The `StructToJSONFieldsWithReadXS` function converts the specified struct pointer to a JSON string, including only the fields with read access allowed based on the provided xsList. It uses reflection to iterate over the fields of the struct and collect the field names and values that have read access allowed, and then marshals the resulting map to a JSON string.

```go
func StructToJSONFieldsWithReadXS(structPtr any, xsList []string) (string, error)
```

#### Example for StructToJSONFieldsWithReadXS

```go
type MyStruct struct {
 Field1 string `readxs:"admin,user"`
 Field2 int    `readxs:"admin"`
 Field3 bool   `readxs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

jsonStr, err := struccy.StructToJSONFieldsWithReadXS(s, []string{"user"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(jsonStr)
// Output: {"Field1":"value1","Field3":true}
```

### StructToJSONFieldsWithWriteXS

The `StructToJSONFieldsWithWriteXS` function converts the specified struct pointer to a JSON string, including only the fields with write access allowed based on the provided xsList. It uses reflection to iterate over the fields of the struct and collect the field names and values that have write access allowed, and then marshals the resulting map to a JSON string.

```go
func StructToJSONFieldsWithWriteXS(structPtr any, xsList []string) (string, error)
```

#### Example for StructToJSONFieldsWithWriteXS

```go
type MyStruct struct {
 Field1 string `writexs:"admin,user"`
 Field2 int    `writexs:"admin"`
 Field3 bool   `writexs:"user"`
}

s := &MyStruct{
 Field1: "value1",
 Field2: 42,
 Field3: true,
}

jsonStr, err := struccy.StructToJSONFieldsWithWriteXS(s, []string{"admin"})
if err != nil {
 log.Fatal(err)
}

fmt.Println(jsonStr)
// Output: {"Field1":"value1","Field2":42}
```

These convenience functions provide additional flexibility and utility when working with structs and their fields based on read and write access rules. They can be used in scenarios where you need to retrieve field names, convert structs to maps, or convert structs to JSON strings while respecting the specified access rules.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or create issues for bugs and feature requests on our [GitHub repository](https://github.com/itsatony/struccy).

## License

`struccy` is licensed under the MIT License. See the LICENSE file for more details.

## Contact

For inquiries, please contact [itsatony](mailto:dev@vaudience.ai).
