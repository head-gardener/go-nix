package wire

import "io"

// EncoderNG writes Nix wire format values to an output stream.
type EncoderNG struct{ w io.Writer }

// NewEncoderNG returns an EncoderNG that writes to w.
func NewEncoderNG(w io.Writer) *EncoderNG {
	return &EncoderNG{w: w}
}

// Uint64 writes a uint64 in Nix wire format.
func (e *EncoderNG) Uint64(n *uint64) error {
	return WriteUint64(e.w, *n)
}

// Bool writes a boolean in Nix wire format.
func (e *EncoderNG) Bool(v *bool) error {
	return WriteBool(e.w, *v)
}

// Bytes writes a byte packet in Nix wire format.
func (e *EncoderNG) Bytes(buf *[]byte) error {
	return WriteBytes(e.w, *buf)
}

// String writes a string packet in Nix wire format.
func (e *EncoderNG) String(s *string) error {
	return WriteString(e.w, *s)
}

// Strings writes a list of strings as count + entries.
func (e *EncoderNG) Strings(ss *[]string) error {
	return WriteStrings(e.w, *ss)
}

// StringMap writes a map as count + sorted key/value pairs.
func (e *EncoderNG) StringMap(m *map[string]string) error {
	return WriteStringMap(e.w, *m)
}

// Writer returns the underlying writer.
func (e *EncoderNG) Writer() io.Writer {
	return e.w
}

// Encode encodes a value that implements Marshaler.
// func (e *EncoderNG) Encode(v Marshaler) error {
// 	return v.MarshalNix(e)
// }
