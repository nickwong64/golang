package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nexmo-community/nexmo-go"
)

func main() {

	// Auth
	auth := nexmo.NewAuthSet()
	auth.SetAPISecret("api_key", "api_secret")

	// Init Nexmo
	client := nexmo.NewClient(http.DefaultClient, auth)

	// SMS
  // change the from, to and text here
	smsContent := nexmo.SendSMSRequest{
		From: "LIS Test",
		To:   "852xxxxxxxx
		Text: "This is a message sent from Go!",
	}

	smsResponse, _, err := client.SMS.SendSMS(smsContent)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", smsResponse.Messages[0].Status)
}
