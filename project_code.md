# project_code.md

## ./struccy_test.go

```go
package struccy

import (
	"errors"
	"reflect"
	"testing"
)

type TestStruct struct {
	Field1 string `writexs:"admin,user" readxs:"*"`
	Field2 int    `writexs:"admin" readxs:"admin,user"`
	Field3 *bool  `writexs:"admin,user" readxs:"admin"`
}

type TestStructDTO struct {
	Field1 *string `writexs:"admin,user" readxs:"*"`
	Field2 *int    `writexs:"admin" readxs:"admin,user"`
	Field3 *bool   `writexs:"admin,user" readxs:"admin"`
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func TestMergeStructUpdateTo(t *testing.T) {
	existingStruct := &TestStruct{
		Field1: "existing value",
		Field2: 10,
	}

	incomingStruct := &TestStructDTO{
		Field1: stringPtr("updated value"),
		Field2: intPtr(20),
		Field3: boolPtr(true),
	}

	testCases := []struct {
		name           string
		existingStruct any
		incomingStruct any
		xsRole         string
		expectedStruct *TestStruct
		expectedError  bool
	}{
		{
			name:           "Merge with admin role",
			existingStruct: existingStruct,
			incomingStruct: incomingStruct,
			xsRole:         "admin",
			expectedStruct: &TestStruct{
				Field1: "updated value",
				Field2: 20,
				Field3: boolPtr(true),
			},
			expectedError: false,
		},
		{
			name:           "Merge with user role",
			existingStruct: existingStruct,
			incomingStruct: incomingStruct,
			xsRole:         "user",
			expectedStruct: &TestStruct{
				Field1: "updated value",
				Field2: 10,
				Field3: boolPtr(true),
			},
			expectedError: false,
		},
		{
			name:           "Merge with guest role",
			existingStruct: existingStruct,
			incomingStruct: incomingStruct,
			xsRole:         "guest",
			expectedStruct: &TestStruct{
				Field1: "existing value",
				Field2: 10,
				Field3: nil,
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mergedStruct, err := MergeStructUpdateTo(tc.existingStruct, tc.incomingStruct, []string{tc.xsRole})
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(mergedStruct, tc.expectedStruct) {
				t.Errorf("Expected merged struct: %+v, got: %+v", tc.expectedStruct, mergedStruct)
			}
		})
	}
}

func TestFilterStructFor(t *testing.T) {
	mergedStruct := &TestStruct{
		Field1: "merged value",
		Field2: 30,
		Field3: boolPtr(true),
	}

	testCases := []struct {
		name           string
		mergedStruct   any
		targetStruct   any
		xsRole         string
		expectedStruct any
		expectedError  bool
	}{
		{
			name:         "Filter with admin role",
			mergedStruct: mergedStruct,
			targetStruct: &TestStruct{},
			xsRole:       "admin",
			expectedStruct: &TestStruct{
				Field1: "merged value",
				Field2: 30,
				Field3: boolPtr(true),
			},
			expectedError: false,
		},
		{
			name:         "Filter with user role",
			mergedStruct: mergedStruct,
			targetStruct: &TestStruct{},
			xsRole:       "user",
			expectedStruct: &TestStruct{
				Field1: "merged value",
				Field2: 30,
			},
			expectedError: false,
		},
		{
			name:         "Filter with guest role",
			mergedStruct: mergedStruct,
			targetStruct: &TestStruct{},
			xsRole:       "guest",
			expectedStruct: &TestStruct{
				Field1: "merged value",
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FilterStructTo(tc.mergedStruct, tc.targetStruct, []string{tc.xsRole}, false)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.targetStruct, tc.expectedStruct) {
				t.Errorf("Expected filtered struct: %+v, got: %+v", tc.expectedStruct, tc.targetStruct)
			}
		})
	}
}
func TestMergeStructUpdateTo_InvalidInput(t *testing.T) {
	testCases := []struct {
		name          string
		targetStruct  any
		updateStruct  any
		xsList        []string
		expectedError bool
	}{
		{
			name:          "Non-pointer target struct",
			targetStruct:  TestStruct{},
			updateStruct:  &TestStructDTO{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
		{
			name:          "Non-pointer update struct",
			targetStruct:  &TestStruct{},
			updateStruct:  TestStructDTO{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
		{
			name:          "Non-struct target value",
			targetStruct:  &[]string{},
			updateStruct:  &TestStructDTO{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
		{
			name:          "Non-struct update value",
			targetStruct:  &TestStruct{},
			updateStruct:  &[]string{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
		{
			name:         "Mismatched field types",
			targetStruct: &TestStruct{},
			updateStruct: &struct {
				Field1 int `writexs:"admin,user" readxs:"*"`
			}{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := MergeStructUpdateTo(tc.targetStruct, tc.updateStruct, tc.xsList)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestFilterStructTo_InvalidInput(t *testing.T) {
	testCases := []struct {
		name           string
		sourceStruct   any
		filteredStruct any
		xsList         []string
		expectedError  bool
	}{
		{
			name:           "Non-pointer source struct",
			sourceStruct:   TestStruct{},
			filteredStruct: &TestStruct{},
			xsList:         []string{"admin"},
			expectedError:  true,
		},
		{
			name:           "Non-pointer filtered struct",
			sourceStruct:   &TestStruct{},
			filteredStruct: TestStruct{},
			xsList:         []string{"admin"},
			expectedError:  true,
		},
		{
			name:           "Non-struct source value",
			sourceStruct:   &[]string{},
			filteredStruct: &TestStruct{},
			xsList:         []string{"admin"},
			expectedError:  true,
		},
		{
			name:           "Non-struct filtered value",
			sourceStruct:   &TestStruct{},
			filteredStruct: &[]string{},
			xsList:         []string{"admin"},
			expectedError:  true,
		},
		{
			name:         "Mismatched field types",
			sourceStruct: &TestStruct{},
			filteredStruct: &struct {
				Field1 int `writexs:"admin,user" readxs:"*"`
			}{},
			xsList:        []string{"admin"},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FilterStructTo(tc.sourceStruct, tc.filteredStruct, tc.xsList, false)
			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestFilterStructTo_ZeroDisallowed(t *testing.T) {
	sourceStruct := &TestStruct{
		Field1: "source value",
		Field2: 30,
		Field3: boolPtr(true),
	}

	filteredStruct := &TestStruct{}

	err := FilterStructTo(sourceStruct, filteredStruct, []string{"user"}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedStruct := &TestStruct{
		Field1: "source value",
		Field2: 0,
	}

	if !reflect.DeepEqual(filteredStruct, expectedStruct) {
		t.Errorf("Expected filtered struct: %+v, got: %+v", expectedStruct, filteredStruct)
	}
}

func TestIsFieldAccessAllowed_NotSyntax(t *testing.T) {
	testCases := []struct {
		name         string
		xsList       []string
		allowedRoles string
		expected     bool
	}{
		{
			name:         "NOT syntax - allowed",
			xsList:       []string{"admin", "user"},
			allowedRoles: "!guest",
			expected:     true,
		},
		{
			name:         "NOT syntax - not allowed",
			xsList:       []string{"admin", "guest"},
			allowedRoles: "!guest",
			expected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isFieldAccessAllowed(tc.xsList, tc.allowedRoles)
			if result != tc.expected {
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}

func TestStructToJSONFields(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
		field3 bool // Unexported field
	}

	testStruct := &TestStruct{
		Field1: "value1",
		Field2: 123,
		field3: true,
	}

	testCases := []struct {
		name          string
		structPtr     any
		fieldNames    []string
		expectedJSON  string
		expectedError error
	}{
		{
			name:         "Valid struct and fields",
			structPtr:    testStruct,
			fieldNames:   []string{"Field1", "Field2"},
			expectedJSON: `{"Field1":"value1","Field2":123}`,
		},
		{
			name:          "Invalid struct pointer",
			structPtr:     123,
			fieldNames:    []string{"Field1"},
			expectedError: ErrInvalidStructPointer,
		},
		{
			name:          "Field not found",
			structPtr:     testStruct,
			fieldNames:    []string{"Field4"},
			expectedError: ErrFieldNotFound,
		},
		{
			name:          "Unexported field",
			structPtr:     testStruct,
			fieldNames:    []string{"field3"},
			expectedError: ErrUnexportedField,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr, err := StructToJSONFields(tc.structPtr, tc.fieldNames)

			if tc.expectedError != nil {
				if err == nil || !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if jsonStr != tc.expectedJSON {
					t.Errorf("Expected JSON: %s, got: %s", tc.expectedJSON, jsonStr)
				}
			}
		})
	}
}
```

## ./struccy.go

```go
package struccy

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

var (
	ErrTargetStructMustBePointer   = errors.New("targetStruct must be a pointer to a struct")
	ErrUpdateStructMustBePointer   = errors.New("updateStruct must be a pointer to a struct")
	ErrSourceStructMustBePointer   = errors.New("sourceStruct must be a pointer to a struct")
	ErrFilteredStructMustBePointer = errors.New("filteredStruct must be a pointer to a struct")
	ErrFieldMismatchedTypes        = "field %s has mismatched types: %v != %v"
	ErrFieldNotFound               = errors.New("field not found")
	ErrUnsupportedFieldType        = errors.New("unsupported field type")
	ErrInvalidAccessRole           = errors.New("invalid access role")
	ErrFieldTypeMismatch           = errors.New("fiel type mismatch")
	ErrInvalidStructPointer        = errors.New("invalid struct pointer")
	ErrUnexportedField             = errors.New("unexported field")
	ErrJSONMarshalFailed           = errors.New("JSON marshal failed")
)

func MergeStructUpdateTo(targetStruct any, updateStruct any, xsList []string) (any, error) {
	targetValue := reflect.ValueOf(targetStruct)
	updateValue := reflect.ValueOf(updateStruct)

	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return nil, ErrTargetStructMustBePointer
	}

	if updateValue.Kind() != reflect.Ptr || updateValue.Elem().Kind() != reflect.Struct {
		return nil, ErrUpdateStructMustBePointer
	}

	targetType := targetValue.Elem().Type()
	updateType := updateValue.Elem().Type()

	mergedStruct := reflect.New(targetType).Elem()

	for i := 0; i < targetType.NumField(); i++ {
		targetField := targetValue.Elem().Field(i)
		mergedStruct.Field(i).Set(targetField)
	}

	for i := 0; i < updateType.NumField(); i++ {
		field := updateType.Field(i)
		updateField := updateValue.Elem().Field(i)

		targetField := mergedStruct.FieldByName(field.Name)
		if !targetField.IsValid() {
			return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, field.Name)
		}

		writexs := field.Tag.Get("writexs")
		if !isFieldAccessAllowed(xsList, writexs) {
			continue
		}

		if updateField.Kind() == reflect.Ptr {
			if !updateField.IsNil() {
				if targetField.Kind() == reflect.Ptr {
					targetField.Set(updateField)
				} else {
					targetField.Set(updateField.Elem())
				}
			}
		} else {
			if targetField.Kind() == reflect.Ptr {
				targetField.Set(reflect.New(targetField.Type().Elem()))
				targetField.Elem().Set(updateField)
			} else {
				if updateField.Type().AssignableTo(targetField.Type()) {
					targetField.Set(updateField)
				} else {
					return nil, fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, field.Name, targetField.Type(), updateField.Type())
				}
			}
		}

		switch updateField.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface:
			return nil, fmt.Errorf("%w: %s, type %v", ErrUnsupportedFieldType, field.Name, updateField.Type())
		}
	}

	return mergedStruct.Addr().Interface(), nil
}

func FilterStructTo(sourceStruct any, filteredStruct any, xsList []string, zeroDisallowed bool) error {
	sourceValue := reflect.ValueOf(sourceStruct)
	filteredValue := reflect.ValueOf(filteredStruct)

	if sourceValue.Kind() != reflect.Ptr || sourceValue.Elem().Kind() != reflect.Struct {
		return ErrSourceStructMustBePointer
	}

	if filteredValue.Kind() != reflect.Ptr || filteredValue.Elem().Kind() != reflect.Struct {
		return ErrFilteredStructMustBePointer
	}

	sourceType := sourceValue.Elem().Type()
	filteredType := filteredValue.Elem().Type()

	sourceFields := make(map[string]reflect.Value)
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		sourceFields[field.Name] = sourceValue.Elem().Field(i)
	}

	for i := 0; i < filteredType.NumField(); i++ {
		field := filteredType.Field(i)
		filteredField := filteredValue.Elem().Field(i)

		sourceField, ok := sourceFields[field.Name]
		if !ok {
			if zeroDisallowed {
				filteredField.Set(reflect.Zero(filteredField.Type()))
			}
			continue
		}

		if sourceField.Type() != filteredField.Type() {
			return fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, field.Name, filteredField.Type(), sourceField.Type())
		}

		readxs := field.Tag.Get("readxs")
		if !isFieldAccessAllowed(xsList, readxs) {
			if zeroDisallowed {
				filteredField.Set(reflect.Zero(filteredField.Type()))
			}
			continue
		}

		if sourceField.Kind() == reflect.Ptr {
			if !sourceField.IsNil() {
				if filteredField.Kind() != reflect.Ptr {
					filteredField.Set(sourceField.Elem())
				} else {
					filteredField.Set(sourceField)
				}
			}
		} else {
			filteredField.Set(sourceField)
		}
	}

	return nil
}

// StructToJSONFields takes a pointer to a struct and a slice of field names,
// and returns a JSON string of the struct fields filtered to the specified field names.
// If any error occurs during the process, an empty string and the error are returned.
func StructToJSONFields(structPtr any, fieldNames []string) (string, error) {
	// Check if structPtr is a pointer to a struct
	structValue := reflect.ValueOf(structPtr)
	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return "", fmt.Errorf("%w: expected pointer to struct, got %v", ErrInvalidStructPointer, structValue.Type())
	}

	// Create a new map to store the filtered fields
	filteredFields := make(map[string]any)

	// Iterate over the field names
	for _, fieldName := range fieldNames {
		// Check if the field exists in the struct
		fieldValue := structValue.Elem().FieldByName(fieldName)
		if !fieldValue.IsValid() {
			return "", fmt.Errorf("%w: %s", ErrFieldNotFound, fieldName)
		}

		// Check if the field is exported (capitalized)
		if !fieldValue.CanInterface() {
			return "", fmt.Errorf("%w: %s", ErrUnexportedField, fieldName)
		}

		// Add the field to the filtered fields map
		filteredFields[fieldName] = fieldValue.Interface()
	}

	// Convert the filtered fields map to JSON
	jsonBytes, err := json.Marshal(filteredFields)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrJSONMarshalFailed, err)
	}

	return string(jsonBytes), nil
}

func isFieldAccessAllowed(xsList []string, allowedRoles string) bool {
	if allowedRoles == "*" {
		return true
	}

	roles := strings.Split(allowedRoles, ",")
	for _, role := range roles {
		if strings.HasPrefix(role, "!") {
			if !slices.Contains(xsList, strings.TrimPrefix(role, "!")) {
				return true
			}
		} else if slices.Contains(xsList, role) {
			return true
		}
	}

	return false
}
```

