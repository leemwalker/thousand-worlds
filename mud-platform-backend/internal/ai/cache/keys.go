package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateContextHash creates a hash of the variable context factors
func GenerateContextHash(mood, desire string, affection, trust, fear int, driftLevel string) string {
	raw := fmt.Sprintf("%s:%s:%d:%d:%d:%s", mood, desire, affection, trust, fear, driftLevel)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
