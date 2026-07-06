package idempotency

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
)

func Hash(parts ...string) string {
	hasher := sha256.New()

	var length [8]byte

	for _, part := range parts {
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))

		_, _ = hasher.Write(length[:])
		_, _ = io.WriteString(hasher, part)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
