package wire

import "io"

// DecoderNG reads Nix wire format values from an input stream.
type DecoderNG struct {
	r        io.Reader
	maxBytes uint64
}

// NewDecoderNG returns a DecoderNG that reads from r. The maxBytes parameter
// limits the maximum size of string and byte packets to prevent excessive
// memory allocation.
func NewDecoderNG(r io.Reader, maxBytes uint64) *DecoderNG {
	return &DecoderNG{r: r, maxBytes: maxBytes}
}

// ReadUint64 reads a uint64 in Nix wire format.
func (d *DecoderNG) Uint64(p *uint64) (err error) {
	*p, err = ReadUint64(d.r)
	return err
}

// ReadBool reads a boolean in Nix wire format.
func (d *DecoderNG) Bool(p *bool) (err error) {
	*p, err = ReadBool(d.r)
	return err
}

// ReadBytes reads a byte packet and returns its contents.
func (d *DecoderNG) Bytes(p *[]byte) (err error) {
	*p, err = ReadBytesFull(d.r, d.maxBytes)
	return err
}

// ReadString reads a string packet in Nix wire format.
func (d *DecoderNG) String(p *string) (err error) {
	*p, err = ReadString(d.r, d.maxBytes)
	return err
}

// ReadStrings reads a list of strings (count + entries).
func (d *DecoderNG) Strings(p *[]string) (err error) {
	*p, err = ReadStrings(d.r, d.maxBytes)
	return err
}

// ReadStringMap reads a map of string key/value pairs (count + entries).
func (d *DecoderNG) StringMap(p *map[string]string) (err error) {
	*p, err = ReadStringMap(d.r, d.maxBytes)
	return err
}

// Decode decodes a value that implements Unmarshaler.
// func (d *DecoderNG) Decode(v Unmarshaler) error {
// 	return v.UnmarshalNix(d)
// }
