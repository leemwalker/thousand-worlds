package damage

import "testing"

func TestCalculateCritical(t *testing.T) {
	// Base threshold 95
	res := CalculateCritical(95, 0, false)
	if !res.IsCritical {
		t.Error("Expected critical at 95 with 0 cunning")
	}

	// Cunning 50 -> +1% -> Threshold 94
	res = CalculateCritical(94, 50, false)
	if !res.IsCritical {
		t.Error("Expected critical at 94 with 50 cunning")
	}

	// Heavy Attack -> +5% -> Threshold 90
	res = CalculateCritical(90, 0, true)
	if !res.IsCritical {
		t.Error("Expected critical at 90 with heavy attack")
	}

	// Fumble
	res = CalculateCritical(5, 0, false)
	if !res.IsCriticalFailure {
		t.Error("Expected fumble at 5")
	}
}

func TestReduceDurability(t *testing.T) {
	w := &Weapon{Durability: 100, MaxDurability: 100}

	// Normal hit -1
	res := ReduceDurability(w, 1)
	if w.Durability != 99 {
		t.Errorf("Expected 99, got %d", w.Durability)
	}
	if res.DamageModifier != 1.0 {
		t.Error("Expected 1.0 modifier")
	}

	// Drop below 50%
	w.Durability = 49
	res = GetDurabilityStatus(w)
	if res.DamageModifier != 0.9 {
		t.Errorf("Expected 0.9 modifier at 49 durability, got %v", res.DamageModifier)
	}

	// Broken
	w.Durability = 0
	res = GetDurabilityStatus(w)
	if !res.IsBroken {
		t.Error("Expected broken status")
	}
	if res.DamageModifier != 0.0 {
		t.Error("Expected 0.0 modifier for broken weapon")
	}
}
