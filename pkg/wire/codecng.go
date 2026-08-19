package wire

type Codec interface {
	Uint64(p *uint64) error
	Bool(p *bool) error
	Bytes(p *[]byte) error
	String(p *string) error
	Strings(p *[]string) error
	StringMap(p *map[string]string) error
}
