package tz

import (
	"time"

	_ "time/tzdata" // Import tzdata so we have it available in slim environments where tzdata may not be present
)

// Go converts a time zone into a Go-compatible time zone if it is not compatible.
func Go(timeZone string) string {
	_, err := time.LoadLocation(timeZone)
	if err == nil {
		return timeZone
	}

	return windowsToIANA[timeZone]
}

// Microsoft converts a time zone into a Microsoft-compatible time zone if it is not compatible.
func Microsoft(timeZone string) string {
	_, ok := iANAToWindows[timeZone]
	if ok {
		return timeZone
	}

	return timeZone
}
