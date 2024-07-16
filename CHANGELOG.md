# Changelog

All notable changes to the struccy package will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.5] - 2024-07-16

[1.5.5]: https://github.com/itsatony/struccy/releases/tag/v1.5.5

### Fixes

- @p0lymorph fixed a bug in UpdateStructFields

## [1.5.4] - 2024-07-09

[1.5.4]: https://github.com/itsatony/struccy/releases/tag/v1.5.4

## [1.5.2] - 2024-07-07

[1.5.2]: https://github.com/itsatony/struccy/releases/tag/v1.5.2

### Extended 1.5.2

- Extended the `FilterMapFieldsByStructAndRole` by a parameter to define whether the resulting map uses the go-fieldnames or the json-fieldnames.

## [1.5.1] - 2024-05-16

[1.5.1]: https://github.com/itsatony/struccy/releases/tag/v1.5.1

### Fixed 1.5.1

- FilterMapFieldsByStructAndRole did not work as expected and was updated and has a new signature.

## [1.5.0] - 2024-05-16

[1.5.0]: https://github.com/itsatony/struccy/releases/tag/v1.5.0

### Added 1.5.0

- Added more args to tightly control SetField and UpdateStructFields functions.
- Added lots of edge case SetField handling (like assigning *int8 to float64, etc.).

## [1.4.0] - 2024-05-13

[1.4.0]: https://github.com/itsatony/struccy/releases/tag/v1.4.0

### Added 1.4.0

- Worked under the assumption that this package will try to convert types where useful/helpful and safe. This includes int8->int or int8->int64 as well as *string->string, where the value behind the pointer is not nil. This makes sense in light of this package's purpose to convert structs to maps and vice versa, where incoming json data might be of a different type than the struct field.
- Introduced a new function `MergeMapStringFieldsToStruct` to merge map fields to a struct based on the provided xsList tags `xswrite`.
- Added new test cases to cover scenarios with map fields and xsList tags.
- Fixed a bug where map fields were not correctly merged to the struct in the `MergeMapStringFieldsToStruct` function.

### Changed 1.4.0

- Updated the `MergeStructUpdateTo` function to handle cases where the source field is a pointer to a slice and the destination field is a non-pointer slice.
- Modified the `FilterStructTo` function to correctly filter fields with pointer types to slices.
- Improved error handling and type mismatch detection in the `FilterStructTo` and `MergeStructUpdateTo` functions.

## [1.3.0] - 2024-05-11

[1.3.0]: https://github.com/itsatony/struccy/releases/tag/v1.3.0

### Added 1.3.0

- Introduced a new function `MergeMapStringFieldsToStruct` to merge map fields to a struct based xsList tags `xswrite`.

## [1.2.0] - 2024-05-11

[1.2.0]: https://github.com/itsatony/struccy/releases/tag/v1.2.0

### Added 1.2.0

- Introduced a set of convenience functions for working with struct fields and their read/write access rules (xsList).
  - `GetFieldNames` function to retrieve all field names of a struct.
  - `GetFieldNamesWithReadXS` function to retrieve field names with read access allowed based on the provided xsList.
  - `GetFieldNamesWithWriteXS` function to retrieve field names with write access allowed based on the provided xsList.
  - `StructToMapFieldsWithReadXS` function to convert a struct to a map, including only fields with read access allowed based on the provided xsList.
  - `StructToMapFieldsWithWriteXS` function to convert a struct to a map, including only fields with write access allowed based on the provided xsList.
  - `StructToJSONFieldsWithReadXS` function to convert a struct to a JSON string, including only fields with read access allowed based on the provided xsList.
  - `StructToJSONFieldsWithWriteXS` function to convert a struct to a JSON string, including only fields with write access allowed based on the provided xsList.
- Added documentation and usage examples for the new convenience functions in the README.md file.

This changelog entry for version 1.2.0 highlights the addition of several convenience functions to the `struccy` package. These functions provide additional utility and flexibility when working with struct fields and their read/write access rules.

The new functions include:

- `GetFieldNames`: Retrieves all field names of a struct.
- `GetFieldNamesWithReadXS`: Retrieves field names with read access allowed based on the provided xsList.
- `GetFieldNamesWithWriteXS`: Retrieves field names with write access allowed based on the provided xsList.
- `StructToMapFieldsWithReadXS`: Converts a struct to a map, including only fields with read access allowed based on the provided xsList.
- `StructToMapFieldsWithWriteXS`: Converts a struct to a map, including only fields with write access allowed based on the provided xsList.
- `StructToJSONFieldsWithReadXS`: Converts a struct to a JSON string, including only fields with read access allowed based on the provided xsList.
- `StructToJSONFieldsWithWriteXS`: Converts a struct to a JSON string, including only fields with write access allowed based on the provided xsList.

The README.md file has been updated with documentation and usage examples for these new convenience functions, providing guidance on how to use them effectively.

The changelog entry also includes a link to compare the changes between version 1.1.0 and 1.2.0 using the GitHub comparison URL.

## [1.1.0] - 2024-05-11

[1.1.0]: https://github.com/itsatony/struccy/releases/tag/v1.1.0

### Added 1.1.0

- Added support for filtering and merging structs with pointer fields to slices.
- Improved error handling and type mismatch detection in `FilterStructTo` and `MergeStructUpdateTo` functions.
- Added new test cases to cover scenarios with pointer fields to slices.

### Changed 1.1.0

- Updated `FilterStructTo` function to handle cases where the source field is a pointer to a slice and the destination field is a non-pointer slice.
- Modified `MergeStructUpdateTo` function to correctly merge pointer fields to slices.

### Fixed

- Fixed an issue where fields with pointer types to slices were not correctly filtered or merged.
- Resolved test failures related to pointer fields and slice type mismatches.

## [1.0.0] - 2024-05-11

[1.0.0]: https://github.com/itsatony/struccy/releases/tag/v1.0.0

### Added 1.0.0

- Initial release of the `struccy` package.
- Implemented `MergeStructUpdateTo` function to merge fields from a source struct to a destination struct.
- Implemented `FilterStructTo` function to filter fields from a source struct to a destination struct based on specified criteria.
- Implemented `CopyStructTo` function to copy fields from a source struct to a destination struct.
- Implemented `StructToMapFields` function to convert specified struct fields to a map.
- Implemented `StructToMap` function to convert all struct fields to a map.
- Added comprehensive test coverage for all functions.
- Provided detailed documentation and usage examples in the README.md file.

This changelog entry for version 1.1.0 highlights the changes and improvements made in this release. The main focus of this version is the added support for filtering and merging structs with pointer fields to slices.

The `FilterStructTo` function has been updated to handle cases where the source field is a pointer to a slice and the destination field is a non-pointer slice. Similarly, the `MergeStructUpdateTo` function has been modified to correctly merge pointer fields to slices.

New test cases have been added to cover scenarios with pointer fields to slices, ensuring the correctness of the functionality.

The error handling and type mismatch detection have been improved in both `FilterStructTo` and `MergeStructUpdateTo` functions to provide clearer error messages and handle type mismatches more effectively.

Additionally, some test failures related to pointer fields and slice type mismatches have been resolved.

The changelog also includes a reference to the previous version (1.0.0) and provides links to compare the changes between versions using the GitHub comparison URL.
