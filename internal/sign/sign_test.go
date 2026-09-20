package sign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMAC(t *testing.T) {
	t.Parallel()

	t.Run("test vector", func(t *testing.T) {
		t.Parallel()
		got := HMAC("Jefe", []byte("what do ya want for nothing?"))
		assert.Equal(t, "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843", got)
	})

	t.Run("determinism", func(t *testing.T) {
		t.Parallel()
		body := []byte(`{"id":"test","type":"gauge","value":1}`)
		assert.Equal(t, HMAC("TestKey", body), HMAC("TestKey", body))
	})

	t.Run("sensitivity to key and body", func(t *testing.T) {
		t.Parallel()
		body := []byte(`{"id":"test","type":"gauge","value":1}`)
		base := HMAC("TestKey", body)
		assert.NotEqual(t, base, HMAC("OtherKey", body), "digest must change with the key")
		assert.NotEqual(t, base, HMAC("TestKey", []byte(`{"id":"test","type":"gauge","value":2}`)),
			"digest must change with the body")
	})

	t.Run("empty key and body", func(t *testing.T) {
		t.Parallel()
		assert.Regexp(t, `^[0-9a-f]{64}$`, HMAC("", nil), "empty key and body must yield a valid hex digest")
		assert.Regexp(t, `^[0-9a-f]{64}$`, HMAC("TestKey", nil), "empty body must yield a valid hex digest")
	})
}
