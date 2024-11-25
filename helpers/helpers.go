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
