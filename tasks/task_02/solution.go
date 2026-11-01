package main

func rotateRunes(s string, shift int) string {

	if len(s) == 0 {
		return s
	}

	runeString := []rune(s)

	n := len(runeString)

	reminderShift := shift % n

	newSlice := make([]rune, 0, n)

	if reminderShift < 0 {
		 newSlice =  append(newSlice, runeString[n+reminderShift:n]...)
		 newSlice = append(newSlice, runeString[:n+reminderShift]...)
	} else {
		newSlice =  append(newSlice, runeString[reminderShift:n]...)
		newSlice = append(newSlice, runeString[:reminderShift]...)
	}

	return string(newSlice)
}
