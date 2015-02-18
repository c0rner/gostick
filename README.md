# Gostick
Gostick is a standalone production ready library for communicating with Tellstick home automation devices from Telldus Technologies and
requires no external software except for [libusb](http://libusb.info/).

The API provided by gostick uses the [Tellstick protocol](http://developer.telldus.com/doxygen/TellStick.html) for sending encoded messages.
Received messages are relayed back raw from the device and need to be decoded accordingly.

## API Documentation
Browse the online reference documentation:

[![GoDoc](https://godoc.org/github.com/c0rner/gostick?status.svg)](https://godoc.org/github.com/c0rner/gostick)
[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/c0rner/gostick).

### Configure udev on Linux
To avoid having to run applications as user root to get device access you must install a new udev rule in */etc/udev/rules.d*. 
Create a new file and name it *99-tellstick.rules* adding the text below. This will give full read/write access
to all members of the group *users*. Preferably you would create a new group for the application and update the rule to match.
```
# Telldus Tellstick
SUBSYSTEMS=="usb", ATTRS{manufacturer}=="Telldus", ATTRS{idVendor}=="1781", GROUP="users", MODE="0660"
```
With the new rule in place you must reload and trigger it to have the change take effect. As root (via sudo) run the commands below,
or simply reboot the operating system.
```
udevadm control --reload
udevadm trigger
```
