package engine

var faceByMood = map[string]string{
	MoodHappy:    "(^_^)",
	MoodNeutral:  "(-_-)",
	MoodSick:     "(x_x)",
	MoodDirty:    "(>_<)",
	MoodSleeping: "(-,-) zzz",
}

const faceFallback = "(?_?)"

func (p PetState) Face() string {
	if face, ok := faceByMood[p.Mood]; ok {
		return face
	}
	return faceFallback
}
