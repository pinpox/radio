package main

import (
	"io"
	"flag"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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
	flag.Parse()

	router := gin.Default()

	// Serve HTML page to trigger connection
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	router.GET("/audio.mp3",  handleAudioStream)

	// Handle WebSocket connections
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// panic(err)
			log.Printf("%s, error while Upgrading websocket connection\n", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for {
			// Read message from client
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				// panic(err)
				log.Printf("%s, error while reading message\n", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				break
			}

			// Echo message back to client
			err = conn.WriteMessage(messageType, p)
			if err != nil {
				// panic(err)
				log.Printf("%s, error while writing message\n", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				break
			}
		}
	})

	err := router.Run(address)
	if err != nil {
		panic(err)
	}
}


func handleAudioStream(c *gin.Context) {
	read, write := io.Pipe()

	go func() {
		defer write.Close()
		resp, err := http.Get("https://hirschmilch.de:7000/psytrance.mp3"); if err != nil {
			return
		}
		defer resp.Body.Close()
		io.Copy(write, resp.Body)
	}()

	io.Copy(c.Writer, read)
}
