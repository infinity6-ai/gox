// Package filez provides a collection of utility functions for file and directory manipulation.
// These functions are designed to simplify common file system operations by providing robust,
// error-checked, and easy-to-use interfaces. The package includes functionalities for path
// manipulation, checking file existence and properties, creating and deleting files and directories,
// writing to and reading from files, and listing directory contents.
//
// The utilities in this package are built on top of the standard Go `os` and `path/filepath`
// packages, but with an emphasis on convenience and safety. For instance, many functions
// automatically create parent directories when writing files, and error handling is simplified
// through the use of `errorz.Check` to panic on unexpected errors, aligning with the project's
// error handling philosophy.
package filez
