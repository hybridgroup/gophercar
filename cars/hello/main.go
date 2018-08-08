// this does not really do anything yet except connect to all of the various devices
package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	r *raspi.Adaptor
	pca9685 *i2c.PCA9685Driver
	oled *i2c.SSD1306Driver
	mpu6050 *i2c.MPU6050Driver

	stage = false
)

func main() {
	r = raspi.NewAdaptor()
	pca9685 = i2c.NewPCA9685Driver(r)
	oled = i2c.NewSSD1306Driver(r)
	mpu6050 = i2c.NewMPU6050Driver(r)

	// just here as placeholder for the real steering or throttle
	servo := gpio.NewServoDriver(pca9685, "15")

	work := func() {
		gobot.Every(1*time.Second, func() {
			handleOLED()
		})

		gobot.Every(100*time.Millisecond, func() {
			handleAccel()
		})

		pca9685.SetPWMFreq(60)
		i := 10
		direction := 1

		gobot.Every(1*time.Second, func() {
			fmt.Println("Turning", i)
			servo.Move(uint8(i))
			if i > 150 {
				direction = -1
			}
			if i < 10 {
				direction = 1
			}
			i += 10 * direction
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r},
		[]gobot.Device{pca9685, oled, mpu6050, servo},
		work,
	)

	robot.Start()
}

func handleOLED() {
	oled.Clear()
	if stage {
		for x := 0; x < oled.Buffer.Width; x += 5 {
			for y := 0; y < oled.Buffer.Height; y++ {
				oled.Set(x, y, 1)
			}
		}
	}
	stage = !stage
	oled.Display()
}

func handleAccel() {
	mpu6050.GetData()

	fmt.Println("Accelerometer", mpu6050.Accelerometer)
	fmt.Println("Gyroscope", mpu6050.Gyroscope)
	fmt.Println("Temperature", mpu6050.Temperature)
}
