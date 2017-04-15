package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"html"
	"net/url"
	"net/http"
	"strings"
	"time"
)

const API_Schedules_URL = "https://schedules.dalton.org/roux/index.php"
const API_Schedules_FormatTime = "2006-01-02 15:04:05"

var API_Schedules_Key string

func API_Schedules_Request(request string) (string, error) {
	params := url.Values{
		"rouxRequest": { request },
	}

	req, err := http.NewRequest("POST", API_Schedules_URL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(strResponse), nil
}

type API_Schedules_LoginResult struct {
	Status int `xml:"status,attr"`
	Key string `xml:"key"`
}

type API_Schedules_LoginResponse struct {
	Result API_Schedules_LoginResult `xml:"result"`
}

type API_Schedules_SchedulePeriodSection struct {
	ID string `xml:"id,attr"`
	Name string `xml:"name"`
}

type API_Schedules_ScheduleTerm struct {
	Number int `xml:"number"`
	Name string `xml:"name"`
}

type API_Schedules_SchedulePeriod struct {
	DayNumber int `xml:"DAY_NUMBER"`
	Location string `xml:"location"`
	Start string `xml:"start"`
	End string `xml:"end"`
	Section API_Schedules_SchedulePeriodSection `xml:"section"`
	Term API_Schedules_ScheduleTerm `xml:"term"`
}

type API_Schedules_ScheduleResult struct {
	Status int `xml:"status,attr"`
	Periods []API_Schedules_SchedulePeriod `xml:"period"`
}

type API_Schedules_ScheduleResponse struct {
	Result API_Schedules_ScheduleResult `xml:"result"`
}

func API_Schedules_SignIn(username string, password string) (string) {
	strResponse, err := API_Schedules_Request("<request><key></key><action>authenticate</action><credentials><username>" + html.EscapeString(username) + "</username><password type=\"plaintext\">" + html.EscapeString(password) + "</password></credentials></request>")
	if err != nil {
		return "Network error"
	}
	response := API_Schedules_LoginResponse{}
	err = xml.Unmarshal([]byte(strResponse), &response)
	if err != nil {
		return "XML error"
	}
	if response.Result.Status != 200 {
		return "Invalid username/password."
	}
	API_Schedules_Key = response.Result.Key
	return ""
}

func API_Schedules_FetchAndSave() (error) {
	owner := strings.Split(API_Schedules_Key, ":")[3]
	strResponse, err := API_Schedules_Request("<request><key>" + API_Schedules_Key + "</key><action>selectStudentSchedule</action><ID>" + owner + "</ID><academicyear>2017</academicyear></request>")
	if err != nil {
		return err
	}
	response := API_Schedules_ScheduleResponse{}
	err = xml.Unmarshal([]byte(strResponse), &response)
	if err != nil {
		return err
	}
	
	daltonSchedule := Schedule{}

	for _, period := range response.Result.Periods {
		if period.Term.Number != 1 {
			continue
		}

		class := ScheduleClass{}

		class.Name = period.Section.Name
		class.Room = period.Location
		class.StartTime, _ = time.Parse(API_Schedules_FormatTime, period.Start)
		endTime, _ := time.Parse(API_Schedules_FormatTime, period.End)
		class.Duration = endTime.Sub(class.StartTime)

		daltonSchedule.Days[period.DayNumber].Classes = append(daltonSchedule.Days[period.DayNumber].Classes, class)
	}

	scheduleText, err := json.Marshal(daltonSchedule)
	if err != nil {
		return err
	}

	Storage_Set("dalton-schedule", string(scheduleText))

	return nil
}

func API_Schedules_Init() {
	
}