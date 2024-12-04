package helpers

func MemberOf[T comparable](target T, slice ...T) bool {
	index, found := FindIndex(slice, func(a T) bool {
		return a == target
	})
	return found && index >= 0
}

func SliceOf[T any](t ...T) (ans []T) {
	ans = t[:]
	return
}

func FindIndex[T any](slice []T, pred func(a T) bool) (ans int, ok bool) {
	ans = -1
	for i, v := range slice {
		if ok = pred(v); ok {
			ans = i
			break
		}
	}
	return
}

func Find[T any](slice []T, pred func(a T) bool) (ans T, ok bool) {
	for _, v := range slice {
		if ok = pred(v); ok {
			ans = v
			break
		}
	}
	return
}

func SliceMap[T any, R any](slice []T, mapper func(a T) R) (ans []R) {
	ans = make([]R, len(slice))
	for i, v := range slice {
		ans[i] = mapper(v)
	}
	return
}
