![Gophercon 2018](https://github.com/hybridgroup/gophercar/blob/master/images/gophercon2018.gif?raw=true)

# Gophercar

Like Donkeycar ([http://www.donkeycar.com/](http://www.donkeycar.com/)), but written in Go.

## How it will work

![Arch](https://github.com/hybridgroup/gophercar/blob/master/images/arch.png?raw=true)

## Car Hardware

- Exceed Short Course Truck
- Raspberry Pi 3 Model B+
- Raspberry Pi wide-angle camera
- PCA9685 I2C servo driver
- SSD1306 I2C OLED display
- MPU6050 I2C Accelerometer/Gyroscope

## Car OS Software

The following will need to be installed on a bootable SD card for the Raspi:

- Raspbian Stretch OS
- OpenCV 3.4.2
- Movidius NCS SDK

## Expected workflow

- Install the gophercar package on your development machine. We probably want a Docker container to make cross-compiling for Raspian easier (due to OpenCV/GoCV)
- Compile the code to run on your car
- Copy the compiled executable to your car's controller using scp
- Execute the car code on the car controller
- Drive!

## Cars

The `cars` directory will contain various car controller programs. Choose one to compile and put on your car.
