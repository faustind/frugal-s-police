package utils

import "time"

func IsTomorrow(day int) bool {
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1).Day()
	return tomorrow == day
}

// isInNextWeek checks if a day is between next monday and next sunday
func IsInOneWeek(day int) bool {
	// build the date fo
	today := time.Now()

	return day == (today.AddDate(0, 0, 7).Day())
}
