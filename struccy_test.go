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
