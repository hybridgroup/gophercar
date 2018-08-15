// drive with your keyboard!
//
// controls:
// 	up arrow - forward
// 	down arrow - backward
//	right arrow - turn right
//	left arrow - turn left
//
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/fogleman/gg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/keyboard"
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
)

func main() {
	r = raspi.NewAdaptor()
	pca9685 = i2c.NewPCA9685Driver(r)
	oled = i2c.NewSSD1306Driver(r)
	mpu6050 = i2c.NewMPU6050Driver(r)
	keys := keyboard.NewDriver()

	ctx = gg.NewContext(oled.Buffer.Width, oled.Buffer.Height)

	work := func() {
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

		keys.On(keyboard.Key, func(data interface{}) {
			key := data.(keyboard.KeyEvent)

			switch key.Key {
			case keyboard.ArrowUp:
				setThrottle(throttlePower)

				gobot.After(1*time.Second, func() {
					setThrottle(0)
				})
			case keyboard.ArrowDown:
				setThrottle(-1 * throttlePower)

				gobot.After(1*time.Second, func() {
					setThrottle(0)
				})
			case keyboard.ArrowRight:
				if steering < 1.0 {
					steering = round(steering+0.1, 0.05)
				}

				setSteering(steering)
			case keyboard.ArrowLeft:
				if round(steering, 0.05) > -1.0 {
					steering = round(steering-0.1, 0.05)
				}

				setSteering(steering)
			}
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r},
		[]gobot.Device{pca9685, oled, mpu6050, keys},
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

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
