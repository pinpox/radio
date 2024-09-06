package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	"gopkg.in/ini.v1"
	// "github.com/gin-contrib/pprof"
	// _ "net/http/pprof"
)

//go:embed static templates
var f embed.FS

var (
	Stations      RadioStations
	address       string
	stationsFile  string
	proxyStations bool = false
	messages      messageBuffer
	templ         *template.Template
	idCounter     atomic.Int64
	m             *melody.Melody
)

func init() {

	for i := range usernames {
		j := rand.Intn(i + 1)
		usernames[i], usernames[j] = usernames[j], usernames[i]
	}

	address = os.Getenv("RADIO_ADDRESS")
	stationsFile = os.Getenv("RADIO_STATIONFILE")

	if address == "" || stationsFile == "" {
		log.Fatal("Set environment variables RADIO_ADDRESS and RADIO_STATIONFILE")
	}

	messages = messageBuffer{}

	proxyStations, _ = strconv.ParseBool(os.Getenv("RADIO_PROXY_STATIONS"))

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

func main() {

	var err error

	m = melody.New()

	router := gin.Default()

	templ = template.Must(template.New("").ParseFS(f, "templates/*"))
	router.SetHTMLTemplate(templ)

	// pprof.Register(router)
	// m.Config.MaxMessageSize = 256

	staticFS, err := fs.Sub(f, "static")
	if err != nil {
		panic(err)
	}
	router.StaticFS("/static", http.FS(staticFS))

	sUrl := Stations[0].Url

	if proxyStations {
		sUrl = fmt.Sprintf("/station/0")
	}

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Url":      sUrl,
			"Messages": messages.Get(),
			// "News": "Currently no news. This is only a test message",
		})
	})

	if proxyStations {
		router.GET("/station/:index", handleRadioStations)
	}

	router.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(handlerWsMessage)

	m.HandleConnect(func(s *melody.Session) {
		id := idCounter.Add(1)
		s.Set("id", id)
		s.Set("station", 0)
	})

	go func() {
		for {

			for k, v := range Stations {

				v.Update()

				tData, err := renderTemplate("metadata.html", gin.H{
					"Url":          v.Url,
					"StationName":  v.StationName,
					"StationTitle": v.StationTitle,
					"NumListeners": m.Len(),
				})

				if err != nil {
					continue
				}

				// broadcast station
				m.BroadcastFilter(
					tData,
					func(s *melody.Session) bool {
						return (getSessionStation(s) == k)
					},
				)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	err = router.Run(address)
	if err != nil {
		panic(err)
	}
}

func renderTemplate(name string, data any) ([]byte, error) {

	var renderedTemplate bytes.Buffer
	err := templ.ExecuteTemplate(&renderedTemplate, name, data)
	if err != nil {
		log.Println("Failed to render template:", name)
		log.Println(err)
		return nil, err
	}

	return renderedTemplate.Bytes(), nil
}

func getSessionID(s *melody.Session) int {
	if s, exists := s.Get("id"); exists {
		if idInt, ok := s.(int64); ok {
			return int(idInt)
		}
	}

	log.Println("Error: no user id found in session")
	return 0
}

func getSessionStation(s *melody.Session) int {

	if s, exists := s.Get("station"); exists {
		if stationInt, ok := s.(int); ok {
			return stationInt
		}
	}

	return 0
}

type WsMessage struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	Headers struct {
		HXRequest     string `json:"HX-Request"`
		HXTrigger     string `json:"HX-Trigger"`
		HXTriggerName any    `json:"HX-Trigger-Name"`
		HXTarget      string `json:"HX-Target"`
		HXCurrentURL  string `json:"HX-Current-URL"`
	} `json:"HEADERS"`
}

func handlerWsMessage(s *melody.Session, msg []byte) {

	station := getSessionStation(s)

	wsMsg := WsMessage{}
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		log.Println("Unparsable message:", string(msg))
		return
	}

	switch wsMsg.Action {
	case "next":
		station = (station + 1) % len(Stations)
		s.Set("station", station)

	case "previous":
		station -= 1

		if station < 0 {
			station = len(Stations) - 1
		}
		s.Set("station", station)

	case "chat":
		if wsMsg.Message != "" {

			tData, err := renderTemplate("messages.html", gin.H{
				"Text": wsMsg.Message,
				"User": usernames[getSessionID(s)%len(usernames)],
			})

			if err != nil {
				return
			}

			messages.Add(getSessionID(s), wsMsg.Message)

			m.Broadcast(tData)
		}

	default:
		log.Println("Unhandled message:", wsMsg, string(msg), "with action", wsMsg.Action)
	}

	sUrl := Stations[station].Url

	if proxyStations {
		sUrl = fmt.Sprintf("/station/%v", station)
	}

	if err := sendTemplateWebsocket(s, "player.html",
		gin.H{"Url": sUrl}); err != nil {
		log.Println(err)
	}
}

func sendTemplateWebsocket(s *melody.Session, templateName string, data gin.H) error {

	renderedTemplate, err := renderTemplate(templateName, data)

	if err != nil {
		log.Fatalf("template execution failed: %s", err)
	}

	return s.Write(renderedTemplate)
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
