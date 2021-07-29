package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type minion struct {
	Hostname string
	Port     string
	Username string
	Password string
	Term     string
}
type TTYSize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

var WebsocketMessageType = map[int]string{
	websocket.BinaryMessage: "binary",
	websocket.TextMessage:   "text",
	websocket.CloseMessage:  "close",
	websocket.PingMessage:   "ping",
	websocket.PongMessage:   "pong",
}

func loginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	}
}

func indexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "hello Ted",
		})
	}
}

func termHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "term.html", gin.H{})
	}
}

func connectHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var login minion
		err := c.BindJSON(&login)
		if err != nil {
			fmt.Printf("bind json error%+v\n", err)
			c.JSON(400, gin.H{
				"msg": fmt.Sprintf("%+v", err),
			})
		} else {
			fmt.Printf("%+v\n", login)
			c.JSON(200, gin.H{
				"hello": "ted",
			})

		}
	}
}

var ws = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connectionErrorLimit = 10
var maxBufferSizeBytes = 1024 * 1024
var keepalivePingTimeout = 20 * time.Second

func wsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		conn, err := ws.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Errorf("failed to upgrade connection: %s", err)
			return
		}

		cmd := exec.Command("/bin/bash", "-l")
		cmd.Env = append(os.Environ(), "TERM=xterm")

		tty, err := pty.Start(cmd)
		if err != nil {
			message := fmt.Sprintf("failed to start tty: %s", err)
			log.Error(message)
			conn.WriteMessage(websocket.TextMessage, []byte(message))
			return
		}

		defer func() {
			log.Info("stopping tty...")
			if err := cmd.Process.Kill(); err != nil {
				log.Warnf("failed to kill process: %s", err)
			}
			if _, err := cmd.Process.Wait(); err != nil {
				log.Warnf("failed to wait for process to exit: %s", err)
			}
			if err := tty.Close(); err != nil {
				log.Warnf("failed to close spawned tty gracefully: %s", err)
			}
			if err := conn.Close(); err != nil {
				log.Warnf("failed to close webscoket connection: %s", err)
			}
		}()

		var connectionClosed bool
		var wg sync.WaitGroup

		wg.Add(1)
		lastPongTime := time.Now()
		conn.SetPongHandler(func(msg string) error {
			lastPongTime = time.Now()
			return nil
		})

		go func() {
			for {
				if err := conn.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
					log.Error("failed to write ping message")
					return
				}
				time.Sleep(keepalivePingTimeout / 2)
				if time.Since(lastPongTime) > keepalivePingTimeout {
					log.Warn("failed to get response from ping, disconnec now...")
					wg.Done()
					return
				}
				log.Debug("got ping")
			}
		}()

		// tty >> xterm.js
		go func() {
			errorCounter := 0
			for {
				if errorCounter > connectionErrorLimit {
					wg.Done()
					break
				}
				buffer := make([]byte, maxBufferSizeBytes)
				readLength, err := tty.Read(buffer)
				if err != nil {
					log.Errorf("failed to read from tty: %s", err)
					if err := conn.WriteMessage(websocket.TextMessage, []byte("bye!")); err != nil {
						log.Errorf("failed to say goodbye: %s", err)
					}
					wg.Done()
					return
				}
				if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:readLength]); err != nil {
					log.Errorf("failed to send %v bytes from tty to xterm", readLength)
					errorCounter++
					continue
				}
				log.Tracef("sent message of size %v bytes from tty to xterm", readLength)
				errorCounter = 0
			}
		}()

		// tty << xterm.js
		go func() {
			for {
				msgType, data, err := conn.ReadMessage()
				if err != nil {
					if !connectionClosed {
						log.Errorf("failed to get next reader: %s", err)
					}
					return
				}
				dataLength := len(data)
				dataBuffer := bytes.Trim(data, "\x00")
				dataType, ok := WebsocketMessageType[msgType]
				if !ok {
					dataType = "unknown"
				}

				log.Infof("received %s (type: %v) message of size %v byte(s) from xterm.js with key sequence: %v", dataType, msgType, dataLength, dataBuffer)

				if dataLength == -1 {
					log.Errorf("failed to get the correct number of bytes read, ignore")
					continue
				}

				// handle resizing
				if msgType == websocket.BinaryMessage {
					if dataBuffer[0] == 1 {
						ttySize := &TTYSize{}
						resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
						if err := json.Unmarshal(resizeMessage, ttySize); err != nil {
							log.Warnf("failed to unmarshal received resize message '%s': %s", string(resizeMessage), err)
							continue
						}
						log.Infof("resizing tty to use %v rows and %v columns...", ttySize.Rows, ttySize.Cols)
						if err := pty.Setsize(tty, &pty.Winsize{
							Rows: ttySize.Rows,
							Cols: ttySize.Cols,
						}); err != nil {
							log.Warnf("failed to resize tty, error: %s", err)
						}
						continue
					}
				}

				// write to tty
				bytesWritten, err := tty.Write(dataBuffer)
				if err != nil {
					log.Warn(fmt.Sprintf("failed to write %v bytes to tty: %s", len(dataBuffer), err))
					continue
				}
				log.Tracef("%v bytes written to tty...", bytesWritten)
			}
		}()

		wg.Wait()
		log.Info("closing connection...")
		connectionClosed = true
	}
}
