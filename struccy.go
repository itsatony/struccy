package struccy

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const Version = "1.0.0"

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
			if sourceField.Kind() == reflect.Ptr && filteredField.Kind() != reflect.Ptr &&
				sourceField.Type().Elem() == filteredField.Type() {
				// Source field is a pointer and filtered field is not, but the underlying types match
				if !sourceField.IsNil() {
					filteredField.Set(sourceField.Elem())
				} else if !zeroDisallowed {
					filteredField.Set(reflect.Zero(filteredField.Type()))
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
