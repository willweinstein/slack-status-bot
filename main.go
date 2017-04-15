package main

import (
	"flag"
	"log"
)

func main() {
	port := 0
	forceManage := false

	flag.IntVar(&port, "port", 4532, "The port to run the management interface on.")
	flag.BoolVar(&forceManage, "manage", false, "Force running the management interface, even if it's not needed.")

	flag.StringVar(&API_MHS_ClientID, "mhs-client", "", "The client ID for MyHomeworkSpace.")

	flag.StringVar(&API_Slack_ClientID, "slack-client-id", "", "The client ID for Slack.")
	flag.StringVar(&API_Slack_ClientSecret, "slack-client-secret", "", "The client secret for Slack.")

	flag.Parse()

	log.Println("slack-status-bot")

	if API_MHS_ClientID == "" {
		log.Println("Please specify a MyHomeworkSpace client ID.")
		return
	}
	if API_Slack_ClientID == "" || API_Slack_ClientSecret == "" {
		log.Println("Please specify a Slack client ID and secret.")
		return
	}

	Storage_Init()

	API_MHS_Init()
	API_Slack_Init()

	needToManage := !Bot_CanActivate()

	if forceManage || needToManage {
		if needToManage {
			log.Println("Some services aren't connected, starting management UI...")
		}
		log.Printf("Visit http://localhost:%d/ in your browser.", port)
		Management_Init(port)
	}

	log.Println("yay bot time")
	Bot_Start()
}