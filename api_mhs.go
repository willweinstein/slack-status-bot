package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const API_MHS_Domain = "https://api-v2.myhomework.space/"
const API_MHS_FormatTime = "2006-01-02"

var API_MHS_AuthToken string
var API_MHS_Connected bool
var API_MHS_ClientID string
var API_MHS_Me map[string]interface{}

func API_MHS_Request(requestType string, path string, params url.Values) (map[string]interface{}, error) {
	url := API_MHS_Domain + path
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

	req.Header.Add("Authorization", "Bearer " + API_MHS_AuthToken)

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

func API_MHS_SignedIn() (bool) {
	response, err := API_MHS_Request("GET", "auth/me", url.Values{})
	if err != nil || response["status"] == "error" {
		return false
	}
	API_MHS_Me = response
	return true
}

func API_MHS_SignOut() {
	// TODO: self-revoke access when this is added to MyHomeworkSpace API
}

func API_MHS_GetRedirectPath() (string) {
	return API_MHS_Domain + "application/requestAuth/" + API_MHS_ClientID
}

func API_MHS_Init() {
	API_MHS_AuthToken = Storage_Get("mhs-auth")
	if API_MHS_AuthToken == "" {
		API_MHS_AuthToken = "invalid"
	}
	API_MHS_Connected = API_MHS_SignedIn()
}