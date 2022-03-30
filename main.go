package main

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"


	_ "image/jpeg"


	bda "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/bda/v20200324"
)

var Client *bda.Client
var test string


func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")


	credential := common.NewCredential(
		"AKIDIQegT0IMAi14sQilY5NPCj13pLcdPvBC",
		"oMfjX6PSHrBF4dx1qPima9SWKsBE1Zsv",
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "bda.tencentcloudapi.com"
	Client, _ = bda.NewClient(credential, "ap-shanghai", cpf)

	f, _ := os.Open("./test.txt")
	t, _ := ioutil.ReadAll(f)
	test = string(t)


	router.GET("/index", RetHomePage)
	router.POST("/post", RetImage)

	router.Run(":" + port)
}
