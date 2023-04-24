package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"
)

type SSHDialInfo struct {
	gorm.Model

	Hostname string
	Username string
	Password string
}

//go:embed frontend/css/* frontend/js/*
var AssetFS embed.FS

func main() {
	orm, err := gorm.Open(sqlite.Open("test.db"), nil)
	if err != nil {
		panic("failed to connect database")
	}
	defer orm.Debug()

	var router = gin.Default()

	router.LoadHTMLGlob("frontend/*.html")
	frontend, _ := fs.Sub(AssetFS, "frontend")
	router.StaticFS("frontend", http.FS(frontend))

	router.NoRoute(func(c *gin.Context) { //404
		c.JSON(http.StatusNotFound, nil)
	})
	router.NoMethod(func(c *gin.Context) { //405
		c.JSON(http.StatusMethodNotAllowed, nil)
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	router.GET("/terminal", func(c *gin.Context) {
		c.HTML(http.StatusOK, "terminal.html", gin.H{})
	})

	var ug = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	router.GET("/api/terminal", func(c *gin.Context) {

		hostname := c.Query("hostname")
		username := c.Query("username")
		password := c.Query("password")

		// 连接ssh
		sshClient, err := ssh.Dial("tcp", hostname, &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{ssh.Password(password)},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
		if err != nil {
			log.Fatal(err.Error())
		}
		defer sshClient.Close()

		// 新建会话
		session, err := sshClient.NewSession()
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}

		// 定义pty请求消息体,请求标准输出的终端
		err = session.RequestPty("xterm", 24, 100, modes)
		if err != nil {
			log.Fatal(err.Error())
		}

		conn, err := ug.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		stdin, _ := session.StdinPipe()
		stdout, _ := session.StdoutPipe()

		// session开始start处理shell
		err = session.Shell()
		if err != nil {
			log.Fatal(err.Error())
		}

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Fatal(err.Error())
					return
				}

				_, err = stdin.Write(msg)
				if err != nil && err.Error() != io.EOF.Error() {
					log.Fatal(err.Error())
					return
				}
			}
		}()

		br := bufio.NewReader(stdout)
		buf := []byte{}
		t := time.NewTimer(time.Microsecond * 100)
		defer t.Stop()

		r := make(chan rune)

		// 接收ssh并转发给ws
		go func() {
			for {
				x, size, err := br.ReadRune()
				if err != nil {
					log.Fatal(err.Error())
					return
				}
				if size > 0 {
					r <- x
				}
			}
		}()

		for {
			select {
			case d := <-r:
				if d != utf8.RuneError {
					p := make([]byte, utf8.RuneLen(d))
					utf8.EncodeRune(p, d)
					buf = append(buf, p...)
				} else {
					buf = append(buf, []byte("@")...)
				}
			case <-t.C:
				if len(buf) != 0 {
					fmt.Println(buf)
					err := conn.WriteMessage(websocket.TextMessage, buf)
					if err != nil {
						return
					}
					buf = []byte{}
				}
				t.Reset(time.Microsecond * 100)
			}
		}
	})

	Server := http.Server{Addr: "127.0.0.1:80", Handler: router}

	go func() {
		if err := Server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	exit := make(chan os.Signal)
	// kill 	SIGTERM
	// kill -2 	SIGINT
	// kill -9 	SIGKILL but can't be caught
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit
	log.Println("Shutting down server...")

	c, f := context.WithTimeout(context.Background(), 5*time.Second)
	defer f()

	if err := Server.Shutdown(c); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
