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
	r       *raspi.Adaptor
	pca9685 *i2c.PCA9685Driver
)

func main() {
	r = raspi.NewAdaptor()
	pca9685 = i2c.NewPCA9685Driver(r)

	// just here as placeholder for the real steering or throttle
	servo := gpio.NewServoDriver(pca9685, "1")

	work := func() {
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
		[]gobot.Device{pca9685, servo},
		work,
	)

	robot.Start()
}
