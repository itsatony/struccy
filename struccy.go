package struccy

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const Version = "1.5.4"

const (
	tagNameReadXS  = "readxs"
	tagNameWriteXS = "writexs"
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
	ErrFieldTypeMismatch           = errors.New("field type mismatch")
	ErrInvalidStructPointer        = errors.New("invalid struct pointer")
	ErrUnexportedField             = errors.New("unexported field")
	ErrJSONMarshalFailed           = errors.New("JSON marshal failed")
	ErrNilArguments                = errors.New("one or more arguments are nil")
	ErrInvalidDstSlice             = errors.New("destination must be a slice")
	ErrDifferentStructType         = errors.New("source and destination must be the same struct type")
	ErrFieldMismatch               = errors.New("source and destination structs have different number of fields")
	ErrZeroValueDisallowed         = errors.New("zero value is not allowed for this field")
	ErrInvalidFieldName            = errors.New("field name not found in source map")
	ErrTypeMismatch                = errors.New("type mismatch between struct and source map field")
	ErrMustBeStructPointer         = errors.New("reference must be a pointer to a struct")
	ErrUnauthorizedFieldSet        = errors.New("unauthorized field set")
	ErrInvalidFieldValue           = errors.New("invalid field value")
	ErrInvalidFieldType            = errors.New("invalid field type")
	ErrInvalidPtrType              = errors.New("invalid pointer type")
	ErrFieldIsNil                  = errors.New("field value is nil")
)

// MergeStructUpdateTo merges the fields of a source struct into a destination struct.
// It takes a pointer to the source struct and a pointer to the destination struct.
// The function iterates over the fields of the destination struct and looks for corresponding fields
// in the source struct. If a matching field is found, the value from the source struct is assigned
// to the field in the destination struct.
//
// The function handles the following cases:
//   - If a field in the destination struct is not found in the source struct, it remains unchanged.
//   - If a field in the source struct is a pointer and the corresponding field in the destination struct is not,
//     the dereferenced value of the source field is assigned to the destination field.
//   - If a field in the source struct is not a pointer and the corresponding field in the destination struct is,
//     a new pointer is created with the same type as the source field, and the value of the source field is assigned
//     to the dereferenced value of the destination field.
//   - If a field in the source struct is a pointer and it is nil, the corresponding field in the destination struct
//     is set to its zero value.
//
// The function returns an error if:
// - The source or destination struct is not a pointer to a struct.
// - The types of the corresponding fields in the source and destination structs do not match.
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

		writexs := field.Tag.Get(tagNameWriteXS)
		if !IsFieldAccessAllowed(xsList, writexs) {
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

// MergeMapStringFieldsToStruct merges the fields from a map[string]any into a target struct.
// The function takes a pointer to the target struct, a map[string]any representing the fields to update,
// and a list of allowed field names (xsList) for merging.
//
// The function iterates over the fields of the target struct and checks if there are corresponding entries
// in the updateMap. If a matching entry is found and the field name is allowed based on the xsList,
// the value from the updateMap is assigned to the struct field.
//
// The function handles the following cases:
//   - If a field in the target struct is not found in the updateMap, it remains unchanged.
//   - If a field in the updateMap is not found in the target struct, it is ignored.
//   - If the type of a field in the updateMap does not match the type of the corresponding struct field,
//     the function attempts to convert the value to the appropriate type.
//   - If the struct field is a pointer and the updateMap value is not a pointer,
//     the function creates a new pointer with the updateMap value.
//   - If the struct field is not a pointer and the updateMap value is a pointer,
//     the function dereferences the updateMap value.
//   - If a field is not allowed based on the xsList, it is skipped.
//
// The function returns the updated struct and an error if any of the following conditions are met:
// - The target struct is not a pointer to a struct.
// - An error occurs during the merging process.

// MergeMapStringFieldsToStruct merges the fields from a map[string]any into a target struct.
func MergeMapStringFieldsToStruct(targetStruct any, updateMap map[string]any, xsList []string) (any, error) {
	targetValue := reflect.ValueOf(targetStruct)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return nil, ErrTargetStructMustBePointer
	}

	structElem := targetValue.Elem()
	for key, updateValue := range updateMap {
		targetField := structElem.FieldByName(key)
		if !targetField.IsValid() {
			continue // Field not found in the struct
		}
		if !targetField.CanSet() {
			continue // Cannot set unexported fields
		}

		updateValueReflect := reflect.ValueOf(updateValue)
		if err := assignValueToField(targetField, updateValueReflect); err != nil {
			return nil, fmt.Errorf("error assigning field '%s': %w", key, err)
		}
	}

	return targetStruct, nil
}

// This function tries to assign values to struct fields while handling type conversions.
func assignValueToField(targetField, updateValueReflect reflect.Value) error {
	// First, check if the update value is valid (not a zero Value)
	if !updateValueReflect.IsValid() {
		if targetField.Kind() == reflect.Ptr {
			// If it's a pointer in the struct, set it to nil
			targetField.Set(reflect.Zero(targetField.Type()))
			return nil
		} else {
			// If it's not a pointer and we're trying to assign nil, that's an error
			return fmt.Errorf("cannot assign nil to non-pointer type %s", targetField.Type())
		}
	}

	// Handle if the update value is a pointer and the target field is not, or vice versa.
	if updateValueReflect.Kind() == reflect.Ptr {
		updateValueReflect = updateValueReflect.Elem() // Dereference pointers to their base value.
	}

	// Handle pointer fields in the struct.
	if targetField.Kind() == reflect.Ptr {
		// Handle initializing nil pointer fields if needed.
		if targetField.IsNil() && targetField.CanSet() {
			targetField.Set(reflect.New(targetField.Type().Elem()))
		}
		if updateValueReflect.Type().AssignableTo(targetField.Type().Elem()) {
			targetField.Elem().Set(updateValueReflect) // Assign compatible types directly.
		} else if checkTypeConvertible(updateValueReflect, targetField.Type().Elem()) {
			convertedValue := updateValueReflect.Convert(targetField.Type().Elem())
			targetField.Elem().Set(convertedValue) // Convert and assign if possible.
		} else {
			return ErrFieldTypeMismatch
		}
	} else {
		if updateValueReflect.Type().AssignableTo(targetField.Type()) {
			targetField.Set(updateValueReflect) // Direct assignment if types are compatible.
		} else if checkTypeConvertible(updateValueReflect, targetField.Type()) {
			convertedValue := updateValueReflect.Convert(targetField.Type())
			targetField.Set(convertedValue) // Perform conversion and assignment.
		} else {
			return ErrFieldTypeMismatch
		}
	}

	return nil
}

// Helper function to check if types are convertible.
func checkTypeConvertible(value reflect.Value, targetType reflect.Type) bool {
	// Implement rules for conversion here.
	// Example: int8 to int might need custom handling if you want to allow it.
	valueType := value.Type()
	if valueType.Kind() == reflect.Int8 && targetType.Kind() == reflect.Int {
		return true // Allow conversion from int8 to int.
	}
	return valueType.ConvertibleTo(targetType)
}

// FilterStructTo filters the fields of a source struct and assigns the allowed fields to a destination struct.
// It takes a pointer to the source struct, a pointer to the destination struct, a list of excluded fields (xsList),
// and a boolean flag (zeroDisallowed) indicating whether zero values are allowed for excluded fields.
//
// The function iterates over the fields of the destination struct and looks for corresponding fields
// in the source struct. If a matching field is found and it is not in the excluded fields list (xsList),
// the value from the source struct is assigned to the field in the destination struct.
//
// The function handles the following cases:
//   - If a field in the destination struct is not found in the source struct and zeroDisallowed is true,
//     the field in the destination struct is set to its zero value.
//   - If a field in the source struct is a pointer and the corresponding field in the destination struct is not,
//     the dereferenced value of the source field is assigned to the destination field if it is not nil.
//     If the source field is nil and zeroDisallowed is false, the destination field is set to its zero value.
//   - If a field in the source struct is not a pointer and the corresponding field in the destination struct is,
//     a new pointer is created with the same type as the source field, and the value of the source field is assigned
//     to the dereferenced value of the destination field.
//   - If a field in the source struct is a pointer and it is not nil, the destination field is set to the source field.
//     If the source field is nil and zeroDisallowed is false, the destination field is set to its zero value.
//
// The function returns an error if:
// - The source or destination struct is not a pointer to a struct.
// - The types of the corresponding fields in the source and destination structs do not match.
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
			if sourceField.Kind() == reflect.Ptr && filteredField.Kind() != reflect.Ptr {
				if sourceField.Type().Elem() == filteredField.Type() {
					// Source field is a pointer and filtered field is not, but the underlying types match
					if !sourceField.IsNil() {
						filteredField.Set(sourceField.Elem())
					} else if !zeroDisallowed {
						filteredField.Set(reflect.Zero(filteredField.Type()))
					}
				} else if sourceField.Type().Elem().Kind() == reflect.Slice && filteredField.Type().Kind() == reflect.Slice &&
					sourceField.Type().Elem().Elem() == filteredField.Type().Elem() {
					// Source field is a pointer to a slice and filtered field is a non-pointer slice
					if !sourceField.IsNil() {
						filteredField.Set(reflect.ValueOf(sourceField.Elem().Interface()))
					} else if !zeroDisallowed {
						filteredField.Set(reflect.Zero(filteredField.Type()))
					}
				} else {
					return fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, field.Name, filteredField.Type(), sourceField.Type())
				}
			} else if sourceField.Kind() != reflect.Ptr && filteredField.Kind() == reflect.Ptr &&
				sourceField.Type() == filteredField.Type().Elem() {
				// Source field is not a pointer and filtered field is a pointer, but the underlying types match
				filteredField.Set(reflect.New(sourceField.Type()))
				filteredField.Elem().Set(sourceField)
			} else {
				return fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, field.Name, filteredField.Type(), sourceField.Type())
			}
		} else {
			readxs := field.Tag.Get(tagNameReadXS)
			if !IsFieldAccessAllowed(xsList, readxs) {
				if zeroDisallowed {
					filteredField.Set(reflect.Zero(filteredField.Type()))
				}
				continue
			}

			if sourceField.Kind() == reflect.Ptr {
				if !sourceField.IsNil() {
					filteredField.Set(sourceField)
				} else if !zeroDisallowed {
					filteredField.Set(reflect.Zero(filteredField.Type()))
				}
			} else {
				filteredField.Set(sourceField)
			}
		}
	}

	return nil
}

// FilterMapFieldsToStruct filters the fields of a source map and assigns the allowed fields to a destination struct.
// It takes a map of string field names and values and a list of allowed field names (xsList).
// The function iterates over the fields of the destination struct and looks for corresponding entries in the source map.
// If a matching entry is found and the field name is allowed based on the xsList, the value from the source map is assigned
func FilterMapFieldsByRole(source map[string]any, xsList []string) (filtered map[string]any, err error) {
	filtered = make(map[string]any)
	allowedFieldNames, err := GetFieldNamesWithWriteXS(source, xsList)
	if err != nil {
		return nil, err
	}
	for _, fieldName := range allowedFieldNames {
		filtered[fieldName] = source[fieldName]
	}
	return filtered, nil
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

// GetFieldNames returns a slice of field names for the given struct pointer.
// It uses reflection to iterate over the fields of the struct and collect their names.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func GetFieldNames(structPtr any) ([]string, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structType := structValue.Elem().Type()
	numFields := structType.NumField()

	fieldNames := make([]string, numFields)
	for i := 0; i < numFields; i++ {
		fieldNames[i] = structType.Field(i).Name
	}

	return fieldNames, nil
}

// GetFieldNamesWithReadXS returns a slice of field names for the given struct pointer,
// filtered by the specified read access rules (xsList).
// It uses reflection to iterate over the fields of the struct and collect the names
// of fields that have read access allowed based on the provided xsList.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func GetFieldNamesWithReadXS(structPtr any, xsList []string) ([]string, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structType := structValue.Elem().Type()
	numFields := structType.NumField()

	fieldNames := make([]string, 0)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		readXS := field.Tag.Get(tagNameReadXS)
		if IsFieldAccessAllowed(xsList, readXS) {
			fieldNames = append(fieldNames, field.Name)
		}
	}

	return fieldNames, nil
}

// GetFieldNamesWithWriteXS returns a slice of field names for the given struct pointer,
// filtered by the specified write access rules (xsList).
// It uses reflection to iterate over the fields of the struct and collect the names
// of fields that have write access allowed based on the provided xsList.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func GetFieldNamesWithWriteXS(structPtr any, xsList []string) ([]string, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structType := structValue.Elem().Type()
	numFields := structType.NumField()

	fieldNames := make([]string, 0)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		writeXS := field.Tag.Get(tagNameWriteXS)
		if IsFieldAccessAllowed(xsList, writeXS) {
			fieldNames = append(fieldNames, field.Name)
		}
	}

	return fieldNames, nil
}

// StructToMapFieldsWithReadXS converts the specified struct pointer to a map,
// including only the fields with read access allowed based on the provided xsList.
// It uses reflection to iterate over the fields of the struct and collect the field
// names and values that have read access allowed.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func StructToMapFieldsWithReadXS(structPtr any, xsList []string) (map[string]any, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structValue = structValue.Elem()
	structType := structValue.Type()
	numFields := structType.NumField()

	fieldMap := make(map[string]any)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		readXS := field.Tag.Get(tagNameReadXS)
		if IsFieldAccessAllowed(xsList, readXS) {
			fieldMap[field.Name] = structValue.Field(i).Interface()
		}
	}

	return fieldMap, nil
}

// StructToMapFieldsWithWriteXS converts the specified struct pointer to a map,
// including only the fields with write access allowed based on the provided xsList.
// It uses reflection to iterate over the fields of the struct and collect the field
// names and values that have write access allowed.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func StructToMapFieldsWithWriteXS(structPtr any, xsList []string) (map[string]any, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structValue = structValue.Elem()
	structType := structValue.Type()
	numFields := structType.NumField()

	fieldMap := make(map[string]any)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		writeXS := field.Tag.Get(tagNameWriteXS)
		if IsFieldAccessAllowed(xsList, writeXS) {
			fieldMap[field.Name] = structValue.Field(i).Interface()
		}
	}

	return fieldMap, nil
}

// StructToJSONFieldsWithReadXS converts the specified struct pointer to a JSON string,
// including only the fields with read access allowed based on the provided xsList.
// It uses reflection to iterate over the fields of the struct and collect the field
// names and values that have read access allowed, and then marshals the resulting map
// to a JSON string.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func StructToJSONFieldsWithReadXS(structPtr any, xsList []string) (string, error) {
	fieldMap, err := StructToMapFieldsWithReadXS(structPtr, xsList)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(fieldMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// StructToJSONFieldsWithWriteXS converts the specified struct pointer to a JSON string,
// including only the fields with write access allowed based on the provided xsList.
// It uses reflection to iterate over the fields of the struct and collect the field
// names and values that have write access allowed, and then marshals the resulting map
// to a JSON string.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func StructToJSONFieldsWithWriteXS(structPtr any, xsList []string) (string, error) {
	fieldMap, err := StructToMapFieldsWithWriteXS(structPtr, xsList)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(fieldMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// StructToMapFields takes a pointer to a struct and a slice of field names,
// and returns a map of the struct fields filtered to the specified field names.
//
// The function uses reflection to iterate over the fields of the struct and
// retrieve the values of the specified fields. It creates a map with the field
// names as keys and the corresponding field values as values.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
//
// If a specified field name does not exist in the struct, it is silently ignored.
func StructToMapFields(structPtr any, fieldNames []string) (map[string]any, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structValue = structValue.Elem()
	structType := structValue.Type()

	fieldMap := make(map[string]any)

	for _, fieldName := range fieldNames {
		field, ok := structType.FieldByName(fieldName)
		if !ok {
			continue
		}

		fieldValue := structValue.FieldByName(fieldName)
		fieldMap[field.Name] = fieldValue.Interface()
	}

	return fieldMap, nil
}

// StructToMap takes a pointer to a struct and returns a map of the struct fields.
//
// The function uses reflection to iterate over all the fields of the struct and
// retrieve their names and values. It creates a map with the field names as keys
// and the corresponding field values as values.
//
// If the provided `structPtr` is not a pointer to a struct, the function returns
// an error (`ErrInvalidStructPointer`).
func StructToMap(structPtr any) (map[string]any, error) {
	structValue := reflect.ValueOf(structPtr)

	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidStructPointer
	}

	structValue = structValue.Elem()
	structType := structValue.Type()

	fieldMap := make(map[string]any)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		fieldMap[field.Name] = fieldValue.Interface()
	}

	return fieldMap, nil
}

func IsFieldAccessAllowed(roles []string, tagValue string) bool {
	// Handle the wildcard which grants access to any role.
	if tagValue == "*" {
		return true
	}
	// fmt.Printf("Roles: %v\n", roles)
	// Split the permissions from the tag into individual roles.
	taggedRoles := strings.Split(tagValue, ",")

	finallyAllowed := false

	// Check each role against the allowed roles from the tag.
	for _, role := range roles {
		for _, taggedRole := range taggedRoles {
			// Direct match grants access.
			if role == taggedRole {
				finallyAllowed = true
				continue
			}

			// Handle negation, which explicitly denies access if the role matches.
			if strings.HasPrefix(taggedRole, "!") {
				// once a negation is mentioned, we assume that the default is to allow access if the denied role is not found
				finallyAllowed = true
				// If a role is denied, access is denied.
				if role == strings.TrimPrefix(taggedRole, "!") {
					return false
				}
			}
		}
	}

	// If no roles match, access is denied.
	return finallyAllowed
}

// @godoc FilterMapFieldsByStructAndRole filters the fields of a source map based on the fields of a reference struct.
// It takes a pointer to the reference struct, a map of string field names and values, a list of allowed field names (xsList),
// a flag indicating whether to ignore nil values, and a flag indicating whether to use JSON field names.
// The function iterates over the allowed field names and checks if the field exists in the source map.
// If a matching field is found, the value is checked for compatibility with the reference struct field type.
// If the value is compatible, it is added to the filtered map with the field name as the key.
// If the value is nil and ignoreNils is true, the field is skipped.
// If the value is nil and ignoreNils is false, an error is returned.
// If the value is not compatible with the reference struct field type, an error is returned.
// If the field name is not found in the reference struct, it is skipped.
// If the field name is not found in the source map, it is skipped.
// If the field name is found in the reference struct but not in the source map, an error is returned.
// If the field name is found in the source map but not in the reference struct, an error is returned.
// If the reference struct is not a pointer to a struct, an error is returned.
// If any error occurs during the process, the function returns an error.
func FilterMapFieldsByStructAndRole(referenceStructPointer any, source map[string]any, xsList []string, ignoreNils bool, useJsonFieldNames bool) (filtered map[string]any, err error) {
	filtered = make(map[string]any)
	allowedFieldNames, err := GetFieldNamesWithWriteXS(referenceStructPointer, xsList)
	if err != nil {
		return nil, err
	}

	refVal := reflect.ValueOf(referenceStructPointer)
	if refVal.Kind() != reflect.Ptr || refVal.Elem().Kind() != reflect.Struct {
		return nil, ErrMustBeStructPointer
	}
	refType := refVal.Elem().Type()

	jsonToFieldName := make(map[string]string)
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0] // Get the first part of the JSON tag
		if jsonTag != "" && jsonTag != "-" {
			jsonToFieldName[field.Name] = jsonTag
		}
	}

	for _, fieldName := range allowedFieldNames {
		jsonFieldName := fieldName
		if actualName, ok := jsonToFieldName[fieldName]; ok {
			jsonFieldName = actualName
		}

		fieldVal, ok := source[jsonFieldName]
		if !ok {
			continue // Skip fields not found in the source map, no error needed
		}

		structField, found := refType.FieldByName(fieldName)
		if !found {
			continue // Skip fields not found in the struct, no error needed
		}

		// Check if the field value is Nil
		if fieldVal == nil {
			if ignoreNils {
				continue
			}
			return nil, fmt.Errorf("%w: %s", ErrFieldIsNil, fieldName)
		}

		// Check if the field value is compatible with the struct field type
		if !isCompatibleType(fieldVal, structField.Type) {
			return nil, fmt.Errorf("%w: %s", ErrTypeMismatch, fieldName)
		}

		// Finally, set the field name in the filtered map
		if useJsonFieldNames {
			filtered[jsonFieldName] = fieldVal
		} else {
			filtered[fieldName] = fieldVal
		}
	}
	return filtered, nil
}

func isCompatibleType(value any, targetType reflect.Type) bool {
	valueType := reflect.TypeOf(value)

	// Check if the types are directly assignable
	if valueType.AssignableTo(targetType) {
		return true
	}

	// Check if the value can be converted to the target type
	if valueType.ConvertibleTo(targetType) {
		return true
	}

	// fmt.Printf("FieldVal(%v) FieldType(%v) != StructField.Type(%v). NOT COMPATIBLE\n", valueType, valueType, targetType)
	return false
}

// UpdateStructFields updates the fields of the given entity with the corresponding non-zero fields from the incomingEntity.
// It returns a map of the updated fields and their values. If a field is not authorized to be set based on the setterRole,
// it is skipped without returning an error. If any other error occurs during the field setting, the function returns an error.
//
// Parameters:
//   - entity: a pointer to the struct to be updated
//   - incomingEntity: a pointer to the struct containing the fields to update
//   - setterRole: the role of the setter, used for authorization checks
//   - skipZeroVals: a flag indicating whether zero values should be skipped
//   - ignoreUnsettables: a flag indicating whether to ignore unsettable fields or throw an error upon attempt
//
// Returns:
//   - A map of the updated field names and their corresponding values
//   - An error if any error occurs during the field setting (except for unauthorized fields)
func UpdateStructFields(entity any, incomingEntity any, roles []string, skipZeroVals bool, ignoreUnsettables bool) (updatedFields map[string]any, unsettableFields map[string]any, err error) {
	updatedFields = make(map[string]any)
	unsettableFields = make(map[string]any)
	incomingValue := reflect.ValueOf(incomingEntity).Elem()
	entityType := reflect.TypeOf(entity).Elem()

	for i := 0; i < incomingValue.NumField(); i++ {
		entityField := entityType.Field(i)
		incomingField := incomingValue.FieldByName(entityField.Name)
		if !incomingField.IsValid() {
			continue
		}
		// Check if the field is settable and authorized
		if IsAllowedToSetField(entity, entityField.Name, roles) {
			fieldValue := incomingField.Interface()
			err := SetField(entity, entityField.Name, fieldValue, skipZeroVals, roles)
			if err == nil {
				updatedFields[entityField.Name] = fieldValue
			} else {
				unsettableFields[entityField.Name] = fieldValue
				if ignoreUnsettables {
					continue
				}
				return nil, unsettableFields, err
			}
		}
	}
	return updatedFields, unsettableFields, nil
}

// SetField sets a field on a struct pointer, with validation and authorization checks based on the setterRole.
//
// Parameters:
//   - entity: a pointer to the struct
//   - fieldName: the name of the field to set
//   - value: the value to set the field to
//   - setterRole: the role of the setter, used for authorization checks
//
// Returns:
//   - An error if the entity is not a pointer to a struct, the field is invalid, the setter is not authorized,
//     or the value type is not convertible to the field type
func SetField(entity any, fieldName string, value any, skipZeroVals bool, roles []string) error {
	// fmt.Printf("SetField: FieldName: (%s), Value: (%v)\n", fieldName, value)
	rv := reflect.ValueOf(entity)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return ErrInvalidStructPointer
	}
	rv = rv.Elem()
	field := rv.FieldByName(fieldName)
	if !field.IsValid() {
		return ErrInvalidFieldName
	}
	if !IsAllowedToSetField(entity, fieldName, roles) {
		return ErrUnauthorizedFieldSet
	}
	val := reflect.ValueOf(value)
	if (val.Kind() == reflect.Ptr && val.IsNil()) || val.IsZero() {
		// Skip nil assignments without an error
		// fmt.Printf("Skip nil/Zero assignment for field(%s) without an error\n", fieldName)
		return nil
	}
	return setReflectField(field, value) // Set the field value
}

func setReflectField(field reflect.Value, value any) error {
	fieldType := field.Type()
	val := reflect.ValueOf(value)
	// Check if the value type is convertible to the field type
	// for all cases, where we have a field type of X and a value type of *X,
	// like string and a value type of *string or int and a value type of *int,
	// we will assign the field to a copy of the value of the pointer as after ensuring that the value is not nil
	// fmt.Printf("Field(%s) Type: (%v) vs. Value-Type:(%v)\n", fieldName, fieldType, val.Type())
	if val.Type().AssignableTo(fieldType) {
		field.Set(val)
		return nil
	} else {
		// fmt.Printf("#notAssignableOuter Field(%s) Type: (%v) vs. Value-Type:(%v)\n", fieldName, fieldType, val.Type())
		if val.Kind() == reflect.Ptr && val.Type().Elem().AssignableTo(fieldType) {
			// fmt.Printf("pointer conversion attempt for field(%s)\n", fieldName)
			if fieldType.Kind() == reflect.Ptr {
				field.Set(reflect.New(fieldType.Elem()))
				field.Elem().Set(val.Elem())
				return nil
			}
			if fieldType.Kind() == reflect.String {
				if val.Elem().Kind() == reflect.String {
					field.SetString(val.Elem().String())
					return nil
				}
			}
			if fieldType.Kind() == reflect.Int || fieldType.Kind() == reflect.Int8 || fieldType.Kind() == reflect.Int16 || fieldType.Kind() == reflect.Int32 || fieldType.Kind() == reflect.Int64 {
				if val.Elem().Kind() == reflect.Int || val.Elem().Kind() == reflect.Int8 || val.Elem().Kind() == reflect.Int16 || val.Elem().Kind() == reflect.Int32 || val.Elem().Kind() == reflect.Int64 {
					field.SetInt(val.Elem().Int())
					return nil
				}
			}
			if fieldType.Kind() == reflect.Uint || fieldType.Kind() == reflect.Uint8 || fieldType.Kind() == reflect.Uint16 || fieldType.Kind() == reflect.Uint32 || fieldType.Kind() == reflect.Uint64 {
				if val.Elem().Kind() == reflect.Uint || val.Elem().Kind() == reflect.Uint8 || val.Elem().Kind() == reflect.Uint16 || val.Elem().Kind() == reflect.Uint32 || val.Elem().Kind() == reflect.Uint64 {
					field.SetUint(val.Elem().Uint())
					return nil
				}
			}
			if fieldType.Kind() == reflect.Float32 || fieldType.Kind() == reflect.Float64 {
				if val.Elem().Kind() == reflect.Float32 || val.Elem().Kind() == reflect.Float64 {
					field.SetFloat(val.Elem().Float())
					return nil
				}
				// also convert any int or pointer to an int to a float type target field
				if val.Elem().Kind() == reflect.Int || val.Elem().Kind() == reflect.Int8 || val.Elem().Kind() == reflect.Int16 || val.Elem().Kind() == reflect.Int32 || val.Elem().Kind() == reflect.Int64 {
					field.SetFloat(float64(val.Elem().Int()))
					return nil
				}
			}
			if fieldType.Kind() == reflect.Bool {
				if val.Elem().Kind() == reflect.Bool {
					field.SetBool(val.Elem().Bool())
					return nil
				}
			}
			// slice to slice
			if fieldType.Kind() == reflect.Slice && val.Elem().Kind() == reflect.Slice {
				if val.Elem().Type().Elem().AssignableTo(fieldType.Elem()) {
					field.Set(val.Elem())
					return nil
				}
			}
			// slice to *slice
			if fieldType.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Slice {
				if val.Elem().Type().Elem().AssignableTo(fieldType.Elem()) {
					field.Set(val.Elem())
					return nil
				}
			}
			// map to map
			if fieldType.Kind() == reflect.Map && val.Elem().Kind() == reflect.Map {
				if val.Elem().Type().Key().AssignableTo(fieldType.Key()) && val.Elem().Type().Elem().AssignableTo(fieldType.Elem()) {
					field.Set(val.Elem())
					return nil
				}
			}
			// map to *map
			if fieldType.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Map {
				if val.Elem().Type().Key().AssignableTo(fieldType.Elem().Key()) && val.Elem().Type().Elem().AssignableTo(fieldType.Elem().Elem()) {
					field.Set(val.Elem())
					return nil
				}
			}
			// struct to struct
			if fieldType.Kind() == reflect.Struct && val.Elem().Kind() == reflect.Struct {
				if val.Elem().Type().AssignableTo(fieldType) {
					field.Set(val.Elem())
					return nil
				}
			}
			// struct to *struct
			if fieldType.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
				if val.Elem().Type().AssignableTo(fieldType.Elem()) {
					field.Set(val.Elem())
					return nil
				}
			}
			return ErrInvalidFieldType
		}
		if !val.Type().ConvertibleTo(fieldType) {
			if fieldType.Kind() == reflect.Ptr && val.Type().AssignableTo(fieldType.Elem()) {
				return ErrInvalidPtrType
			} else if fieldType.Kind() == reflect.Int || fieldType.Kind() == reflect.Int8 || fieldType.Kind() == reflect.Int16 || fieldType.Kind() == reflect.Int32 || fieldType.Kind() == reflect.Int64 || fieldType.Kind() == reflect.Uint || fieldType.Kind() == reflect.Uint8 || fieldType.Kind() == reflect.Uint16 || fieldType.Kind() == reflect.Uint32 || fieldType.Kind() == reflect.Uint64 {
				if !canConvertInt(val) {
					// fmt.Printf("!canConvertInt --> Field(%s) Type: (%v) vs. Value-Type:(%v)\n", fieldName, fieldType, val.Type())
					return ErrInvalidFieldType
				} else {
					convertedValue, ok := tryConvertInt(val, fieldType)
					if ok {
						field.Set(convertedValue)
						// finalValue := field.Interface()
						// fmt.Printf("Field(%s) Type: (%v) vs. Value-Type:(%v) --> Set value to(%v)\n", fieldName, fieldType, val.Type(), finalValue)
						return nil
					}
				}
			} else if fieldType.Kind() == reflect.Float32 || fieldType.Kind() == reflect.Float64 {
				if !canConvertFloat(val) {
					// fmt.Printf("!canConvertFloat --> Field(%s) Type: (%v) vs. Value-Type:(%v)\n", fieldName, fieldType, val.Type())
					return ErrInvalidFieldType
				} else {
					convertedValue, ok := tryConvertFloat(val, fieldType)
					if ok {
						field.Set(convertedValue)
						// finalValue := field.Interface()
						// fmt.Printf("Field(%s) Type: (%v) vs. Value-Type:(%v) --> Set value to(%v)\n", fieldName, fieldType, val.Type(), finalValue)
						return nil
					}
				}
			} else if val.Type().Kind() == reflect.Ptr && val.Type().Elem().AssignableTo(fieldType) {
				// fmt.Printf("#notAssignableINNER Field Type: (%v) vs. Value-Type:(%v)\n", fieldType, val.Type())
				return ErrInvalidFieldType
			}
		} else {
			if fieldType.Kind() == reflect.Ptr && val.Type().AssignableTo(fieldType.Elem()) {
				// Field is a pointer and value is assignable to the underlying type
				field.Set(reflect.New(fieldType.Elem()))
				field.Elem().Set(val)
			} else if val.Type().Kind() == reflect.Ptr && val.Type().Elem().AssignableTo(fieldType) {
				// Value is a pointer and its underlying type is assignable to the field type
				field.Set(val.Elem())
			} else if val.Type().ConvertibleTo(fieldType) {
				field.Set(val.Convert(fieldType))
			} else if convertedValue, ok := tryConvertInt(val, fieldType); ok {
				field.Set(convertedValue)
			} else if convertedValue, ok := tryConvertFloat(val, fieldType); ok {
				field.Set(convertedValue)
			} else {
				// fmt.Printf("#fdgf Field Type: (%v) vs. Value-Type:(%v)\n", fieldType, val.Type())
				return ErrInvalidFieldValue
			}
			return ErrInvalidFieldType
		}
	}

	// field.Set(val.Convert(fieldType))

	// if fieldType.Kind() == reflect.Ptr && val.Type().AssignableTo(fieldType.Elem()) {
	// 	// Field is a pointer and value is assignable to the underlying type
	// 	field.Set(reflect.New(fieldType.Elem()))
	// 	field.Elem().Set(val)
	// } else if val.Type().Kind() == reflect.Ptr && val.Type().Elem().AssignableTo(fieldType) {
	// 	// Value is a pointer and its underlying type is assignable to the field type
	// 	field.Set(val.Elem())
	// } else if val.Type().ConvertibleTo(fieldType) {
	// 	field.Set(val.Convert(fieldType))
	// } else if convertedValue, ok := tryConvertInt(val, fieldType); ok {
	// 	field.Set(convertedValue)
	// } else if convertedValue, ok := tryConvertFloat(val, fieldType); ok {
	// 	field.Set(convertedValue)
	// } else {
	// 	return ErrInvalidFieldValue
	// }
	return nil
}

// IsAllowedToSetField checks if a field can be set based on the setter's role and the field's `writexs` tag.
//
// Parameters:
//   - entity: a pointer to the struct or the struct itself
//   - fieldName: the name of the field to check
//   - setterRole: the role of the setter
//
// Returns:
//   - A boolean indicating whether the field can be set by the given setter role
func IsAllowedToSetField(entity any, fieldName string, roles []string) bool {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		return false
	}
	return IsFieldAccessAllowed(roles, field.Tag.Get(tagNameWriteXS))
}

// tryConvertInt attempts to convert an integer value from one type to another
func tryConvertInt(val reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := val.Convert(reflect.TypeOf(int64(0))).Int()
		return reflect.ValueOf(intVal).Convert(targetType), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := val.Convert(reflect.TypeOf(uint64(0))).Uint()
		return reflect.ValueOf(uintVal).Convert(targetType), true
	default:
		return reflect.Value{}, false
	}
}

// tryConvertFloat attempts to convert a floating-point value from one type to another
func tryConvertFloat(val reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	if targetType.Kind() == reflect.Float32 || targetType.Kind() == reflect.Float64 {
		floatVal := val.Convert(reflect.TypeOf(float64(0))).Float()
		return reflect.ValueOf(floatVal).Convert(targetType), true
	}
	return reflect.Value{}, false
}

// canConvertInt checks if a value can be converted to an integer type
func canConvertInt(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

// canConvertFloat checks if a value can be converted to a floating-point type
func canConvertFloat(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
