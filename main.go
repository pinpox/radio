package main

import (
	"bytes"
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
)

type RadioStationMetadata struct {
	// Bitrate string
	Title   string
	Updated time.Time
}

type RadioStation struct {
	Url         string
	Name        string
	CurrentMeta *RadioStationMetadata
}

func (rs *RadioStation) Update() {

	if title, err := rs.GetStreamTitle(); err != nil {
		log.Println(err)
	} else {
		if rs.CurrentMeta.Title != title {
			rs.CurrentMeta.Title = title
			rs.CurrentMeta.Updated = time.Now()
		}
	}
}

type RadioStations []RadioStation

// TODO read from config file
var Stations RadioStations = []RadioStation{
	{
		Url:  "https://hirschmilch.de:7000/psytrance.mp3",
		Name: "Hirschmilch Psytrance",

		CurrentMeta: &RadioStationMetadata{
			Title:   "",
			Updated: time.Now(),
		},
	},

	{
		Url:  "https://hirschmilch.de:7000/progressive.mp3",
		Name: "Hirschmilch Progressive",

		CurrentMeta: &RadioStationMetadata{
			Title:   "",
			Updated: time.Now(),
		},
	},
}

func (s *RadioStations) Update() {
	for {
		for _, v := range *s {
			v.Update()
		}

		time.Sleep(time.Second * 5)
	}
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

	go Stations.Update()

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

func updateClientPlayer(stationIndex chan int, conn *websocket.Conn) {

	userStationIndex := 0

	for {
		var jsonMsg gin.H
		if err := conn.ReadJSON(&jsonMsg); err != nil {
			log.Println("failed json parsing: ", err)
			return
		} else {
			if val, ok := jsonMsg["action"]; ok {
				if val == "next" {
					userStationIndex = (userStationIndex + 1) % len(Stations)
				}

				if val == "previous" {
					userStationIndex = (len(Stations) + userStationIndex - 1) % len(Stations)
				}

				log.Println("updating player")
				if err := sendTemplateWebsocket(conn, "templates/player.html",
					gin.H{"Url": userStationIndex}); err != nil {
					log.Println(err)
				}
				stationIndex <- userStationIndex
			}
		}
	}
}

func updateClientMetadata(userStation RadioStation, conn *websocket.Conn) error {

	if err := sendTemplateWebsocket(conn, "templates/metadata.html", gin.H{
		"ArtistName":  userStation.CurrentMeta.Title,
		"StationName": userStation.Name,
	}); err != nil {
		log.Printf("%s, error while writing message\n", err.Error())
		return err
	}
	return nil

}

func handleWebSocket(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	stationIndex := make(chan int)

	// Read messages from client to control the player
	go updateClientPlayer(stationIndex, conn)

	// Keep updating client metadata periodically
	userStation := Stations[0]
	var lastMetaUpdate time.Time
	for {

		// Check if station has changed
		select {
		case i := <-stationIndex:
			userStation = Stations[i]
		default:
			// fmt.Println("no message received")
		}

		// Update client's metadata if it's newer
		if userStation.CurrentMeta.Updated.After(lastMetaUpdate) {
			log.Println("TIME: PUDATING ")
			if err = updateClientMetadata(userStation, conn); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			lastMetaUpdate = time.Now()
		}

		time.Sleep(time.Second * 2)
	}
}

func sendTemplateWebsocket(conn *websocket.Conn, templateName string, data gin.H) error {

	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	// Render the template with the message as data.
	var renderedMetadata bytes.Buffer

	err = tmpl.Execute(&renderedMetadata, data)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}

	log.Println("writing message", renderedMetadata.String())
	return conn.WriteMessage(websocket.TextMessage, renderedMetadata.Bytes())
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
