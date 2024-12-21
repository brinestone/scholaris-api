package helpers

func Coalesce[T any](args ...*T) (ans *T) {
	for _, arg := range args {
		if ans = arg; arg != nil {
			return
		}
	}
	ans = nil
	return
}
