package stream

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"strings"
)

func HashKey(s string) []byte {
	k := sha256.Sum256([]byte(s))
	return k[:]
}

// Hash should be used to hash file names.
func HashName(s string) string {
	k := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(k[:])
}

// HashPath expects a HashPath with trailing path separator.
// It returns the hashed file names separated by a forward slash.
func HashPath(s string) string {
	const pathSep = string(os.PathSeparator)

	names := strings.Split(s, pathSep)
	for i, name := range names[1:] {
		names[1+i] = HashName(name)
	}
	return strings.Join(names, "/")
}
