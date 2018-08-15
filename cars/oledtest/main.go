// this does not really do anything yet except connect to all of the various devices
package main

import (
	"time"

	"github.com/fogleman/gg"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	r    *raspi.Adaptor
	oled *i2c.SSD1306Driver

	ctx *gg.Context
)

func main() {
	r = raspi.NewAdaptor()
	oled = i2c.NewSSD1306Driver(r)

	ctx = gg.NewContext(oled.Buffer.Width, oled.Buffer.Height)

	work := func() {
		gobot.Every(1*time.Second, func() {
			handleOLED()
		})
	}

	robot := gobot.NewRobot("gophercar",
		[]gobot.Connection{r},
		[]gobot.Device{oled},
		work,
	)

	robot.Start()
}

func handleOLED() {
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	ctx.SetRGB(1, 1, 1)
	ctx.DrawStringAnchored(time.Now().Format("15:04:05"), 0, 0, 0, 1)
	oled.ShowImage(ctx.Image())
}
