package main

import (
	"fmt"
	"log"
	"html"
	"net"
	"net/http"
	"strconv"

	"github.com/hydrogen18/stoppableListener"
)

var Management_Socket *stoppableListener.StoppableListener

func Management_Handler_StopManagement(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !Bot_CanActivate() {
		fmt.Fprintf(w, "You haven't signed in to every service yet; the bot cannot start.")
		fmt.Fprintf(w, "<br /> <a href='/'>Go back</a>")
	} else {
		fmt.Fprintf(w, "Management UI stopped.")
		Management_Socket.Stop()
		Management_Socket.TCPListener.Close()
	}
}

/*
 * Schedule
 */
func Management_Handler_Schedule_SignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		fmt.Fprintf(w, "Username and password are required.<br /><a href='/schedule_signin'>Go back</a>")
		return
	}
	message := API_Schedules_SignIn(username, password)
	if message != "" {
		fmt.Fprintf(w, message)
		fmt.Fprintf(w, "<br /><a href='/schedule_signin'>Go back</a>")
		return
	}
	err := API_Schedules_FetchAndSave()
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "There was an error fetching your schedule.<br /><a href='/schedule_signin'>Go back</a>")
		return
	}
	fmt.Fprintf(w, "Your schedule has been downloaded.<br /><a href='/'>Go back</a>")
}

func Management_Handler_Schedule_SignInPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<form action='/schedule_signin_h' method='POST'>")
	fmt.Fprintf(w, "<input type='text' name='username' placeholder='Username' /><br />")
	fmt.Fprintf(w, "<input type='password' name='password' placeholder='Password' /><br />")
	fmt.Fprintf(w, "<input type='submit' />")
	fmt.Fprintf(w, "</form>")
}

/*
 * MyHomeworkSpace
 */
func Management_Handler_MHS_Callback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["token"][0]
	if token == "" {
		fmt.Fprintf(w, "No token parameter.")
		return
	}
	Storage_Set("mhs-auth", token)
	API_MHS_Init()
	http.Redirect(w, r, "/", http.StatusFound)
}

func Management_Handler_MHS_SignOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if API_MHS_Connected {
		API_MHS_SignOut()
		fmt.Fprintf(w, "Signed out successfully. <br /> <a href='/'>Go back</a>")
	} else {
		fmt.Fprintf(w, "You aren't signed in to MyHomeworkSpace!")
	}
}

func Management_Handler_MHS_SignIn(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, API_MHS_GetRedirectPath(), http.StatusFound)
}

/*
 * Slack
 */
func Management_Handler_Slack_Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query()["code"][0]
	if code == "" {
		fmt.Fprintf(w, "No code parameter.")
		return
	}
	token, err := API_Slack_GetTokenFromCode(code)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Error getting Slack access token.")
		return
	}
	Storage_Set("slack-auth", token)
	API_Slack_Init()
	http.Redirect(w, r, "/", http.StatusFound)
}

func Management_Handler_Slack_SignOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if API_Slack_Connected {
		API_Slack_SignOut()
		fmt.Fprintf(w, "Signed out successfully. <br /> <a href='/'>Go back</a>")
	} else {
		fmt.Fprintf(w, "You aren't signed in to Slack!")
	}
}

func Management_Handler_Slack_SignIn(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, API_Slack_GetRedirectPath(), http.StatusFound)
}

func Management_Handler_Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h2>slack-status-bot</h2>")

	fmt.Fprintf(w, "Schedule: <a href='/schedule_signin'>Fetch schedule</a><br />")

	fmt.Fprintf(w, "MyHomeworkSpace: ")
	if API_MHS_Connected {
		fmt.Fprintf(w, "Signed in as user <strong>%s</strong> <a href='/mhs_signout'>Sign out</a>", html.EscapeString(API_MHS_Me["name"].(string)))
	} else {
		fmt.Fprintf(w, "Not signed in! <a href='/mhs_signin'>Sign in</a>")
	}
	fmt.Fprintf(w, "<br />")

	fmt.Fprintf(w, "Slack: ")
	if API_Slack_Connected {
		fmt.Fprintf(w, "Signed in as user <strong>%s</strong> <a href='/slack_signout'>Sign out</a>", html.EscapeString(API_Slack_Me["user"].(string)))
	} else {
		fmt.Fprintf(w, "Not signed in! <a href='/slack_signin'>Sign in</a>")
	}
	fmt.Fprintf(w, "<br />")

	fmt.Fprintf(w, "<br /><a href='/stop_management'>Stop management and start bot</a>")
}

func Management_Init(port int) {
	http.HandleFunc("/stop_management", Management_Handler_StopManagement)

	http.HandleFunc("/schedule_signin_h", Management_Handler_Schedule_SignIn)
	http.HandleFunc("/schedule_signin", Management_Handler_Schedule_SignInPage)

	http.HandleFunc("/mhs_cb", Management_Handler_MHS_Callback)
	http.HandleFunc("/mhs_signout", Management_Handler_MHS_SignOut)
	http.HandleFunc("/mhs_signin", Management_Handler_MHS_SignIn)

	http.HandleFunc("/slack_cb", Management_Handler_Slack_Callback)
	http.HandleFunc("/slack_signout", Management_Handler_Slack_SignOut)
	http.HandleFunc("/slack_signin", Management_Handler_Slack_SignIn)

	http.HandleFunc("/", Management_Handler_Index)

	log.Println("Management UI listening on port", port)

	socket, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	Management_Socket, err = stoppableListener.New(socket)
	if err != nil {
		panic(err)
	}

	http.Serve(Management_Socket, nil)
	Management_Socket.TCPListener.Close()
}