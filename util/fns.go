package util

import (
	"crypto/md5"
	"encoding/hex"
)

func HashThese(args ...string) string {
	hash := md5.New()
	for _, arg := range args {
		hash.Sum([]byte(arg))
	}
	buf := make([]byte, 0)
	hash.Write(buf)

	return hex.EncodeToString(buf)
}
