package main

import (
	"github.com/gin-gonic/gin"
	_ "image/jpeg"
	"io/ioutil"
	"os"

	bda "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/bda/v20200324"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

var Client *bda.Client
var test string

func main() {

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

	e := gin.Default()
	e.Routes()

	defer e.Run()

	e.LoadHTMLGlob("templates/*")

	e.GET("/index", RetHomePage)
	e.POST("/post", RetImage)

}
