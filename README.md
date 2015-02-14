# Gostick
Gostick is a standalone library for communicating with Tellstick home automation devices from Telldus Technologies and requires no
external software except for [libusb](http://libusb.info/).

The API provided by gostick uses the [Tellstick protocol](http://developer.telldus.com/doxygen/TellStick.html) for sending encoded messages.
Received messages are relayed back raw from the device and need to be decoded accordingly.

## API Documentation
Browse the online reference documentation:

[![GoDoc](https://godoc.org/github.com/c0rner/gostick?status.svg)](https://godoc.org/github.com/c0rner/gostick)
[![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/c0rner/gostick).

## Status
Functionally the library is complete and fully working. In the future some type of convenience function for sending messages might be added.
