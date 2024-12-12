package util

import (
	"crypto/md5"
	"encoding/hex"

	"encore.dev/rlog"
)

func HashThese(args ...string) string {
	hash := md5.New()
	for _, arg := range args {
		rlog.Debug("hash", "results", hash.Sum([]byte(arg)))
	}
	buf := make([]byte, hash.Size())
	rlog.Debug("bytes", "bytes", string(buf))
	hash.Write(buf)
	rlog.Debug("bytes", "bytes", string(buf))

	return hex.EncodeToString(buf)
}
