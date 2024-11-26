package helpers

func Find[T any](slice []T, pred func(a T) bool) (ans int, ok bool) {
	ans = -1
	for i, v := range slice {
		if ok = pred(v); ok {
			ans = i
		}
	}
	return
}

func Map[T any, R any](slice []T, mapper func(a T) R) (ans []R) {
	ans = make([]R, len(slice))
	for i, v := range slice {
		ans[i] = mapper(v)
	}
	return
}
