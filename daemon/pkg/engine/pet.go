package engine

import (
	"fmt"

	"tamagops/daemon/pkg/collector"
)

type PetState struct {
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	CurrentXP   int      `json:"current_xp"`
	NextLevelXP int      `json:"next_level_xp"`
	HP          int      `json:"hp"`
	MaxHP       int      `json:"max_hp"`
	Mood        string   `json:"mood"`
	Debuffs     []string `json:"active_debuffs"`

	ConsecutiveHighCPU int `json:"-"`
	ConsecutiveIdleCPU int `json:"-"`
}

const (
	MoodHappy    = "HAPPY"
	MoodNeutral  = "NEUTRAL"
	MoodSick     = "SICK"
	MoodDirty    = "DIRTY"
	MoodSleeping = "SLEEPING"
)

const (
	DebuffRAMOverload   = "RAM_OVERLOAD"
	DebuffDiskOverload  = "DISK_OVERLOAD"
	DebuffZombieProcess = "ZOMBIE_PROCESS_DETECTED"
	DebuffOverheating   = "OVERHEATING"
)

const (
	cpuIdealMin = 20.0
	cpuIdealMax = 60.0
	cpuIdleMax  = 5.0
	cpuFeverMin = 90.0

	ramDirtyThreshold  = 85.0
	diskDirtyThreshold = 90.0

	cpuTempOverheatThreshold = 85.0

	feverStreakRequired = 5
	sleepStreakRequired = 10

	xpPerTickInIdealRange = 5
	hpLossPerFeverTick    = 1
)

func NewPet(name string) PetState {
	return PetState{
		Name:        name,
		Level:       1,
		CurrentXP:   0,
		NextLevelXP: 100,
		HP:          100,
		MaxHP:       100,
		Mood:        MoodNeutral,
		Debuffs:     []string{},
	}
}

func CalculatePetState(snap collector.SystemSnapshot, prev PetState) PetState {
	next := prev
	next.Debuffs = []string{}

	applyCPURules(snap, &next)
	applyHygieneRules(snap, &next)
	applyProcessRules(snap, &next)
	applyTemperatureRules(snap, &next)
	applyLevelUp(&next)
	clampHP(&next)
	resolveMoodPriority(&next)

	return next
}

func applyCPURules(snap collector.SystemSnapshot, next *PetState) {
	switch {
	case snap.CPUPercent >= cpuIdealMin && snap.CPUPercent <= cpuIdealMax:
		next.CurrentXP += xpPerTickInIdealRange
		next.ConsecutiveHighCPU = 0
		next.ConsecutiveIdleCPU = 0

	case snap.CPUPercent < cpuIdleMax:
		next.ConsecutiveHighCPU = 0
		next.ConsecutiveIdleCPU++

	case snap.CPUPercent > cpuFeverMin:
		next.ConsecutiveIdleCPU = 0
		next.ConsecutiveHighCPU++

	default:
		next.ConsecutiveHighCPU = 0
		next.ConsecutiveIdleCPU = 0
	}

	if next.ConsecutiveHighCPU >= feverStreakRequired {
		next.HP -= hpLossPerFeverTick
	}
}

func applyHygieneRules(snap collector.SystemSnapshot, next *PetState) {
	if snap.RAMUsedPercent > ramDirtyThreshold {
		next.Debuffs = append(next.Debuffs, DebuffRAMOverload)
	}
	if snap.DiskUsedPercent > diskDirtyThreshold {
		next.Debuffs = append(next.Debuffs, DebuffDiskOverload)
	}
}

func applyProcessRules(snap collector.SystemSnapshot, next *PetState) {
	if snap.TopProcess.CPUPerc > 80 {
		next.Debuffs = append(next.Debuffs, DebuffZombieProcess)
	}
}

func applyTemperatureRules(snap collector.SystemSnapshot, next *PetState) {
	if !snap.SensorsAvailable {
		return
	}

	if snap.CPUTemp >= cpuTempOverheatThreshold {
		next.Debuffs = append(next.Debuffs, DebuffOverheating)
		next.HP -= hpLossPerFeverTick
	}
}

func applyLevelUp(next *PetState) {
	for next.CurrentXP >= next.NextLevelXP {
		next.CurrentXP -= next.NextLevelXP
		next.Level++
		next.NextLevelXP = int(float64(next.NextLevelXP) * 1.5)
	}
}

func clampHP(next *PetState) {
	if next.HP < 0 {
		next.HP = 0
	}
	if next.HP > next.MaxHP {
		next.HP = next.MaxHP
	}
}

func resolveMoodPriority(next *PetState) {
	switch {
	case next.HP <= 0:
		next.Mood = MoodSick

	case next.ConsecutiveHighCPU >= feverStreakRequired:
		next.Mood = MoodSick

	case hasDebuff(next.Debuffs, DebuffOverheating):
		next.Mood = MoodSick

	case hasDebuff(next.Debuffs, DebuffRAMOverload) || hasDebuff(next.Debuffs, DebuffDiskOverload):
		next.Mood = MoodDirty

	case next.ConsecutiveIdleCPU >= sleepStreakRequired:
		next.Mood = MoodSleeping

	default:
		next.Mood = MoodHappy
	}
}

func hasDebuff(debuffs []string, target string) bool {
	for _, d := range debuffs {
		if d == target {
			return true
		}
	}
	return false
}

func BuildLogMessage(snap collector.SystemSnapshot, state PetState) string {
	switch state.Mood {
	case MoodSick:
		if hasDebuff(state.Debuffs, DebuffOverheating) {
			return fmt.Sprintf(
				"Estou pegando fogo! CPU a %.1f°C — por favor, verifique a refrigeração!",
				snap.CPUTemp,
			)
		}
		return fmt.Sprintf(
			"Febre alta detectada! O processo %s (PID %d) está me dando uma baita indigestão!",
			state.criticalProcessNameOrFallback(snap), snap.TopProcess.PID,
		)
	case MoodDirty:
		return "Estou ficando sujo... alguém precisa limpar essa bagunça de RAM/disco!"
	case MoodSleeping:
		return "Zzz... está tudo tão calmo que acabei cochilando."
	default:
		return "Tudo tranquilo por aqui, sistema saudável!"
	}
}

func (p PetState) criticalProcessNameOrFallback(snap collector.SystemSnapshot) string {
	if snap.TopProcess.Name == "" {
		return "processo desconhecido"
	}
	return snap.TopProcess.Name
}
