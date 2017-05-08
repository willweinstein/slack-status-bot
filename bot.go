package main

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"time"
)

func Bot_CanActivate() (bool) {
	if !API_MHS_Connected || !API_Slack_Connected || Storage_Get("dalton-schedule") == "" {
		return false
	}

	return true
}

// HACK: this gets new york time but with the location set to UTC
// HACK: this is *very ugly* but it works
func Bot_GetNYTimeInUTC() (time.Time) {
	hackyFormatString := "2006-01-02 15:04:05.999999999"
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	todayNY := time.Now().In(loc)
	todayUTC, _ := time.Parse(hackyFormatString, todayNY.Format(hackyFormatString))
	return todayUTC
}

func Bot_FindClassByName(data map[string]interface{}, name string) (int) {
	classes := data["classes"].([]interface{})
	for _, class := range classes {
		classInfo := class.(map[string]interface{})
		if classInfo["name"].(string) == name {
			return int(classInfo["id"].(float64))
		}
	}
	return -1
}

func Bot_SetStatusToClass(class ScheduleClass) {
	log.Println("Setting status for " + class.Name)
	API_Slack_UpdateStatus(API_Slack_StatusInfo{
		class.Name + " in " + class.Room,
		":school:",
	})
}

func Bot_ResetStatus() {
	API_Slack_UpdateStatus(API_Slack_StatusInfo{
		"",
		"",
	})
}

func Bot_GetMonday(now time.Time) (string) {
	for {
		if now.Weekday() != time.Monday {
			now = now.AddDate(0, 0, -1)
		} else {
			return now.Format(API_MHS_FormatTime)
		}
	}
}

func Bot_AppendMyHomeworkSpaceEvents(schedule *Schedule, now time.Time, classId int) {
	response, err := API_MHS_Request("GET", "homework/getWeek/" + Bot_GetMonday(now), url.Values{})
	if err != nil {
		panic(err)
	}
	homework := response["homework"].([]interface{})
	targetDue := now.Format(API_MHS_FormatTime)
	targetDayNum := Schedule_GetDayNum(now)
	for _, hwItem := range homework {
		hw := hwItem.(map[string]interface{})
		if hw["due"].(string) == targetDue {
			if int(hw["classId"].(float64)) == classId {
				prefix := strings.Split(hw["name"].(string), " ")
				if prefix[0] == "BuildSession" {
					descLines := strings.Split(hw["desc"].(string), "\n")
					name := ""
					room := ""
					startTimeStr := ""
					endTimeStr := ""

					for _, line := range descLines {
						lineParts := strings.Split(line, ": ")
						if len(lineParts) < 2 {
							continue
						}
						if lineParts[0] == "Name" {
							name = lineParts[1]
						}
						if lineParts[0] == "Room" {
							room = lineParts[1]
						}
						if lineParts[0] == "Start" {
							startTimeStr = lineParts[1]
						}
						if lineParts[0] == "End" {
							endTimeStr = lineParts[1]
						}
					}

					if startTimeStr == "" || endTimeStr == "" {
						log.Println("Warning: failed to parse build session with the following description")
						log.Println(hw["desc"].(string))
						continue
					}

					startTime, err := time.Parse("3:04pm", startTimeStr)
					endTime, err2 := time.Parse("3:04pm", endTimeStr)

					if err != nil || err2 != nil {
						log.Println("Warning: failed to parse build session with the following description")
						log.Println(hw["desc"].(string))
						continue
					}

					if name == "" {
						name = "Build session"
					}
					if room == "" {
						room = "501"
					}

					duration := endTime.Sub(startTime)
					class := ScheduleClass{
						name,
						room,
						startTime.AddDate(1899 - startTime.Year(), int(time.January - startTime.Month()), 1 - startTime.Day()),
						duration,
					}

					(*schedule).Days[targetDayNum].Classes = append((*schedule).Days[targetDayNum].Classes, class)
				}
			}
		}
	}
}

func Bot_Start() {
	scheduleString := Storage_Get("dalton-schedule")
	schedule := Schedule{}

	json.Unmarshal([]byte(scheduleString), &schedule)

	response, err := API_MHS_Request("GET", "classes/get", url.Values{})
	if err != nil {
		panic(err)
	}
	eventClassID := Bot_FindClassByName(response, "Other")
	if eventClassID == -1 {
		log.Println("Warning: Could not find 'Other' class in MyHomeworkSpace.")
	}

	gotEventsForToday := false
	classesDoneToday := []ScheduleClass{}
	for {
		now := Bot_GetNYTimeInUTC()
		currentDayNumber := Schedule_GetDayNum(now)

		if !gotEventsForToday {
			Bot_AppendMyHomeworkSpaceEvents(&schedule, now, eventClassID)
			gotEventsForToday = true
		}

		foundAClass, class, isCurrentlyIn, endTime := Schedule_FindNextClass(schedule, now, currentDayNumber, classesDoneToday)
		if isCurrentlyIn {
			classesDoneToday = append(classesDoneToday, class)
			Bot_SetStatusToClass(class)
			time.Sleep(endTime.Sub(time.Now()))
			Bot_ResetStatus()
		} else {
			if foundAClass {
				// wait for that class
				duration := Schedule_GetNormalTime(class.StartTime, now).Sub(now)
				time.Sleep(duration)
			} else {
				// no more classes left, wait for tomorrow
				midnight, _ := time.Parse("2006-01-02", now.AddDate(0, 0, 1).Format("2006-01-02"))
				duration := midnight.Sub(now)

				// hibernate until midnight
				time.Sleep(duration)
				time.Sleep(time.Second)

				// ok it's tomorrow now
				now = Bot_GetNYTimeInUTC()
				currentDayNumber = Schedule_GetDayNum(now)
				gotEventsForToday = false
				classesDoneToday = []ScheduleClass{}

				if currentDayNumber == 0 {
					// it's monday! reset the schedule, otherwise it will get clogged up with build sessions from yesterday
					schedule = Schedule{}
					json.Unmarshal([]byte(scheduleString), &schedule)
				}
			}
		}
	}
}
