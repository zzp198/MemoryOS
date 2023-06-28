package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// //go:embed frontend/css/* frontend/js/*
//
//go:embed frontend/*
var AssetFS embed.FS

//https://github.com/XTLS/Xray-core/releases/download/v1.8.1/Xray-linux-64.zip

func main() {
	var router = gin.Default().Delims("[[", "]]")

	router.LoadHTMLGlob("frontend/*.gohtml")
	frontend, _ := fs.Sub(AssetFS, "frontend")
	router.StaticFS("frontend", http.FS(frontend))

	router.NoRoute(func(c *gin.Context) { //404
		c.JSON(http.StatusNotFound, nil)
	})
	router.NoMethod(func(c *gin.Context) { //405
		c.JSON(http.StatusMethodNotAllowed, nil)
	})

	router.GET("/", home_view)
	router.GET("/v2ray", func(c *gin.Context) { c.HTML(http.StatusOK, "v2ray.gohtml", gin.H{}) })

	Server := http.Server{Addr: "127.0.0.1:80", Handler: router}
	_ = Server.ListenAndServe()
}

func home_view(c *gin.Context) {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	name := filepath.Dir(exe)
	fmt.Println(name)
}
