package util

import (
	"crypto/md5"
	"encoding/hex"
)

func HashThese(args ...string) string {
	hash := md5.New()
	for _, arg := range args {
		hash.Write([]byte(arg))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
