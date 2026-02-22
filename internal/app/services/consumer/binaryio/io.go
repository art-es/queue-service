package binary

type IO struct {
	*reader
	*writer
}

func New() *IO {
	return &IO{
		reader: newReader(),
		writer: newWriter(),
	}
}
