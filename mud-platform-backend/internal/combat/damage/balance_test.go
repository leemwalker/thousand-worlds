package damage

import (
	"math/rand"
	"mud-platform-backend/internal/character"
	"testing"
	"time"
)

func TestDamageBalance(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	attrs := character.Attributes{Might: 50, Agility: 50, Cunning: 50}

	shortSword := &Weapon{Name: "Short Sword", Type: WeaponSlashing, BaseDamage: 15, Durability: 100, MaxDurability: 100}
	longSword := &Weapon{Name: "Longsword", Type: WeaponSlashing, BaseDamage: 22, Durability: 100, MaxDurability: 100}
	greatSword := &Weapon{Name: "Greatsword", Type: WeaponSlashing, BaseDamage: 35, Durability: 100, MaxDurability: 100}

	runSim := func(w *Weapon, skill int) float64 {
		total := 0
		for i := 0; i < 1000; i++ {
			roll := rand.Intn(100) + 1
			res := CalculateDamage(attrs, w, skill, nil, roll, false)
			total += res.FinalDamage
		}
		return float64(total) / 1000.0
	}

	avgShort := runSim(shortSword, 50)
	avgLong := runSim(longSword, 50)
	avgGreat := runSim(greatSword, 50)

	t.Logf("Averages: Short=%.2f, Long=%.2f, Great=%.2f", avgShort, avgLong, avgGreat)

	if avgGreat <= avgLong || avgLong <= avgShort {
		t.Error("Weapon hierarchy violated: Expected Great > Long > Short")
	}

	// Skill Scaling
	avgSkill0 := runSim(longSword, 0)
	avgSkill100 := runSim(longSword, 100)

	t.Logf("Skill Scaling: 0=%.2f, 100=%.2f", avgSkill0, avgSkill100)

	// Skill 100 should be roughly 1.5x Skill 0 (since skill mod goes 1.0 -> 1.5)
	ratio := avgSkill100 / avgSkill0
	if ratio < 1.4 || ratio > 1.6 {
		t.Errorf("Skill scaling off: Expected ~1.5x, got %.2fx", ratio)
	}
}
