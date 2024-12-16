package helpers

type Pointer interface{}

func Coalesce(args ...Pointer) (ans Pointer) {
	for _, arg := range args {
		if ans = arg; arg != nil {
			return
		}
	}
	ans = nil
	return
}
