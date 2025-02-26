package main

import (
	"encoding/json"
	"fmt"
	"github.com/forquare/balancepush/config"
	"github.com/forquare/balancepush/gocardless"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

const (
	redirectPort = ":3000"
)

func main() {
	c := config.GetConfig()

	client, err := gocardless.NewGoCardlessClient(c.GoCardless.Credentials.SecretID, c.GoCardless.Credentials.SecretKey)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// This is needed to get the account details, but I think once we have them
	// it is no longer needed for 180 days...
	agreementID, err := client.GetAgreement(c.GoCardless.Bank.Institution)
	if err != nil {
		log.Fatalf("Error creating agreement: %v", err)
	}

	req, err := client.CreateRequisition(c.GoCardless.Bank.Institution, agreementID)
	if err != nil {
		log.Fatalf("Error creating requisition: %v", err)
	}

	var requisition gocardless.Requisition
	go openBrowser(req.Link)

	ch := make(chan bool, 1)

	go catchRedirect(redirectPort, ch)

	<-ch

	for req.Status == "CR" {
		req, err = client.GetRequisition(req.ID)
		time.Sleep(5 * time.Second)
		if err != nil {
			log.Fatalf("Error getting requisition: %v", err)
		} else {
			requisition = req
		}
	}

	jsonPretty, err := json.MarshalIndent(requisition, "", "  ")
	fmt.Println("Institution Details:", string(jsonPretty))

	return
}
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func catchRedirect(port string, ch chan bool) {
	handler := func(chan bool) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ch <- true
			w.Write([]byte("You can close this window now"))
		})
	}
	http.Handle("/", handler(ch))

	err := http.ListenAndServe(port, nil)

	if err != nil {
		panic(err)
	}
}
