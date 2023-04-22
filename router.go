package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"log"
	"net/http"
)

//go:embed frontend/jss/*
var Frontend embed.FS

func main() {
	var router = gin.Default()

	router.LoadHTMLGlob("frontend/*.html")
	frontend, _ := fs.Sub(Frontend, "frontend")
	router.StaticFS("frontend", http.FS(frontend))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ping": "pong",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	log.Fatal(router.Run(":80"))
}
