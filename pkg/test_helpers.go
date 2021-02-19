package main

import (
    "encoding/json"
    "time"
    "testing"
)

func MultiReturnHelperParseDuration(result time.Duration, err error) time.Duration {
    return result
}

func MultiReturnHelperParse(result time.Time, err error) time.Time {
    return result
}

func InitString(value string) *string{
    new_string := value
    return &new_string
}

func InitRawMsg(value string) *json.RawMessage{
    new_msg := json.RawMessage(value)
    return &new_msg
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

func SingleDataCompareHelper(result []SingleData, wanted []SingleData, t *testing.T){
    // Raise no errors if two []SingleData are identical, raise errors if they are not
    if len(result) != len(wanted) {
        t.Errorf("Input and output SingleData differ in length. Wanted %v, got %v", len(wanted), len(result))
        return
    }
    for udx, _ := range wanted {
        if result[udx].Name != wanted[udx].Name {
            t.Errorf("Input and output SingleData have different Pvs. Wanted %v, got %v", wanted[udx].Name, result[udx].Name)
        }
        if len(wanted[udx].Values) != len(result[udx].Values) {
            t.Errorf("Input and output arrays differ in length. Wanted %v, got %v", len(wanted[udx].Values), len(result[udx].Values))
            return
        }
        for idx, _ := range(wanted[udx].Values) {
            if result[udx].Values[idx] != wanted[udx].Values[idx] {
                t.Errorf("Values at index %v do not match, Wanted %v, got %v", idx, wanted[udx].Values[idx], result[udx].Values[idx])
            }
        }
    }
}
