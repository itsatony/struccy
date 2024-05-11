package struccy

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const Version = "1.3.0"

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
	ErrNilArguments                = errors.New("one or more arguments are nil")
	ErrInvalidDstSlice             = errors.New("destination must be a slice")
	ErrDifferentStructType         = errors.New("source and destination must be the same struct type")
	ErrFieldMismatch               = errors.New("source and destination structs have different number of fields")
	ErrZeroValueDisallowed         = errors.New("zero value is not allowed for this field")
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
func MergeMapStringFieldsToStruct(targetStruct any, updateMap map[string]any, xsList []string) (any, error) {
	targetValue := reflect.ValueOf(targetStruct)

	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return targetStruct, ErrTargetStructMustBePointer
	}

	targetValue = targetValue.Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldName := field.Name
		fieldValue := targetValue.Field(i)

		updateValue, ok := updateMap[fieldName]
		if !ok {
			continue
		}

		writexs := field.Tag.Get("xswrite")
		if !isFieldAccessAllowed(xsList, writexs) {
			continue
		}

		updateValueType := reflect.TypeOf(updateValue)

		if fieldValue.Kind() == reflect.Ptr {
			if updateValueType.Kind() != reflect.Ptr {
				if updateValueType.ConvertibleTo(fieldValue.Type().Elem()) {
					// Target field is a pointer, but the updateMap value is not a pointer
					convertedValue := reflect.ValueOf(updateValue).Convert(fieldValue.Type().Elem())
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
					fieldValue.Elem().Set(convertedValue)
				} else {
					return targetStruct, fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, fieldName, fieldValue.Type(), updateValueType)
				}
			} else {
				fieldValue.Set(reflect.ValueOf(updateValue))
			}
		} else {
			if updateValueType.AssignableTo(fieldValue.Type()) {
				fieldValue.Set(reflect.ValueOf(updateValue))
			} else if updateValueType.Kind() == reflect.Ptr && updateValueType.Elem().AssignableTo(fieldValue.Type()) {
				// Target field is not a pointer, but the updateMap value is a pointer
				fieldValue.Set(reflect.ValueOf(updateValue).Elem())
			} else if updateValueType.ConvertibleTo(fieldValue.Type()) {
				fieldValue.Set(reflect.ValueOf(updateValue).Convert(fieldValue.Type()))
			} else {
				return targetStruct, fmt.Errorf("%w: %s, expected %v, got %v", ErrFieldTypeMismatch, fieldName, fieldValue.Type(), updateValueType)
			}
		}
	}

	return targetStruct, nil
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
			readxs := field.Tag.Get("readxs")
			if !isFieldAccessAllowed(xsList, readxs) {
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
		readXS := field.Tag.Get("readxs")
		if isFieldAccessAllowed(xsList, readXS) {
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
		writeXS := field.Tag.Get("writexs")
		if isFieldAccessAllowed(xsList, writeXS) {
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
		readXS := field.Tag.Get("readxs")
		if isFieldAccessAllowed(xsList, readXS) {
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
		writeXS := field.Tag.Get("writexs")
		if isFieldAccessAllowed(xsList, writeXS) {
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

// helpers

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
