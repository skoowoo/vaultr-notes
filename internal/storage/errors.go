package storage

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is the base sentinel returned when a path does not exist in
	// the vault. Callers should prefer the more specific ErrMetadataNotFound
	// or ErrFileNotFound when the origin matters; all three satisfy
	// errors.Is(err, ErrNotFound).
	ErrNotFound = errors.New("not found")

	// ErrMetadataNotFound is returned when the metadata DB has no record for the
	// requested path (i.e. the note was never registered or has been removed from
	// the index). It wraps ErrNotFound for backward-compatible Is checks.
	ErrMetadataNotFound = fmt.Errorf("metadata record not found: %w", ErrNotFound)

	// ErrFileNotFound is returned when the note exists in the metadata DB but its
	// backing file is absent on disk (stale index row). It wraps ErrNotFound for
	// backward-compatible Is checks.
	ErrFileNotFound = fmt.Errorf("file not found on disk: %w", ErrNotFound)

	// ErrIsDir is returned when a file operation is attempted on a directory.
	ErrIsDir = errors.New("path is a directory")

	// ErrIsFile is returned when a directory operation is attempted on a file.
	ErrIsFile = errors.New("path is a file")

	// ErrInvalidPath is returned when a path escapes the vault root or is otherwise invalid.
	ErrInvalidPath = errors.New("invalid path")

	// ErrUnsupportedType is returned when a file path does not have a markdown extension.
	ErrUnsupportedType = errors.New("only markdown files (.md, .markdown) are supported")

	// ErrBinaryContent is returned when the content to be written is not valid UTF-8 text.
	ErrBinaryContent = errors.New("content must be valid UTF-8 text, not binary data")

)
