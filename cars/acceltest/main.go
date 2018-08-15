// this does not really do anything yet except connect to all of the various devices
package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	r       *raspi.Adaptor
	mpu6050 *i2c.MPU6050Driver
)

func main() {
	r = raspi.NewAdaptor()
	mpu6050 = i2c.NewMPU6050Driver(r)

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			handleAccel()
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r},
		[]gobot.Device{mpu6050},
		work,
	)

	robot.Start()
}

func handleAccel() {
	mpu6050.GetData()

	fmt.Println("Accelerometer", mpu6050.Accelerometer)
	fmt.Println("Gyroscope", mpu6050.Gyroscope)
	fmt.Println("Temperature", mpu6050.Temperature)
}
