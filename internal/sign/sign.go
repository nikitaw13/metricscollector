// Package sign provides HMAC-SHA256 hashing for signing metric batches
// exchanged between the agent and the server.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMAC returns the hex-encoded HMAC-SHA256 digest of body computed with the given key.
func HMAC(key string, body []byte) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write(body)
	digest := hasher.Sum(nil)
	hexDigest := hex.EncodeToString(digest)
	return hexDigest
}
