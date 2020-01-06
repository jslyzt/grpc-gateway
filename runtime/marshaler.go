package runtime

import (
	"io"
)

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)      // Marshal marshals "v" into byte sequence
	Unmarshal(data []byte, v interface{}) error // Unmarshal unmarshals "data" into "v". "v" must be a pointer value
	NewDecoder(r io.Reader) Decoder             // NewDecoder returns a Decoder which reads byte sequence from "r"
	NewEncoder(w io.Writer) Encoder             // NewEncoder returns an Encoder which writes bytes sequence into "w"
	ContentType() string                        // ContentType returns the Content-Type which this marshaler is responsible for
}

// Decoder decodes a byte sequence
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder encodes gRPC payloads / fields into byte sequence.
type Encoder interface {
	Encode(v interface{}) error
}

// DecoderFunc adapts an decoder function into Decoder.
type DecoderFunc func(v interface{}) error

// Decode delegates invocations to the underlying function itself.
func (f DecoderFunc) Decode(v interface{}) error {
	return f(v)
}

// EncoderFunc adapts an encoder function into Encoder
type EncoderFunc func(v interface{}) error

// Encode delegates invocations to the underlying function itself.
func (f EncoderFunc) Encode(v interface{}) error {
	return f(v)
}

// Delimited defines the streaming delimiter.
type Delimited interface {
	Delimiter() []byte // Delimiter returns the record seperator for the stream
}
