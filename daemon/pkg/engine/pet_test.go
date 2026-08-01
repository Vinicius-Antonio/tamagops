package engine

import (
	"testing"

	"tamagops/daemon/pkg/collector"
)

func TestIdealCPUGrantsXP(t *testing.T) {
	prev := NewPet("Linus")
	snap := collector.SystemSnapshot{CPUPercent: 40}

	next := CalculatePetState(snap, prev)

	if next.CurrentXP != prev.CurrentXP+xpPerTickInIdealRange {
		t.Errorf("expected XP %d, got %d", prev.CurrentXP+xpPerTickInIdealRange, next.CurrentXP)
	}
	if next.Mood != MoodHappy {
		t.Errorf("expected mood HAPPY, got %s", next.Mood)
	}
}

func TestFeverOnlyAppearsAfterConsecutiveStreak(t *testing.T) {
	state := NewPet("Linus")
	snap := collector.SystemSnapshot{CPUPercent: 95}

	for i := 0; i < feverStreakRequired-1; i++ {
		state = CalculatePetState(snap, state)
	}
	if state.Mood == MoodSick {
		t.Fatalf("pet got sick before reaching the required streak (tick %d)", feverStreakRequired-1)
	}

	state = CalculatePetState(snap, state)
	if state.Mood != MoodSick {
		t.Errorf("expected mood SICK after %d ticks of high CPU, got %s", feverStreakRequired, state.Mood)
	}
	if state.HP != 99 {
		t.Errorf("expected HP 99 after 1 fever tick, got %d", state.HP)
	}
}

func TestHighRAMTriggersDirtyDebuff(t *testing.T) {
	prev := NewPet("Linus")
	snap := collector.SystemSnapshot{CPUPercent: 40, RAMUsedPercent: 90}

	next := CalculatePetState(snap, prev)

	if !hasDebuff(next.Debuffs, DebuffRAMOverload) {
		t.Errorf("expected debuff %s, got %v", DebuffRAMOverload, next.Debuffs)
	}
	if next.Mood != MoodDirty {
		t.Errorf("expected mood DIRTY, got %s", next.Mood)
	}
}

func TestProlongedIdleCPUPutsPetToSleep(t *testing.T) {
	state := NewPet("Linus")
	snap := collector.SystemSnapshot{CPUPercent: 1}

	for i := 0; i < sleepStreakRequired; i++ {
		state = CalculatePetState(snap, state)
	}

	if state.Mood != MoodSleeping {
		t.Errorf("expected mood SLEEPING, got %s", state.Mood)
	}
}

func TestHPNeverExceedsMaxHP(t *testing.T) {
	prev := NewPet("Linus")
	prev.HP = prev.MaxHP
	snap := collector.SystemSnapshot{CPUPercent: 40}

	next := CalculatePetState(snap, prev)

	if next.HP > next.MaxHP {
		t.Errorf("HP (%d) exceeded MaxHP (%d)", next.HP, next.MaxHP)
	}
}

func TestLevelUpResetsExcessXPAndRaisesNextGoal(t *testing.T) {
	prev := NewPet("Linus")
	prev.CurrentXP = prev.NextLevelXP - 1
	snap := collector.SystemSnapshot{CPUPercent: 40}

	next := CalculatePetState(snap, prev)

	if next.Level != prev.Level+1 {
		t.Errorf("expected level %d, got %d", prev.Level+1, next.Level)
	}
	if next.NextLevelXP <= prev.NextLevelXP {
		t.Errorf("NextLevelXP should have increased, stayed at %d", next.NextLevelXP)
	}
}