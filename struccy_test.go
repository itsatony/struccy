package struccy

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestFilterStructTo_PointerToSlice(t *testing.T) {
	type Source struct {
		Field1 *[]string `readxs:"*"`
	}

	type Filtered struct {
		Field1 []string
	}

	source := &Source{
		Field1: &[]string{"value1", "value2"},
	}

	var filtered Filtered
	err := FilterStructTo(source, &filtered, []string{"user"}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := Filtered{
		Field1: []string{"value1", "value2"},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("Expected filtered struct: %+v, got: %+v", expected, filtered)
	}
}

func TestFilterStructTo_AgentExample(t *testing.T) {
	type AgentWriteDto struct {
		Name                *string   `json:"name" validate:"omitempty,min=1,max=255" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		Description         *string   `json:"description" validate:"omitempty,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		ModelID             *string   `json:"model_id" validate:"omitempty,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AvatarURL           *string   `json:"avatar_url" validate:"omitempty,url" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		SystemMessages      *[]string `json:"system_messages" validate:"omitempty,unique,dive,min=1,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		InitialUserMessages *[]string `json:"initial_user_messages" validate:"omitempty,unique,dive,min=1,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AttachedFileIDs     *[]string `json:"attached_file_ids" validate:"omitempty,unique,dive,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AssignedTools       *[]string `json:"assigned_tools" validate:"omitempty,unique,dive,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
	}

	type Agent struct {
		ID                  string   `json:"id" validate:"required,min=1,max=64" xswrite:"system" xsread:"system,admin,owner,org"`
		Name                string   `json:"name" validate:"required,min=1,max=255" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		Description         string   `json:"description" validate:"omitempty,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		ModelID             string   `json:"model_id" validate:"required,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AvatarURL           string   `json:"avatar_url" validate:"omitempty,url" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		SystemMessages      []string `json:"system_messages" validate:"unique,dive,min=1,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		InitialUserMessages []string `json:"initial_user_messages" validate:"unique,dive,min=1,max=1024" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AttachedFileIDs     []string `json:"attached_file_ids" validate:"unique,dive,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		AssignedTools       []string `json:"assigned_tools" validate:"unique,dive,min=1,max=64" xswrite:"system,admin,owner" xsread:"system,admin,owner,org"`
		OwnerId             string   `json:"owner_id" validate:"required,min=1,max=64" xswrite:"system,admin" xsread:"system,admin,owner,org"`
		OwnerOrganizationId string   `json:"owner_organization_id" validate:"required,min=1,max=64" xswrite:"system,admin" xsread:"system,admin,owner,org"`
		CreatedAt           int64    `json:"created_at" xswrite:"system,admin" xsread:"system,admin,owner,org"`
		UpdatedBy           string   `json:"updated_by" validate:"omitempty,min=1,max=64" xswrite:"system,admin" xsread:"system,admin,owner,org"`
		UpdatedAt           int64    `json:"updated_at" xswrite:"system,admin" xsread:"system,admin,owner,org"`
	}

	agentWriteDto := &AgentWriteDto{
		Name:           strPtr("Agent Name"),
		Description:    strPtr("Agent Description"),
		ModelID:        strPtr("model123"),
		AvatarURL:      strPtr("https://example.com/avatar.png"),
		SystemMessages: &[]string{"System message 1", "System message 2"},
	}

	var agent Agent
	err := FilterStructTo(agentWriteDto, &agent, []string{"admin", "owner"}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := Agent{
		Name:           "Agent Name",
		Description:    "Agent Description",
		ModelID:        "model123",
		AvatarURL:      "https://example.com/avatar.png",
		SystemMessages: []string{"System message 1", "System message 2"},
	}

	if !reflect.DeepEqual(agent, expected) {
		t.Errorf("Expected filtered struct: %+v, got: %+v", expected, agent)
	}
}

func strPtr(s string) *string {
	return &s
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
		Field2: 30,
	}

	if !reflect.DeepEqual(filteredStruct, expectedStruct) {
		t.Errorf("Expected filtered struct: %+v, got: %+v", expectedStruct, filteredStruct)
	}
}

func TestIsFieldAccessAllowed_NotSyntax(t *testing.T) {
	testCases := []struct {
		name        string
		xsList      []string
		taggedRoles string
		expected    bool
	}{
		{
			name:        "NOT syntax - allowed #1",
			xsList:      []string{"admin", "user"},
			taggedRoles: "!guest",
			expected:    true,
		},
		{
			name:        "NOT syntax - allowed #2",
			xsList:      []string{"admin", "guest"},
			taggedRoles: "!guest",
			expected:    false,
		},
		{
			name:        "NOT syntax - not allowed",
			xsList:      []string{"guest"},
			taggedRoles: "!guest",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsFieldAccessAllowed(tc.xsList, tc.taggedRoles)
			if result != tc.expected {
				t.Errorf("Expected: %v, got: %v  <-- for %v  vs.  %v", tc.expected, result, tc.xsList, tc.taggedRoles)
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

func TestStructToMapFields(t *testing.T) {
	type MyStruct struct {
		Field1 string
		Field2 int
		Field3 *string
	}

	str := "Hello"

	tests := []struct {
		name       string
		structPtr  any
		fieldNames []string
		expected   map[string]interface{}
		err        error
	}{
		{
			name: "Valid struct with selected fields",
			structPtr: &MyStruct{
				Field1: "Value1",
				Field2: 42,
				Field3: &str,
			},
			fieldNames: []string{"Field1", "Field3"},
			expected: map[string]interface{}{
				"Field1": "Value1",
				"Field3": &str,
			},
			err: nil,
		},
		{
			name: "Valid struct with non-existent field",
			structPtr: &MyStruct{
				Field1: "Value1",
				Field2: 42,
				Field3: &str,
			},
			fieldNames: []string{"Field1", "Field4"},
			expected: map[string]interface{}{
				"Field1": "Value1",
			},
			err: nil,
		},
		{
			name:       "Invalid struct pointer",
			structPtr:  "not a struct",
			fieldNames: []string{"Field1"},
			expected:   nil,
			err:        ErrInvalidStructPointer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StructToMapFields(tt.structPtr, tt.fieldNames)
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected result: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	type MyStruct struct {
		Field1 string
		Field2 int
		Field3 *string
	}

	str := "Hello"

	tests := []struct {
		name      string
		structPtr any
		expected  map[string]interface{}
		err       error
	}{
		{
			name: "Valid struct",
			structPtr: &MyStruct{
				Field1: "Value1",
				Field2: 42,
				Field3: &str,
			},
			expected: map[string]interface{}{
				"Field1": "Value1",
				"Field2": 42,
				"Field3": &str,
			},
			err: nil,
		},
		{
			name:      "Invalid struct pointer",
			structPtr: "not a struct",
			expected:  nil,
			err:       ErrInvalidStructPointer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StructToMap(tt.structPtr)
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected result: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestGetFieldNames(t *testing.T) {
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

	fieldNames, err := GetFieldNames(s)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{"Field1", "Field2", "Field3"}
	if !reflect.DeepEqual(fieldNames, expected) {
		t.Errorf("Expected field names: %v, got: %v", expected, fieldNames)
	}
}

func TestGetFieldNamesWithReadXS(t *testing.T) {
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

	fieldNames, err := GetFieldNamesWithReadXS(s, []string{"user"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{"Field1", "Field3"}
	if !reflect.DeepEqual(fieldNames, expected) {
		t.Errorf("Expected field names: %v, got: %v", expected, fieldNames)
	}
}

func TestGetFieldNamesWithWriteXS(t *testing.T) {
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

	fieldNames, err := GetFieldNamesWithWriteXS(s, []string{"admin"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []string{"Field1", "Field2"}
	if !reflect.DeepEqual(fieldNames, expected) {
		t.Errorf("Expected field names: %v, got: %v", expected, fieldNames)
	}
}

func TestStructToMapFieldsWithReadXS(t *testing.T) {
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

	fieldMap, err := StructToMapFieldsWithReadXS(s, []string{"user"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := map[string]any{
		"Field1": "value1",
		"Field3": true,
	}
	if !reflect.DeepEqual(fieldMap, expected) {
		t.Errorf("Expected field map: %v, got: %v", expected, fieldMap)
	}
}

func TestStructToMapFieldsWithWriteXS(t *testing.T) {
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

	fieldMap, err := StructToMapFieldsWithWriteXS(s, []string{"admin"}, false, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := map[string]any{
		"Field1": "value1",
		"Field2": 42,
	}
	if !reflect.DeepEqual(fieldMap, expected) {
		t.Errorf("Expected field map: %v, got: %v", expected, fieldMap)
	}
}

func TestStructToMapFieldsWithWriteXS_SkipNil(t *testing.T) {
	type MyStruct struct {
		Field1 *string `writexs:"admin,user"`
		Field2 *int    `writexs:"admin"`
		Field3 *bool   `writexs:"user"`
	}

	field1 := "value1"
	field3 := true
	s := &MyStruct{
		Field1: &field1,
		Field2: nil,
		Field3: &field3,
	}

	fieldMap, err := StructToMapFieldsWithWriteXS(s, []string{"admin"}, true, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := map[string]any{
		"Field1": &field1,
	}
	if !reflect.DeepEqual(fieldMap, expected) {
		t.Errorf("Expected field map: %v, got: %v", expected, fieldMap)
	}
}

func TestStructToMapFieldsWithWriteXS_SkipNilJsonFieldNames(t *testing.T) {
	type MyStruct struct {
		Field1 *string `writexs:"admin,user" json:"field_1"`
		Field2 *int    `writexs:"admin" json:"field2"`
		Field3 *bool   `writexs:"user" json:"field3"`
	}

	field1 := "value1"
	field3 := true
	s := &MyStruct{
		Field1: &field1,
		Field2: nil,
		Field3: &field3,
	}

	fieldMap, err := StructToMapFieldsWithWriteXS(s, []string{"admin"}, true, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := map[string]any{
		"field_1": &field1,
	}
	if !reflect.DeepEqual(fieldMap, expected) {
		t.Errorf("Expected field map: %v, got: %v", expected, fieldMap)
	}
}

func TestStructToJSONFieldsWithReadXS(t *testing.T) {
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

	jsonStr, err := StructToJSONFieldsWithReadXS(s, []string{"user"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `{"Field1":"value1","Field3":true}`
	if jsonStr != expected {
		t.Errorf("Expected JSON string: %s, got: %s", expected, jsonStr)
	}
}

func TestStructToJSONFieldsWithWriteXS(t *testing.T) {
	type MyStruct struct {
		Field1 string `writexs:"admin,user" json:"field1"`
		Field2 int    `writexs:"admin" json:"field2"`
		Field3 bool   `writexs:"user" json:"field3"`
	}

	s := &MyStruct{
		Field1: "value1",
		Field2: 42,
		Field3: true,
	}

	jsonStr, err := StructToJSONFieldsWithWriteXS(s, []string{"admin"}, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `{"field1":"value1","field2":42}`
	if jsonStr != expected {
		t.Errorf("Expected JSON string: %s, got: %s", expected, jsonStr)
	}
}

func TestStructToJSONFieldsWithWriteXS_SkipNil(t *testing.T) {
	type MyStruct struct {
		Field1 *string `writexs:"admin,user" json:"field1"`
		Field2 *int    `writexs:"admin" json:"field2"`
		Field3 *bool   `writexs:"user" json:"field3"`
	}

	field1 := "value1"
	field3 := true
	s := &MyStruct{
		Field1: &field1,
		Field2: nil,
		Field3: &field3,
	}

	jsonStr, err := StructToJSONFieldsWithWriteXS(s, []string{"admin"}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `{"field1":"value1"}`
	if jsonStr != expected {
		t.Errorf("Expected JSON string: %s, got: %s", expected, jsonStr)
	}
}

func TestMergeMapStringFieldsToStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
		Field3 *bool
		Field4 *string
		Field5 []int
	}

	trueVal := true
	falseVal := false
	initialString := "Hello, World!"
	newString := "Goodbye, World!"

	tests := []struct {
		name         string
		targetStruct any
		updateMap    map[string]any
		expected     TestStruct
		expectError  bool
	}{
		{
			name: "Merge fields with matching types",
			targetStruct: &TestStruct{
				Field1: "initial1",
				Field2: 10,
				Field3: nil,
			},
			updateMap: map[string]any{
				"Field1": "updated1",
				"Field2": 20,
				"Field3": &trueVal,
			},
			expected: TestStruct{
				Field1: "updated1",
				Field2: 20,
				Field3: &trueVal,
			},
			expectError: false,
		},
		{
			name: "Successfully merge fields with type conversion",
			targetStruct: &TestStruct{
				Field1: "initial1",
				Field2: 10,
				Field3: nil,
			},
			updateMap: map[string]any{
				"Field1": "updated1",
				"Field2": int8(20),
				"Field3": trueVal,
			},
			expected: TestStruct{
				Field1: "initial1",
				Field2: 20,
				Field3: nil,
			},
			expectError: false,
		},
		{
			name: "Handle nil assignment to existing pointers",
			targetStruct: &TestStruct{
				Field1: "initial1",
				Field2: 10,
				Field3: &trueVal,
				Field4: &initialString,
			},
			updateMap: map[string]any{
				"Field3": nil,
				"Field4": nil,
			},
			expected: TestStruct{
				Field1: "initial1",
				Field2: 10,
				Field3: nil,
				Field4: nil,
			},
			expectError: false,
		},
		{
			name: "Handle non-pointer to pointer conversion with valid types",
			targetStruct: &TestStruct{
				Field3: nil,
			},
			updateMap: map[string]any{
				"Field3": trueVal,
			},
			expected: TestStruct{
				Field3: &trueVal,
			},
			expectError: false,
		},
		{
			name: "Reject incompatible type conversions",
			targetStruct: &TestStruct{
				Field2: 10,
			},
			updateMap: map[string]any{
				"Field2": "should fail",
			},
			expected: TestStruct{
				Field2: 10,
			},
			expectError: true,
		},
		{
			name: "Reject assignments where target is not a slice but input is a slice",
			targetStruct: &TestStruct{
				Field2: 10,
			},
			updateMap: map[string]any{
				"Field2": []int{1, 2, 3},
			},
			expected: TestStruct{
				Field2: 10,
			},
			expectError: true,
		},
		{
			name: "Accept slice assignments when types match",
			targetStruct: &TestStruct{
				Field5: []int{1, 2, 3},
			},
			updateMap: map[string]any{
				"Field5": []int{4, 5, 6},
			},
			expected: TestStruct{
				Field5: []int{4, 5, 6},
			},
			expectError: false,
		},
		{
			name: "Successfully convert boolean to pointer boolean",
			targetStruct: &TestStruct{
				Field3: &falseVal,
			},
			updateMap: map[string]any{
				"Field3": trueVal,
			},
			expected: TestStruct{
				Field3: &trueVal,
			},
			expectError: false,
		},
		{
			name: "Update string pointer field",
			targetStruct: &TestStruct{
				Field4: &initialString,
			},
			updateMap: map[string]any{
				"Field4": newString,
			},
			expected: TestStruct{
				Field4: &newString,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeMapStringFieldsToStruct(tt.targetStruct, tt.updateMap, nil)
			if (err != nil) != tt.expectError {
				t.Errorf("Test '%s' failed: expected error %v, got %v", tt.name, tt.expectError, err)
			}
			if err == nil && !reflect.DeepEqual(result, tt.targetStruct) {
				t.Errorf("Test '%s' failed: expected result %+v, got %+v", tt.name, tt.expected, result)
			}
		})
	}
}

type RoleBasedStruct struct {
	PublicField string `writexs:"*"`
	AdminField  string `writexs:"admin"`
	UserField   string `writexs:"user"`
}

func TestUpdateStructFields(t *testing.T) {
	initial := &RoleBasedStruct{
		PublicField: "initial",
		AdminField:  "admin only",
		UserField:   "user only",
	}
	initial2 := &RoleBasedStruct{
		PublicField: "initial",
		AdminField:  "admin only",
		UserField:   "user only",
	}

	updates := &RoleBasedStruct{
		PublicField: "updated",
		AdminField:  "updated admin",
		UserField:   "updated user",
	}

	expectedAdmin := &RoleBasedStruct{
		PublicField: "updated",
		AdminField:  "updated admin",
		UserField:   "user only",
	}

	expectedUser := &RoleBasedStruct{
		PublicField: "updated",
		AdminField:  "admin only",
		UserField:   "updated user",
	}

	// Testing with admin role
	updatedFields, _, err := UpdateStructFields(initial, updates, []string{"admin"}, true, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedAdmin, initial)
	assert.Len(t, updatedFields, 2, "Two fields should have been updated for admin")

	// Testing with user role
	updatedFields2, _, err := UpdateStructFields(initial2, updates, []string{"user"}, true, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, initial2)
	assert.Len(t, updatedFields2, 2, "Two fields should have been updated for user")
}

type RoleBasedDto struct {
	AdminField string
}

func TestUpdateStructFieldsPartial(t *testing.T) {
	initial := &RoleBasedStruct{
		PublicField: "initial",
		AdminField:  "admin only",
		UserField:   "user only",
	}

	updates := &RoleBasedDto{
		AdminField: "updated admin",
	}

	expectedAdmin := &RoleBasedStruct{
		PublicField: "initial",
		AdminField:  "updated admin",
		UserField:   "user only",
	}

	updatedFields, _, err := UpdateStructFields(initial, updates, []string{"admin"}, true, false)
	assert.NoError(t, err)
	assert.Equal(t, expectedAdmin, initial)
	assert.Len(t, updatedFields, 1, "Admin field should have been updated")
}

// TestSetField tests the SetField function
func TestSetField(t *testing.T) {
	type TestStruct struct {
		Field1 string `writexs:"admin,user"`
		Field2 int    `writexs:"admin"`
	}

	entity := &TestStruct{
		Field1: "initial value",
		Field2: 10,
	}

	// Test case 1: Set field with admin role
	err := SetField(entity, "Field1", "updated value", true, []string{"admin"})
	assert.NoError(t, err)
	assert.Equal(t, "updated value", entity.Field1)

	// Test case 2: Set field with user role
	err = SetField(entity, "Field1", "updated value", true, []string{"user"})
	assert.NoError(t, err)
	assert.Equal(t, "updated value", entity.Field1)

	// Test case 3: Set field with unauthorized role
	err = SetField(entity, "Field2", 20, true, []string{"user"})
	assert.Equal(t, ErrUnauthorizedFieldSet, err)
	assert.Equal(t, 10, entity.Field2)

	// Test case 4: Set invalid field
	err = SetField(entity, "InvalidField", "value", true, []string{"admin"})
	assert.Equal(t, ErrInvalidFieldName, err)

	// Test case 5: Set field with valid value type
	err = SetField(entity, "Field1", "valid value", true, []string{"admin"})
	assert.NoError(t, err)
	assert.Equal(t, "valid value", entity.Field1)
}

// TestIsAllowedToSetField tests the IsAllowedToSetField function
func TestIsAllowedToSetField(t *testing.T) {
	type TestStruct struct {
		Field1 string `writexs:"admin,user"`
		Field2 int    `writexs:"admin"`
	}

	entity := &TestStruct{}

	// Test case 1: Check field with admin role
	assert.True(t, IsAllowedToSetField(entity, "Field1", []string{"admin"}))
	assert.True(t, IsAllowedToSetField(entity, "Field2", []string{"admin"}))

	// Test case 2: Check field with user role
	assert.True(t, IsAllowedToSetField(entity, "Field1", []string{"user"}))
	assert.False(t, IsAllowedToSetField(entity, "Field2", []string{"user"}))

	// Test case 3: Check field with unauthorized role
	assert.False(t, IsAllowedToSetField(entity, "Field1", []string{"guest"}))
	assert.False(t, IsAllowedToSetField(entity, "Field2", []string{"guest"}))

	// Test case 4: Check invalid field
	assert.False(t, IsAllowedToSetField(entity, "InvalidField", []string{"admin"}))
}

func TestIsAllowedToSetFieldWithWildcardAndNegation(t *testing.T) {
	entity := &RoleBasedStruct{} // Assume this is already defined with the appropriate struct tags

	assert.True(t, IsAllowedToSetField(entity, "PublicField", []string{"admin"}), "Admin should access PublicField with wildcard")
	assert.True(t, IsAllowedToSetField(entity, "PublicField", []string{"user"}), "User should access PublicField with wildcard")
	assert.True(t, IsAllowedToSetField(entity, "PublicField", []string{"guest"}), "Guest should access PublicField with wildcard")

	// Assuming "AdminField" is tagged with "!user"
	assert.True(t, IsAllowedToSetField(entity, "AdminField", []string{"admin"}), "Admin should access AdminField")
	assert.False(t, IsAllowedToSetField(entity, "AdminField", []string{"user"}), "User should not access AdminField with negation")
}

func TestFilterStringSlice(t *testing.T) {
	type TestStruct struct {
		StringSlice []string `json:"string_slice" writexs:"admin" readxs:"admin"`
	}

	entity := &TestStruct{
		StringSlice: []string{"initial1", "initial2"},
	}

	updateEntity := map[string]any{
		"string_slice": []string{"updated1", "updated2"},
	}

	xsList := []string{"admin"}

	// Test case 1: Update string slice
	filtered, err := FilterMapFieldsByStructAndRole(entity, updateEntity, xsList, true, true)
	assert.NoError(t, err)
	filteredField, ok := filtered["string_slice"]
	assert.True(t, ok)
	filteredSlice, ok := filteredField.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"updated1", "updated2"}, filteredSlice)
	assert.Equal(t, []string{"initial1", "initial2"}, entity.StringSlice)

	// Test case 2: Update string slice with invalid type
	updateEntity["string_slice"] = "invalid"
	filtered, err = FilterMapFieldsByStructAndRole(entity, updateEntity, xsList, true, true)
	assert.Error(t, err)
	assert.Nil(t, filtered)

	// Test case 3: Update string slice that is nil before update
	entity.StringSlice = nil
	updateEntity["string_slice"] = []string{"updated1", "updated2"}
	filtered, err = FilterMapFieldsByStructAndRole(entity, updateEntity, xsList, true, true)
	assert.NoError(t, err)
	filteredField, ok = filtered["string_slice"]
	assert.True(t, ok)
	filteredSlice, ok = filteredField.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"updated1", "updated2"}, filteredSlice)
	assert.Nil(t, entity.StringSlice)

	// Test case 4: Update string slice with invalid type
	entity.StringSlice = []string{"initial1", "initial2"}
	updateEntity["string_slice"] = 123
	filtered, err = FilterMapFieldsByStructAndRole(entity, updateEntity, xsList, true, true)
	assert.Error(t, err)
	assert.Nil(t, filtered)

	// Test case 5: Update string slice with struct that is marshalled to JSON
	entity.StringSlice = []string{"initial1", "initial2"}
	updateStruct := TestStruct{
		StringSlice: []string{"updated1", "updated2"},
	}
	jsonBytes, err := json.Marshal(updateStruct)
	assert.NoError(t, err)
	var updateMap map[string]any
	err = json.Unmarshal(jsonBytes, &updateMap)
	assert.NoError(t, err)
	// list the types of the map fields
	for k, v := range updateMap {
		log.Printf("key: %v, value: %v, type: %T", k, v, v)
	}
	filtered, err = FilterMapFieldsByStructAndRole(entity, updateMap, xsList, true, true)
	assert.Error(t, err)
	assert.Nil(t, filtered)
	assert.Equal(t, err.Error(), "type mismatch between struct and source map field: StringSlice")
	anySlice, ok := updateMap["string_slice"].([]any)
	assert.True(t, ok)
	typedSlice, err := ConvertSliceTypeFromAnyTo[string](anySlice, true)
	assert.NoError(t, err)
	updateMap["string_slice"] = typedSlice
	filtered, err = FilterMapFieldsByStructAndRole(entity, updateMap, xsList, true, true)
	assert.NoError(t, err)
	filteredField, ok = filtered["string_slice"]
	assert.True(t, ok)
	filteredSlice, ok = filteredField.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"updated1", "updated2"}, filteredSlice)
	assert.Equal(t, []string{"initial1", "initial2"}, entity.StringSlice)
}

// test ConvertMapFieldsToTypedSlices
func TestConvertMapFieldsToTypedSlices(t *testing.T) {
	type TestStruct struct {
		StringSlice  []string      `json:"string_slice"`
		IntSlice     []int         `json:"int_slice"`
		FloatSlice   []float64     `json:"float_slice"`
		MixedField   []interface{} `json:"mixed_field"`
		SingleString string        `json:"single_string"`
		SingleInt    int           `json:"single_int"`
	}

	tests := []struct {
		name                string
		input               map[string]interface{}
		ignoreNonAssignable bool
		want                map[string]interface{}
		wantErr             bool
	}{
		{
			name: "Happy path - all fields convert correctly",
			input: map[string]interface{}{
				"string_slice":  []interface{}{"a", "b", "c"},
				"int_slice":     []interface{}{1, 2, 3},
				"float_slice":   []interface{}{1.1, 2.2, 3.3},
				"mixed_field":   []interface{}{1, "two", true},
				"single_string": "hello",
				"single_int":    42,
			},
			ignoreNonAssignable: false,
			want: map[string]interface{}{
				"string_slice":  []string{"a", "b", "c"},
				"int_slice":     []int{1, 2, 3},
				"float_slice":   []float64{1.1, 2.2, 3.3},
				"mixed_field":   []interface{}{1, "two", true},
				"single_string": "hello",
				"single_int":    42,
			},
			wantErr: false,
		},
		{
			name: "Ignore non-assignable values",
			input: map[string]interface{}{
				"string_slice": []interface{}{"a", "b", 3, "c"},
				"int_slice":    []interface{}{1, "two", 3},
			},
			ignoreNonAssignable: true,
			want: map[string]interface{}{
				"string_slice": []string{"a", "b", "c"},
				"int_slice":    []int{1, 3},
			},
			wantErr: false,
		},
		{
			name: "Error on non-assignable values",
			input: map[string]interface{}{
				"string_slice": []interface{}{"a", "b", 3, "c"},
			},
			ignoreNonAssignable: false,
			want:                nil,
			wantErr:             true,
		},
		{
			name: "Handle empty slices",
			input: map[string]interface{}{
				"string_slice": []interface{}{},
				"int_slice":    []interface{}{},
			},
			ignoreNonAssignable: false,
			want: map[string]interface{}{
				"string_slice": []string{},
				"int_slice":    []int{},
			},
			wantErr: false,
		},
		{
			name: "Handle missing fields",
			input: map[string]interface{}{
				"string_slice": []interface{}{"a", "b", "c"},
				// int_slice is missing
			},
			ignoreNonAssignable: false,
			want: map[string]interface{}{
				"string_slice": []string{"a", "b", "c"},
			},
			wantErr: false,
		},
		{
			name: "Handle extra fields in input",
			input: map[string]interface{}{
				"string_slice": []interface{}{"a", "b", "c"},
				"extra_field":  "should be ignored",
			},
			ignoreNonAssignable: false,
			want: map[string]interface{}{
				"string_slice": []string{"a", "b", "c"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertMapFieldsToTypedSlices(tt.input, TestStruct{}, tt.ignoreNonAssignable)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertMapFieldsToTypedSlices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertMapFieldsToTypedSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertMapFieldsToTypedSlices_EdgeCases(t *testing.T) {
	t.Run("Non-struct target", func(t *testing.T) {
		_, err := ConvertMapFieldsToTypedSlices(map[string]interface{}{}, "not a struct", false)
		if err == nil {
			t.Errorf("Expected error when passing non-struct target, got nil")
		}
	})

	t.Run("Nil input map", func(t *testing.T) {
		type EmptyStruct struct{}
		_, err := ConvertMapFieldsToTypedSlices(nil, EmptyStruct{}, false)
		if err != nil {
			t.Errorf("Unexpected error with nil input map: %v", err)
		}
	})

	t.Run("Empty struct", func(t *testing.T) {
		type EmptyStruct struct{}
		got, err := ConvertMapFieldsToTypedSlices(map[string]interface{}{"foo": "bar"}, EmptyStruct{}, false)
		if err != nil {
			t.Errorf("Unexpected error with empty struct: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("Expected empty result map, got %v", got)
		}
	})
}
