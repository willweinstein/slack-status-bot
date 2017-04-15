package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const API_Slack_Domain = "https://slack.com/api/"

var API_Slack_AuthToken string
var API_Slack_Connected bool
var API_Slack_ClientID string
var API_Slack_ClientSecret string
var API_Slack_Me map[string]interface{}

type API_Slack_StatusInfo struct {
	StatusText string `json:"status_text"`
	StatusEmoji string `json:"status_emoji"`
}

func API_Slack_Request(requestType string, path string, params url.Values) (map[string]interface{}, error) {
	if path != "oauth.access" {
		params.Add("token", API_Slack_AuthToken)
	}

	url := API_Slack_Domain + path
	if (requestType == "GET") {
		url = url + "?" + params.Encode()
	}

	var requestBody io.Reader
	if (requestType == "POST") {
		requestBody = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(requestType, url, requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decodedResponse := map[string]interface{}{}
	err = json.Unmarshal(strResponse, &decodedResponse)
	if err != nil {
		return nil, err
	}

	return decodedResponse, nil
}

func API_Slack_SignedIn() (bool) {
	response, err := API_Slack_Request("GET", "auth.test", url.Values{})
	if err != nil || response["ok"] == false {
		return false
	}
	API_Slack_Me = response
	return true
}

func API_Slack_SignOut() {
	
}

func API_Slack_GetRedirectPath() (string) {
	return "https://slack.com/oauth/authorize?client_id=" + API_Slack_ClientID + "&scope=users.profile:read%20users.profile:write"
}

func API_Slack_GetTokenFromCode(code string) (string, error) {
	resp, err := API_Slack_Request("GET", "oauth.access", url.Values{
		"client_id": { API_Slack_ClientID },
		"client_secret": { API_Slack_ClientSecret },
		"code": { code },
	})
	if err != nil {
		return "", err
	}
	return resp["access_token"].(string), nil
}

func API_Slack_UpdateStatus(info API_Slack_StatusInfo) (error) {
	statusText, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = API_Slack_Request("GET", "users.profile.set", url.Values{
		"profile": { string(statusText) },
	})
	return err
}

func API_Slack_Init() {
	API_Slack_AuthToken = Storage_Get("slack-auth")
	if API_Slack_AuthToken == "" {
		API_Slack_AuthToken = "invalid"
	}
	API_Slack_Connected = API_Slack_SignedIn()
}