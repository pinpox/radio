package main

import (
	"bytes"
	"context"
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
	Slug        string
	CurrentMeta RadioStationMetadata
}

type RadioStations []RadioStation

var Stations RadioStations = []RadioStation{
	{
		Url:  "https://hirschmilch.de:7000/psytrance.mp3",
		Name: "Hirschmilch Psytrance",
		Slug: "hirschmilch-psytrance",

		CurrentMeta: RadioStationMetadata{
			ArtistName: "TODOArtist",
			TrackName:  "TODOTrackname",
			Updated:    time.Now(),
		},
	},
}




func (s *RadioStations) Update()  {

	// TODO
var hirschStreamUrl string = "https://hirschmilch.de:7000/psytrance.mp3"
	title, err := icymeta.GetCurrentStreamTitle(context.Background(), hirschStreamUrl)
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

func main() {

	Stations.Update()


	var err error
	flag.Parse()

	router := gin.Default()

	router.Static("/static", "./static")

	// Serve HTML page to trigger connection
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	router.POST("/playercontrol", handlePlayerControl)

	router.GET("/audio.mp3", handleAudioStream)
	router.GET("/station/:name", handleRadioStations)

	// Handle WebSocket connections
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// panic(err)
			log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		msgCount := 0

		// var metadataTemplate = template.Must(template.ParseFiles("./templates/metadata.html"))

		for {

			// Read message from client
			// messageType, p, err := conn.ReadMessage()
			// if err != nil {
			// 	// panic(err)
			// 	log.Printf("%s, error while reading message\n", err.Error())
			// 	c.AbortWithError(http.StatusInternalServerError, err)
			// 	break
			// }

			// newMsg := ` <div id="artist" hx-swap-oob="true"> Saved! ` + strconv.Itoa(msgCount) + ` </div> `

			// homeTemplate.Execute(w, r.Host)

			tmpl, err := template.ParseFiles("templates/metadata.html")
			if err != nil {
				log.Fatalf("template parsing: %s", err)
			}

			// Render the template with the message as data.
			var renderedMetadata bytes.Buffer

			data := struct {
				Count       string
				Status      string
				ArtistName  string
				TrackName   string
				StationName string
			}{
				Count:       "test" + strconv.Itoa(msgCount),
				Status:      "test" + strconv.Itoa(msgCount),
				ArtistName:  "test" + strconv.Itoa(msgCount),
				TrackName:   "test" + strconv.Itoa(msgCount),
				StationName: "test" + strconv.Itoa(msgCount),
			}

			err = tmpl.Execute(&renderedMetadata, data)
			if err != nil {
				log.Fatalf("template execution: %s", err)
			}

			msgCount += 1

			log.Println("writing message", renderedMetadata.String())

			err = conn.WriteMessage(websocket.TextMessage, renderedMetadata.Bytes())

			if err != nil {
				// panic(err)
				log.Printf("%s, error while writing message\n", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				break
			}

			time.Sleep(time.Second * 2)
		}
	})

	err = router.Run(address)
	if err != nil {
		panic(err)
	}
}

func handlePlayerControl(c *gin.Context) {
	//TODO
}


func handleRadioStations(c *gin.Context) {
		name := c.Param("name")


	read, write := io.Pipe()

	go func() {
		defer write.Close()
		resp, err := http.Get(hirschStreamUrl)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		io.Copy(write, resp.Body)
	}()

	io.Copy(c.Writer, read)

}

func handleAudioStream(c *gin.Context) {
	read, write := io.Pipe()

	go func() {
		defer write.Close()
		resp, err := http.Get(hirschStreamUrl)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		io.Copy(write, resp.Body)
	}()

	io.Copy(c.Writer, read)
}
