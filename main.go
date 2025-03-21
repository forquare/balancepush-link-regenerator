package main

import (
	"fmt"
	gocardless "github.com/forquare/balancepush-gocardless"
	"github.com/forquare/balancepush/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	c := config.GetConfig()
	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.LoadHTMLFiles("templates/error.tmpl", "templates/redirect.tmpl")

	router.GET("/", requisitionHandler)
	router.GET("/redirect", redirectHandler)

	err := router.Run(fmt.Sprintf("%s:%d", c.Requisitioner.Listen.Host, c.Requisitioner.Listen.Port))
	if err != nil {
		logrus.Println(err)
	}
}

func requisitionHandler(gc *gin.Context) {
	c := config.GetConfig()

	client, err := gocardless.NewGoCardlessClient(c.GoCardless.Credentials.SecretID, c.GoCardless.Credentials.SecretKey)
	if err != nil {
		logrus.Fatalf("Error creating client: %v", err)
	}

	// This is needed to get the account details, but once we have them it's good for 90 days
	agreementID, err := client.GetAgreement(c.GoCardless.Bank.Institution)
	if err != nil {
		logrus.New().Fatalf("Error creating agreement: %v", err)
	}

	req, err := client.CreateRequisition(c.GoCardless.Bank.Institution, agreementID, c.Requisitioner.Redirect.Proto, c.Requisitioner.Redirect.Host, strconv.Itoa(c.Requisitioner.Redirect.Port), c.Requisitioner.Redirect.Path)

	if err != nil {
		logrus.Fatalf("Error creating requisition: %v", err)
	}

	gc.Redirect(303, req.Link)
}

func redirectHandler(gc *gin.Context) {
	gc.HTML(http.StatusOK, "redirect.tmpl", gin.H{})
}
