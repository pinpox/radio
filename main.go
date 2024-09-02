package main

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"gopkg.in/ini.v1"
)

var (
	Stations     RadioStations
	address      string
	stationsFile string
)

func init() {

	address = os.Getenv("RADIO_ADDRESS")
	stationsFile= os.Getenv("RADIO_STATIONFILE")

	if address == ""|| stationsFile== "" {
		log.Fatal("Set environment variables RADIO_ADDRESS and RADIO_STATIONFILE")
	}

	inidata, err := ini.Load(stationsFile)
	if err != nil {
		panic(err)
	}

	for _, v := range inidata.Sections() {
		if v.Name() != "DEFAULT" {
			Stations.Add(v.Name(), v.Key("url").Value())
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//go:embed static templates
var f embed.FS

func main() {

	go Stations.Update()

	var err error

	router := gin.Default()

	templ := template.Must(template.New("").ParseFS(f, "templates/*"))
	router.SetHTMLTemplate(templ)

	staticFS, err := fs.Sub(f, "static")
	if err != nil {
		panic(err)
	}
	router.StaticFS("/static", http.FS(staticFS))

	router.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "index.html", gin.H{}) })
	router.GET("/station/:index", handleRadioStations)
	router.GET("/ws", handleWebSocket)

	err = router.Run(address)
	if err != nil {
		panic(err)
	}
}

func updateClientPlayer(stationIndex chan int, conn *MutexConn) {

	userStationIndex := 0

	for {
		var jsonMsg gin.H
		if err := conn.Sock.ReadJSON(&jsonMsg); err != nil {
			log.Println("failed json parsing: ", err)
			return
		} else {
			if val, ok := jsonMsg["action"]; ok {

				// TODO implement messages

				if val == "next" {
					userStationIndex = (userStationIndex + 1) % len(Stations)
				}

				if val == "previous" {
					userStationIndex = (len(Stations) + userStationIndex - 1) % len(Stations)
				}

				if err := sendTemplateWebsocket(conn, "templates/player.html",
					gin.H{"Url": userStationIndex}); err != nil {
					log.Println(err)
				}
				stationIndex <- userStationIndex
			}
		}
	}
}

func updateClientMetadata(userStation RadioStation, conn *MutexConn) error {

	if err := sendTemplateWebsocket(conn, "templates/metadata.html", gin.H{
		"StationTitle": userStation.CurrentMeta.Title,
		"StationName":  userStation.Name,
	}); err != nil {
		log.Printf("%s, error while writing message\n", err.Error())
		return err
	}
	return nil

}

type MutexConn struct {
	Sock *websocket.Conn
	Mut   sync.Mutex
}

func (m *MutexConn) send(mtype int, data []byte) error {
	m.Mut.Lock()
	defer m.Mut.Unlock()
	return m.Sock.WriteMessage(mtype, data)
}

func handleWebSocket(c *gin.Context) {

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	conn := MutexConn{
		Sock: wsConn,
		Mut:   sync.Mutex{},
	}

	stationIndex := make(chan int)

	// Read messages from client to control the player
	go updateClientPlayer(stationIndex, &conn)

	// Keep updating client metadata periodically
	userStation := Stations[0]
	var lastMetaUpdate time.Time
	for {

		// Check if station has changed
		select {
		case i := <-stationIndex:
			userStation = Stations[i]
			// Zero lastMetaUpdate to force update
			lastMetaUpdate = time.Time{}
		default:
		}

		// Update client's metadata if it's newer
		if userStation.CurrentMeta.Updated.After(lastMetaUpdate) {
			if err = updateClientMetadata(userStation, &conn); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			lastMetaUpdate = time.Now()
		}

		time.Sleep(time.Second * 2)
	}
}

func sendTemplateWebsocket(conn *MutexConn, templateName string, data gin.H) error {

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

	return conn.send(websocket.TextMessage, renderedMetadata.Bytes())

	// // log.Println("writing message", renderedMetadata.String())
	// return conn.WriteMessage(websocket.TextMessage, renderedMetadata.Bytes())
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
