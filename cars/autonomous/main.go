// What it does:
//
// This is the code from the self-driving car that completed the course at Gophercon 2018.
// Written by @erutherford and friends, but then modified by @deadprogram
//
// This example also streams MJPEG video from the car's camera.
// Once running point your browser to the hostname/port you passed in the
// command line (for example http://localhost:8080) and you should see
// the live video stream.
//
// How to run:
//
// autonomous [camera ID] [host:port]
//
//		go get -u github.com/hybridgroup/mjpeg
//		sudo modprobe bcm2835-v4l2
// 		go run ./cars/autonomous/main.go 0 0.0.0.0:8080
//

package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"image"
	"image/color"

	"github.com/fogleman/gg"
	"github.com/hybridgroup/mjpeg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"gocv.io/x/gocv"
)

var (
	// cv related
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream

	// car related
	r       *raspi.Adaptor
	pca9685 *i2c.PCA9685Driver
	mpu6050 *i2c.MPU6050Driver

	ctx *gg.Context

	// self-driving
	steering, throttle atomic.Value

	throttleZero = 350
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("How to run:\n\tmjpeg-streamer [camera ID] [host:port] [throttle]")
		return
	}

	// parse args
	deviceID := os.Args[1]
	host := os.Args[2]
	t, _ := strconv.ParseFloat(os.Args[3], 64)

	steering.Store(float64(0.0))
	throttle.Store(float64(0.0))

	r = raspi.NewAdaptor()
	pca9685 = i2c.NewPCA9685Driver(r)
	mpu6050 = i2c.NewMPU6050Driver(r)

	work := func() {
		// init the PWM controller
		pca9685.SetPWMFreq(60)

		// init the ESC controller for throttle zero
		pca9685.SetPWM(0, 0, uint16(throttleZero))
		time.Sleep(300 * time.Millisecond)
		throttle.Store(t)

		gobot.Every(100*time.Millisecond, func() {
			handleSteering()
			handleThrottle()
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r},
		[]gobot.Device{pca9685, mpu6050},
		work,
	)

	// open webcam
	webcam, err = gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// create the mjpeg stream
	stream = mjpeg.NewStream()

	// start capturing
	go capture()

	fmt.Println("Capturing. Point your browser to " + host)

	// start http server
	http.Handle("/", stream)
	go func() { log.Fatal(http.ListenAndServe(host, nil)) }()

	robot.Start()
}

// capture video and process it to perform autonomous driving.
func capture() {
	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); ok {
		gocv.IMWrite("/tmp/img.jpg", img)
	}
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		img, rawSteering := processVision(img)
		applySteeringCurve(rawSteering)

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf)
	}
}

func applySteeringCurve(raw float64) {
	steering.Store(float64(raw) * -7)
}

func handleSteering() {
	steeringVal := getSteeringPulse(steering.Load().(float64))
	pca9685.SetPWM(1, 0, uint16(steeringVal))
}

func handleThrottle() {
	throttleVal := getThrottlePulse(throttle.Load().(float64))
	pca9685.SetPWM(0, 0, uint16(throttleVal))
}

// adjusts the steering from -1.0 (hard left) <-> 1.0 (hardright) to the correct
// pwm pulse values.
func getSteeringPulse(val float64) float64 {
	return gobot.Rescale(val, -1, 1, 290, 490)
}

// adjusts the throttle from -1.0 (hard back) <-> 1.0 (hard forward) to the correct
// pwm pulse values.
func getThrottlePulse(val float64) int {
	if val > 0 {
		return int(gobot.Rescale(val, 0, 1, 350, 300))
	}
	return int(gobot.Rescale(val, -1, 0, 490, 350))
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

// processVision processes each frame and returns a Mat with the modified image frame showing the analysis results,
// along with the correct steering direction to keep the car on the track.
func processVision(original gocv.Mat) (gocv.Mat, float64) {
	bwImg := gocv.NewMat()
	defer bwImg.Close()
	blurredImg := gocv.NewMat()
	defer blurredImg.Close()
	thresholdImg := gocv.NewMat()
	defer thresholdImg.Close()
	erodedImg := gocv.NewMat()
	defer erodedImg.Close()
	outputImg := gocv.NewMat()
	defer outputImg.Close()

	dim := original.Size()
	cropHeight := int(float64(dim[0]) * 0.4)
	region := original.Region(image.Rectangle{image.Point{0, cropHeight}, image.Point{dim[1], dim[0]}})

	gocv.CvtColor(region, &bwImg, gocv.ColorBGRToGray)
	gocv.GaussianBlur(bwImg, &blurredImg, image.Point{X: 5, Y: 5}, float64(5), float64(5), gocv.BorderDefault)
	gocv.Threshold(blurredImg, &thresholdImg, float32(100), float32(255), gocv.ThresholdBinary)

	gocv.Erode(thresholdImg, &erodedImg, gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 6, Y: 6}))
	gocv.Dilate(erodedImg, &outputImg, gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 6, Y: 6}))

	contours := gocv.FindContours(outputImg, gocv.RetrievalList, gocv.ChainApproxNone)
	maxArea := float64(0)
	maxContour := 0

	for idx, contour := range contours {
		area := gocv.ContourArea(contour)
		if area > maxArea {
			maxArea = area
			maxContour = idx
		}
	}

	line := gocv.NewMatWithSize(region.Rows(), region.Cols(), gocv.MatTypeCV8U)
	gocv.FillPoly(&line, contours[maxContour:maxContour+1], color.RGBA{R: 255, G: 255, B: 255, A: 255})
	M := gocv.Moments(line, true)

	cx := M["m10"] / M["m00"]

	gocv.DrawContours(&region, contours, maxContour, color.RGBA{R: 255, A: 255}, 3)
	dim = region.Size()
	centerX := dim[1] / 2
	gocv.Line(&region, image.Point{X: centerX, Y: 0}, image.Point{X: centerX, Y: dim[0]}, color.RGBA{B: 255, A: 255}, 1)
	gocv.Circle(&region, image.Point{X: int(cx), Y: dim[0] / 2}, 1, color.RGBA{G: 255, A: 255}, 2)

	steer := cx/float64(centerX) - 0.5

	return region, steer
}
