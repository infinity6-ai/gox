/*
Package iobytez provides functionality for reading bytes from an io.Reader with more control over the reading process.

The Read function allows specifying minimum and maximum bytes to read, and a timeout. This is useful for scenarios where you need to read a certain amount of data but also want to avoid blocking indefinitely.
*/
package iobytez
