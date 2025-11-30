package testutil

import (
	"github.com/google/uuid"
)

// GenerateTestEmail generates a unique test email
func GenerateTestEmail() string {
	return "test-" + uuid.New().String() + "@example.com"
}

// GenerateTestName generates a unique test name
func GenerateTestName(prefix string) string {
	return prefix + "-" + uuid.New().String()[:8]
}
