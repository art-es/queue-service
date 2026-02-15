package ops

func Pointer[T any](v T) *T {
	return &v
}

func PointerOrNil[T comparable](v T) *T {
	var dv T
	if v == dv {
		return nil
	}
	return &v
}

func Value[T any](p *T) (v any) {
	if p != nil {
		v = *p
	}
	return
}

func ErrorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
