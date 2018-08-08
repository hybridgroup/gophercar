![Gophercon 2018](https://github.com/hybridgroup/gophercar/blob/master/images/gophercon2018.gif?raw=true)

# Gophercar

Like Donkeycar ([http://www.donkeycar.com/](http://www.donkeycar.com/)), but written in Go. The idea is to make Gophercar able to run on any of the supported Donkeycar cars/hardware without any modification.

## How it will work

![Arch](https://github.com/hybridgroup/gophercar/blob/master/images/arch.png?raw=true)

## Car and Controller Hardware

- Exceed Short Course Truck (https://www.amazon.com/Exceed-Racing-Desert-Course-2-4ghz/dp/9269802094)
- Donkeycar chassis kit(https://squareup.com/store/donkeycar/item/desert-monster-short-course-truck-or-blaze-partial-kit)
- Raspberry Pi 3 Model B+
- Raspberry Pi wide-angle camera (included in Donkeycar kit)
- PCA9685 I2C servo driver (included in Donkeycar kit)
- SSD1306 I2C OLED display (optional)
- MPU6050 I2C Accelerometer/Gyroscope (optional)

## Car OS Software

The following will need to be installed on a bootable SD card for the Raspi:

- Raspbian Stretch OS
- OpenCV 3.4.2
- Movidius NCS SDK (optional)

## Expected workflow

- Install Go 1.10+
- Install Gobot and GoCV
- Install the Gophercar package on your development machine. We will need a Docker container to make cross-compiling for Raspian easier (due to OpenCV/GoCV)
- Compile the code to run on your car
- Copy the compiled executable to your car's controller using scp
- Execute the car code on the car controller
- Drive!

## Cars

The `cars` directory will contain various car controller programs. Choose one to compile and put on your car controller.
