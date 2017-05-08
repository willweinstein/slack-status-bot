package main

import (
	"log"
	"net/url"
	"time"
)

type ScheduleClass struct {
	Name string `json:"name"`
	Room string `json:"room"`
	StartTime time.Time `json:"startTime"`
	Duration time.Duration `json:"duration"`
}

type ScheduleDay struct {
	Classes []ScheduleClass `json:"classes"`
}

type Schedule struct {
	Days [10]ScheduleDay `json:"days"` // 10 days = mon through thurs + fridays 1 through 4 + sat + sun
}

func Schedule_GetDayNum(t time.Time) (int) {
	weekday := t.Weekday()
	if weekday == time.Monday {
		return 0
	} else if weekday == time.Tuesday {
		return 1
	} else if weekday == time.Wednesday {
		return 2
	} else if weekday == time.Thursday {
		return 3
	} else if weekday == time.Friday {
		// check what friday it is
		// TODO: some sort of caching or something?
		response, err := API_MHS_Request("GET", "planner/fridays/get/" + t.Format(API_MHS_FormatTime), url.Values{})
		if err != nil || response["status"] == "error" || response["friday"] == "" {
			log.Println("Error getting Friday number:")
			log.Println(err)
			return 4
		}
		return 3 + int((response["friday"].(map[string]interface{})["index"]).(float64))
	} else if weekday == time.Saturday {
		return 8
	} else if weekday == time.Sunday {
		return 9
	}
	return 0
}

func Schedule_GetNormalTime(startTime time.Time, now time.Time) (time.Time) {
	return startTime.AddDate(now.Year() - startTime.Year(), int(now.Month() - startTime.Month()), now.Day() - startTime.Day())
}

func Schedule_FindNextClass(schedule Schedule, now time.Time, currentDayNumber int, classesDoneToday []ScheduleClass) (bool, ScheduleClass, bool, time.Time) {
	earliestClass := ScheduleClass{}
	earliestClassNormalStartTime := time.Now()
	foundAClass := false
	for _, class := range schedule.Days[currentDayNumber].Classes {
		// normalStartTime is needed because otherwise the class's date will be Jan 1st, 1899
		normalStartTime := Schedule_GetNormalTime(class.StartTime, now)
		endTime := normalStartTime.Add(class.Duration)

		if now.After(normalStartTime) {
			// class has started. but has it ended?
			if endTime.After(now) {
				// is it one of the classes we've already had today?
				shouldEscape := false
				for _, classToCheck := range classesDoneToday{
					if classToCheck == class {
						// it is! ignore it
						shouldEscape = true
						break
					}
				}
				if shouldEscape {
					continue
				}
				// we're in a class, break out now
				return true, class, true, endTime
			}
		} else {
			// class hasn't started yet. is its start time before the earliest one we've found so far?
			// (or is there not even an earliest one yet)
			if foundAClass == false || normalStartTime.Before(earliestClassNormalStartTime) {
				earliestClass = class
				earliestClassNormalStartTime = normalStartTime
				foundAClass = true
			}
		}
	}
	return foundAClass, earliestClass, false, time.Now()
}
