package testhelper

import (
	"encoding/json"
	"time"
)

func MultiReturnHelperParseDuration(result time.Duration, err error) time.Duration {
	return result
}

func MultiReturnHelperParse(result time.Time, err error) time.Time {
	return result
}

func InitString(value string) *string {
	new_string := value
	return &new_string
}

func InitRawMsg(value string) *json.RawMessage {
	new_msg := json.RawMessage(value)
	return &new_msg
}

func InitIntPointer(value int) *int {
	new_int := value
	return &new_int
}

func TimeHelper(minutes int) time.Time {
	// Shortcut for generating consistent timestamps using only a single int
	return time.Date(2021, time.January, 10, 1, minutes, 0, 0, time.UTC)
}

func TimeArrayHelper(start int, end int) []time.Time {
	// Shortcut for generating a slice of consistent timestamps using the [beginning, ending) ints
	counter := start
	var increment int
	if start < end {
		increment = 1
	} else {
		increment = -1
	}
	var response []time.Time
	for counter != end {
		counter = counter + increment
		response = append(response, TimeHelper(counter))
	}
	return response
}
