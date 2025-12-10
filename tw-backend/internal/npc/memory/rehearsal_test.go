package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateRehearsalBonus(t *testing.T) {
	assert.Equal(t, 0.0, CalculateRehearsalBonus(0))
	assert.Equal(t, 0.25, CalculateRehearsalBonus(5))
	assert.Equal(t, 0.5, CalculateRehearsalBonus(10))
	assert.Equal(t, 0.5, CalculateRehearsalBonus(20))
	assert.Equal(t, 0.5, CalculateRehearsalBonus(100))
}

func TestRecordAccess(t *testing.T) {
	now := time.Now()
	mem := Memory{AccessCount: 0}

	RecordAccess(&mem, now)

	assert.Equal(t, 1, mem.AccessCount)
	assert.Equal(t, now, mem.LastAccessed)
}
