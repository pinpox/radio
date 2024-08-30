package main

import (
	"bytes"
	"context"
	// "encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/inspectorgoget/icymeta"
)

type RadioStationMetadata struct {
	// Bitrate string
	ArtistName string
	TrackName  string
	Updated    time.Time
}

type RadioStation struct {
	Url         string
	Name        string
	CurrentMeta RadioStationMetadata
}

type RadioStations []RadioStation

var Stations RadioStations = []RadioStation{
	{
		Url:  "https://hirschmilch.de:7000/psytrance.mp3",
		Name: "Hirschmilch Psytrance",

		CurrentMeta: RadioStationMetadata{
			ArtistName: "TODOArtist",
			TrackName:  "TODOTrackname",
			Updated:    time.Now(),
		},
	},

	{
		Url:  "https://hirschmilch.de:7000/progressive.mp3",
		Name: "Hirschmilch Progressive",

		CurrentMeta: RadioStationMetadata{
			ArtistName: "TODOArtist",
			TrackName:  "TODOTrackname",
			Updated:    time.Now(),
		},
	},
}

func (s *RadioStations) Update() {

	// TODO implement, this is a placeholder!

	title, err := icymeta.GetCurrentStreamTitle(context.Background(), Stations[0].Url)
	// icymeta.ReadMeta()

	if err != nil {
		panic(err)
	}

	log.Printf("Current stream title: %s\n", title)

}

var (
	address string
)

func init() {
	flag.StringVar(&address, "a", ":7000", "address to use")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PlayerCommand int

const (
	Next PlayerCommand = iota
	Previous
	// Pause
)

func main() {

	Stations.Update()

	var err error
	flag.Parse()

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) { c.File("index.html") })
	router.GET("/station/:index", handleRadioStations)
	router.GET("/ws", handleWebSocket)

	err = router.Run(address)
	if err != nil {
		log.Println("server paniced. Doh!")
		panic(err)
	}
}

func handleWebSocket(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// panic(err)
		log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	msgCount := 0

	messages := make(chan PlayerCommand)
	userStationIndex := 0

	// Read messages from client to control the player
	go func() {
		for {

			// Read message from client
			// _, p, err := conn.ReadMessage()
			// if err != nil {
			// 	// panic(err)
			// 	log.Printf("%s, error while reading message\n", err.Error())
			// 	c.AbortWithError(http.StatusInternalServerError, err)
			// 	break
			// }

			var jsonMsg gin.H
			if err = conn.ReadJSON(&jsonMsg); err != nil {
				log.Println("failed json parsing: %s", err)
				return
			} else {
				if val, ok := jsonMsg["action"]; ok {
					if val == "next" {
						messages <- Next
					}

					if val == "previous" {
						messages <- Previous
					}
				}
			}
		}
	}()

	for {

		select {
		case msg := <-messages:
			log.Println("received message", msg)
			if msg == Next {
				userStationIndex = (userStationIndex + 1) % len(Stations)
				c.HTML(http.StatusOK, "templates/player.html", gin.H{"Url": userStationIndex})
			}
			if msg == Previous {
				userStationIndex = (userStationIndex - 1) % len(Stations)
				c.HTML(http.StatusOK, "templates/player.html", gin.H{"Url": userStationIndex})
			}
		default:
			log.Println("no message received")
		}

		userStation := Stations[userStationIndex]

		tmpl, err := template.ParseFiles("templates/metadata.html")
		if err != nil {
			log.Fatalf("template parsing: %s", err)
		}

		// Render the template with the message as data.
		var renderedMetadata bytes.Buffer

		data := struct {
			Count       string
			ArtistName  string
			TrackName   string
			StationName string
		}{
			Count:       strconv.Itoa(msgCount),
			TrackName:   userStation.CurrentMeta.TrackName,
			ArtistName:  userStation.CurrentMeta.ArtistName,
			StationName: userStation.Name,
		}

		err = tmpl.Execute(&renderedMetadata, data)
		if err != nil {
			log.Fatalf("template execution: %s", err)
		}

		log.Println("writing message", renderedMetadata.String())

		err = conn.WriteMessage(websocket.TextMessage, renderedMetadata.Bytes())

		if err != nil {
			// panic(err)
			log.Printf("%s, error while writing message\n", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			break
		}

		msgCount += 1
		time.Sleep(time.Second * 2)
	}
}

func handleRadioStations(c *gin.Context) {

	streamIndex, err := strconv.Atoi(c.Param("index"))
	if err != nil || streamIndex >= len(Stations) {
		log.Println("Client tried to access invalid radio station:", streamIndex)
		return
	}

	streamUrl := Stations[streamIndex]

	read, write := io.Pipe()

	go func() {
		defer write.Close()
		resp, err := http.Get(streamUrl.Url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		io.Copy(write, resp.Body)
	}()

	io.Copy(c.Writer, read)

}
