package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"photos-app/cfg"
	"photos-app/cloud"
	"photos-app/db"
	"photos-app/models"
	"time"
)

type appController struct {
	Config cfg.Config
	Mongo  db.MongoBongo
	Client http.Client
	AWS    cloud.AWSClient
}

var con appController

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	c, err := cfg.GetConfig("config.json")
	if err != nil {
		log.Fatalf("error getting configuration [ %v ]", err)
	}

	d := db.InitDatabase(c)
	cli := initializeClient(c)
	aws := cloud.InitAWSClient("neon-photos")

	con = appController{
		Config: c,
		Mongo:  d,
		Client: cli,
		AWS:    aws,
	}

	gob.Register(&models.User{})
	gob.Register(&models.Post{})
	gob.Register(&models.Comment{})

	r := registerRoutes()
	r.HTMLRender = loadTemplates("./templates")

	port := os.Getenv("PHOTOS_APP_PORT")
	if port == "" {
		port = c.Port
	}

	log.Printf("Listening on port %s\n", port)
	r.Run(":" + port)

}

func initializeClient(c cfg.Config) (client http.Client) {
	client = http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       60 * time.Second,
	}
	return
}
