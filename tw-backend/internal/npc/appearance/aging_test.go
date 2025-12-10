package appearance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAgeCategory(t *testing.T) {
	lifespan := 100

	assert.Equal(t, AgeChild, GetAgeCategory(10, lifespan))
	assert.Equal(t, AgeYoungAdult, GetAgeCategory(30, lifespan))
	assert.Equal(t, AgeAdult, GetAgeCategory(50, lifespan))
	assert.Equal(t, AgeMiddleAged, GetAgeCategory(70, lifespan))
	assert.Equal(t, AgeElder, GetAgeCategory(90, lifespan))
}

func TestApplyAgeModifiers(t *testing.T) {
	base := "A tall human"

	desc := ApplyAgeModifiers(base, AgeElder)
	assert.Contains(t, desc, "elderly")
	assert.Contains(t, desc, "stooped posture")

	desc = ApplyAgeModifiers(base, AgeChild)
	assert.Contains(t, desc, "youthful")
}
