// Package vfsutil implements some I/O utility functions for webdav.FileSystem.
package vfsutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/net/webdav"
)

// Create creates the named file with mode 0644 (before umask), truncating
// it if it already exists. If successful, methods on the returned
// File can be used for I/O; the associated file descriptor has mode O_RDWR.
// If there is an error, it will be of type *os.PathError.
func Create(ctx context.Context, fs webdav.FileSystem, name string) (webdav.File, error) {
	return fs.OpenFile(ctx, name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

// Open opens the named file for reading.  If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *os.PathError.
func Open(ctx context.Context, fs webdav.FileSystem, name string) (http.File, error) {
	return fs.OpenFile(ctx, name, os.O_RDONLY, 0)
}

// ReadDir reads the contents of the directory associated with file and
// returns a slice of FileInfo values in directory order.
func ReadDir(ctx context.Context, fs webdav.FileSystem, name string) ([]os.FileInfo, error) {
	f, err := fs.OpenFile(ctx, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(0)
}

// Stat returns the FileInfo structure describing file.
func Stat(ctx context.Context, fs webdav.FileSystem, name string) (os.FileInfo, error) {
	f, err := fs.OpenFile(ctx, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}

// WriteFile writes data to a file named by name.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func WriteFile(ctx context.Context, fs webdav.FileSystem, name string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(ctx, name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error. The permission bits perm are used
// for all directories that MkdirAll creates. If path is already a directory,
// MkdirAll does nothing and returns nil.
func MkdirAll(ctx context.Context, fs webdav.FileSystem, path string, perm os.FileMode) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := fs.Stat(ctx, path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: fmt.Errorf("not a directory")}
	}

	// Slow path: make sure parent exists and then call Mkdir for path.
	i := len(path)
	for i > 0 && path[i-1] == '/' { // Skip trailing path separator.
		i--
	}

	j := i
	for j > 0 && path[j-1] != '/' { // Scan backward over element.
		j--
	}

	if j > 1 {
		// Create parent.
		err = MkdirAll(ctx, fs, path[0:j-1], perm)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = fs.Mkdir(ctx, path, perm)
	if err != nil {
		// Handle arguments like "foo/." by double-checking that directory doesn't exist.
		dir, err1 := fs.Stat(ctx, path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}
