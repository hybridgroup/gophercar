// drive with your ds3 controller
//
// controls:
// 	left stick - throttle
//	right stick - steering
//
package main

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/fogleman/gg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/joystick"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	r       *raspi.Adaptor
	pca9685 *i2c.PCA9685Driver
	oled    *i2c.SSD1306Driver
	mpu6050 *i2c.MPU6050Driver

	ctx *gg.Context

	throttleZero  = 350
	throttlePower = 0.25
	steering      = 0.0

	// joystick
	leftX, leftY, rightX, rightY atomic.Value
)

type pair struct {
	x float64
	y float64
}

func main() {
	r = raspi.NewAdaptor()
	pca9685 = i2c.NewPCA9685Driver(r)
	oled = i2c.NewSSD1306Driver(r)
	mpu6050 = i2c.NewMPU6050Driver(r)

	joystickAdaptor := joystick.NewAdaptor()
	stick := joystick.NewDriver(joystickAdaptor, "dualshock3")

	ctx = gg.NewContext(oled.Buffer.Width, oled.Buffer.Height)

	work := func() {
		leftX.Store(float64(0.0))
		leftY.Store(float64(0.0))
		rightX.Store(float64(0.0))
		rightY.Store(float64(0.0))

		gobot.Every(1*time.Second, func() {
			handleOLED()
		})

		gobot.Every(100*time.Millisecond, func() {
			handleAccel()
		})

		// init the PWM controller
		pca9685.SetPWMFreq(60)

		// init the ESC controller for throttle zero
		pca9685.SetPWM(0, 0, uint16(throttleZero))

		stick.On(joystick.LeftX, func(data interface{}) {
			val := float64(data.(int16))
			leftX.Store(val)
		})

		stick.On(joystick.LeftY, func(data interface{}) {
			val := float64(data.(int16))
			leftY.Store(val)
		})

		stick.On(joystick.RightX, func(data interface{}) {
			val := float64(data.(int16))
			rightX.Store(val)
		})

		stick.On(joystick.RightY, func(data interface{}) {
			val := float64(data.(int16))
			rightY.Store(val)
		})

		gobot.Every(10*time.Millisecond, func() {
			// right stick is steering
			rightStick := getRightStick()

			switch {
			case rightStick.x > 10:
				setSteering(gobot.Rescale(rightStick.x, -32767.0, 32767.0, -1.0, 1.0))
			case rightStick.x < -10:
				setSteering(gobot.Rescale(rightStick.x, -32767.0, 32767.0, -1.0, 1.0))
			default:
				setSteering(0)
			}
		})

		gobot.Every(10*time.Millisecond, func() {
			leftStick := getLeftStick()
			// left stick is throttle

			switch {
			case leftStick.y < -10:
				setThrottle(gobot.Rescale(leftStick.y, -32767.0, 32767.0, -1.0, 1.0))
			case leftStick.y > 10:
				setThrottle(gobot.Rescale(leftStick.y, -32767.0, 32767.0, -1.0, 1.0))
			default:
				setThrottle(0)
			}
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r, joystickAdaptor},
		[]gobot.Device{pca9685, oled, mpu6050, stick},
		work,
	)

	robot.Start()
}

func handleOLED() {
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	ctx.SetRGB(1, 1, 1)
	ctx.DrawStringAnchored(time.Now().Format("15:04:05"), 0, 0, 0, 1)

	ctx.DrawStringAnchored(fmt.Sprint("Steering: ", steering), 0, 32, 0, 1)
	oled.ShowImage(ctx.Image())
}

func handleAccel() {
	mpu6050.GetData()

	// fmt.Println("Accelerometer", mpu6050.Accelerometer)
	// fmt.Println("Gyroscope", mpu6050.Gyroscope)
	// fmt.Println("Temperature", mpu6050.Temperature)
}

func setSteering(steering float64) {
	steeringVal := getSteeringPulse(steering)
	pca9685.SetPWM(1, 0, uint16(steeringVal))
}

func setThrottle(throttle float64) {
	throttleVal := getThrottlePulse(throttle)
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

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
