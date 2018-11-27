package log

import (
	"go.uber.org/zap"
)

var (

	// Binary constructs a field that carries an opaque binary blob.
	//
	// Binary data is serialized in an encoding-appropriate format. For example,
	// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
	// use ByteString.
	Binary = zap.Binary

	// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
	// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
	// Binary.
	ByteString = zap.ByteString

	// String adds a string-valued key:value pair to a Span.LogFields() record
	String = zap.String

	// Bool adds a bool-valued key:value pair to a Span.LogFields() record
	Bool = zap.Bool

	// Int adds an int-valued key:value pair to a Span.LogFields() record
	Int = zap.Int

	// Int32 adds an int32-valued key:value pair to a Span.LogFields() record
	Int32 = zap.Int32

	// Int64 adds an int64-valued key:value pair to a Span.LogFields() record
	Int64 = zap.Int64

	// Uint32 adds a uint32-valued key:value pair to a Span.LogFields() record
	Uint32 = zap.Uint32

	// Uint64 adds a uint64-valued key:value pair to a Span.LogFields() record
	Uint64 = zap.Uint64

	// Float32 adds a float32-valued key:value pair to a Span.LogFields() record
	Float32 = zap.Float32

	// Float64 adds a float64-valued key:value pair to a Span.LogFields() record
	Float64 = zap.Float64

	// Error adds an error with the key "error" to a Span.LogFields() record
	Error = zap.Error

	// Object adds an object-valued key:value pair to a Span.LogFields() record
	Object = zap.Object

	// Noop creates a no-op log field that should be ignored by the tracer.
	Noop = zap.Skip

	// Any adds an any-valued key:value pair to a Span.LogFields() record
	Any = zap.Any
)
