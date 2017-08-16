package main

import (
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"time"
)

// 定义应用版本
// go build -ldflags "-X main.VersionName=`cat VERSION`" gin/gin_demo.go

var VersionName = "No Version Provided"

func main() {
	fmt.Println("App Version is:", VersionName)

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, time.Now().String(), "Usage: /user/name or /user/name/action\n App Version is: %s", VersionName)
	})

	// This handler will match /user/john but will not match neither /user/ or /user
	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK,"Hello %s\n App Version is: %s", name, VersionName)
	})

	// However, this one will match /user/john/ and also /user/john/send
	// If no other routers match /user/john, it will redirect to /user/john/
	router.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		c.String(http.StatusOK, "%s\n App Version is: %s", message, VersionName)
	})

	router.Run(":8081")
}
