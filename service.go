package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	bda "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/bda/v20200324"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"strconv"
)

const threshold = 80
const Athreshold = 250

func RetHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"title": "Main website",
	})
}

func RetImage(c *gin.Context) {

	blur, _ := strconv.Atoi(c.Request.FormValue("blur"))

	f, _, _ := c.Request.FormFile("upload")
	fileBytes, _ := ioutil.ReadAll(f)

	if blur == 0 {
		c.JSON(http.StatusOK, map[string]interface{}{"new_image": fileBytes})
		return
	}

	responseImg := fileBytes
	mat, _ := strconv.Atoi(c.Request.FormValue("matting"))
	if mat == 1 {
		mattingStr := matting(fileBytes)
		responseImg, _ = base64.StdEncoding.DecodeString(mattingStr)
	}

	reader := bytes.NewReader(responseImg)
	srcImg, _, _ := image.Decode(reader)

	retBytes := mosaic(srcImg, blur)

	tracing, _ := strconv.Atoi(c.Request.FormValue("tracing"))
	if tracing == 1 {
		retBytes = trace(retBytes, blur)
	}

	retBytes = trim(retBytes)

	pad, _ := strconv.Atoi(c.Request.FormValue("padding"))
	if pad == 1 {
		retBytes = padding(retBytes)
	}

	c.JSON(http.StatusOK, map[string]interface{}{"new_image": retBytes})

	return
}

func padding(imgBytes []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Write(imgBytes)
	srcImg, _, _ := image.Decode(buf)

	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: srcImg.Bounds().Dx() * 3,
			Y: srcImg.Bounds().Dy() * 3,
		},
	})

	for x := 0; x < srcImg.Bounds().Dx(); x++ {
		for y := 0; y < srcImg.Bounds().Dy(); y++ {
			img.Set(x+srcImg.Bounds().Dx(), y+srcImg.Bounds().Dy(), srcImg.At(x, y))
		}
	}

	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			r, g, b, a := img.At(x, y).RGBA()
			if !(0 == r && 0 == g && 0 == b && a == 0) {
				continue
			}
			img.Set(x, y, color.White)
		}
	}

	buf = new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func trim(imgBytes []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Write(imgBytes)
	srcImg, _, _ := image.Decode(buf)

	leftx, lefty, rightx, righty := 0, 0, 0, 0
	for x := 0; x < srcImg.Bounds().Dx(); x++ {
		for y := 0; y < srcImg.Bounds().Dy(); y++ {
			r, g, b, a := srcImg.At(x, y).RGBA()
			if !(0 == r && 0 == g && 0 == b && a == 0) {
				leftx = x
				break
			}
		}
	}
	for y := 0; y < srcImg.Bounds().Dy(); y++ {
		for x := 0; x < srcImg.Bounds().Dx(); x++ {
			r, g, b, a := srcImg.At(x, y).RGBA()
			if !(0 == r && 0 == g && 0 == b && a == 0) {
				lefty = y
				break
			}
		}
	}

	for x := srcImg.Bounds().Dx() - 1; x > 0; x-- {
		for y := 0; y < srcImg.Bounds().Dy(); y++ {
			r, g, b, a := srcImg.At(x, y).RGBA()
			if !(0 == r && 0 == g && 0 == b && a == 0) {
				rightx = x
				break
			}
		}
	}
	for y := srcImg.Bounds().Dy() - 1; y > 0; y-- {
		for x := 0; x < srcImg.Bounds().Dx(); x++ {
			r, g, b, a := srcImg.At(x, y).RGBA()
			if !(0 == r && 0 == g && 0 == b && a == 0) {
				righty = y
				break
			}
		}
	}

	fmt.Printf("rx:%v, ry:%v, lx:%v, ly:%v\n", rightx, righty, leftx, lefty)

	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: -rightx + leftx + 1,
			Y: -righty + lefty + 1,
		},
	})

	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			c := srcImg.At(rightx+x, righty+y)
			img.Set(x, y, c)
		}
	}

	buf = new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func trace(imgBytes []byte, blur int) []byte {

	buf := new(bytes.Buffer)
	buf.Write(imgBytes)
	srcImg, _, _ := image.Decode(buf)
	img := image.NewRGBA(srcImg.Bounds())

	downSampleX := srcImg.Bounds().Dx() * blur / 1000
	downSampleY := srcImg.Bounds().Dy() * blur / 1000

	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			// 这个像素本身就有颜色
			r, g, b, a := srcImg.At(x, y).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			a >>= 8
			if r > threshold || g > threshold || b > threshold || a > Athreshold {
				img.Set(x, y, color.RGBA{
					R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a),
				})
				continue
			}

			// 颜色过淡，描边
			r, g, b, a = srcImg.At(x+downSampleX, y).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			a >>= 8
			down := r > threshold || g > threshold || b > threshold || a > Athreshold
			r, g, b, a = srcImg.At(x-downSampleX, y).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			a >>= 8
			up := r > threshold || g > threshold || b > threshold || a > Athreshold

			r, g, b, a = srcImg.At(x, y+downSampleY).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			a >>= 8
			left := r > threshold || g > threshold || b > threshold || a > Athreshold

			r, g, b, a = srcImg.At(x, y-downSampleY).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			a >>= 8
			right := r > threshold || g > threshold || b > threshold || a > threshold

			if down || up || left || right {
				img.Set(x, y, color.Black)
			}
		}
	}

	buf = new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()

}

func matting(fileBytes []byte) string {
	request := bda.NewSegmentPortraitPicRequest()
	request.Image = common.StringPtr(base64.StdEncoding.EncodeToString(fileBytes))

	response, err := Client.SegmentPortraitPic(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return ""
	}
	if err != nil {
		panic(err)
	}
	return *response.Response.ResultImage
}

func mosaic(srcImg image.Image, blur int) []byte {
	downSampleX := srcImg.Bounds().Dx() * blur / 1000
	downSampleY := srcImg.Bounds().Dy() * blur / 1000

	m := image.NewRGBA(image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: srcImg.Bounds().Dx(),
			Y: srcImg.Bounds().Dy(),
		},
	})

	for x := 0; x < srcImg.Bounds().Dx()/downSampleX; x++ {
		for y := 0; y < srcImg.Bounds().Dy()/downSampleY; y++ {
			leftx := x * downSampleX
			rightx := (x + 1) * downSampleX
			lefty := y * downSampleY
			righty := (y + 1) * downSampleY

			var r, g, b, a uint32 = 0, 0, 0, 0
			counter := 0
			for xx := leftx; xx < rightx; xx++ {
				for yy := lefty; yy < righty; yy++ {
					counter++
					rr, gg, bb, aa := srcImg.At(xx, yy).RGBA()

					// log.Printf("rr:%+v, gg:%+v, bb:%+v, aa:%+v", rr, gg, bb, aa)

					r += rr >> 8
					g += gg >> 8
					b += bb >> 8
					a += aa >> 8
				}
			}

			for xx := leftx; xx < rightx; xx++ {
				for yy := lefty; yy < righty; yy++ {

					m.Set(xx, yy, color.RGBA{
						R: uint8(int(r) / counter),
						G: uint8(int(g) / counter),
						B: uint8(int(b) / counter),
						A: uint8(int(a) / counter),
					})
				}
			}

		}

	}

	buf := new(bytes.Buffer)
	_ = png.Encode(buf, m)
	return buf.Bytes()
}
