package gostick

//#cgo pkg-config: libusb-1.0
//#include <libusb.h>
//#include "usbhelper.h"
import "C"
import (
	"unsafe"
)

const (
	// Maximum length of a description string
	USBStringDescMaxLen = (256 / 2) - 2
)

const (
	// FTDI Outgoing Request Type
	FTDIDeviceOutReqtype int = (C.LIBUSB_REQUEST_TYPE_VENDOR | C.LIBUSB_RECIPIENT_DEVICE | C.LIBUSB_ENDPOINT_OUT)
	// FTDI Incoming Request Type
	FTDIDeviceInReqtype int = (C.LIBUSB_REQUEST_TYPE_VENDOR | C.LIBUSB_RECIPIENT_DEVICE | C.LIBUSB_ENDPOINT_IN)
)

// Definitions for flow control
const (
	// Reset the port
	SIOReset, SIOResetRequest int = iota, iota
	_, _
	// Set flow control register
	SIOSetFlowCtrl, SIOSetFlowCtrlRequest
	// Set baud rate
	SIOSetBaudrate, SIOSetBaudrateRequest

	// Set latency timer
	SIOSetLatencyTimerRequest = 9
	// Set bitbang mode
	SIOSetBitmodeRequest = 11
)

const (
	// Reset
	SIOResetSIO int = iota
	// Purge receive buffers
	SIOResetPurgeRX
	// Purge transmit buffers
	SIOResetPurgeTX
)

// USBContext maps directly to a libusb_context struct
type USBContext C.libusb_context

// New returns a new initialized libusb context
func NewUSBContext() (*USBContext, error) {
	var ctx *C.struct_libusb_context
	if ret := C.libusb_init(&ctx); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	var c *USBContext = (*USBContext)(ctx)

	return c, nil
}

// Exit end the usb session
func (c *USBContext) Exit() {
	C.libusb_exit(c.ptr())
}

// FundFunc is used to iterate connected USB devices
func (c *USBContext) FindFunc(match func(d *USBDevice) bool) error {
	var devs **C.libusb_device
	if ret := C.libusb_get_device_list(c.ptr(), &devs); ret < 0 {
		return newLibUSBError(C.int(ret))
	}
	defer C.libusb_free_device_list(devs, 1)

	for usbdev := *devs; usbdev != nil; usbdev = C.next_device(&devs) {
		var dev *USBDevice = (*USBDevice)(usbdev)
		if match(dev) {
			break
		}
	}

	return nil
}

func (c *USBContext) ptr() *C.struct_libusb_context {
	return (*C.struct_libusb_context)(c)
}

// USBDevice maps directly to a libusb_device struct
type USBDevice C.libusb_device

// DeviceDescriptor returns the USB device descriptor
func (d *USBDevice) DeviceDescriptor() (*C.struct_libusb_device_descriptor, error) {
	var desc C.struct_libusb_device_descriptor
	if ret := C.libusb_get_device_descriptor(d.ptr(), &desc); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	return &desc, nil
}

// Open returns a USB device handle after successfully opening the device
func (d *USBDevice) Open() (*USBHandle, error) {
	var hdl *C.libusb_device_handle
	if ret := C.libusb_open(d.ptr(), &hdl); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	var h *USBHandle = (*USBHandle)(hdl)
	return h, nil
}

// Reference increases the device reference count
func (d *USBDevice) Reference() {
	C.libusb_ref_device(d.ptr())
}

// Reference decreases the device reference count
func (d *USBDevice) Unreference() {
	C.libusb_unref_device(d.ptr())
}

func (d *USBDevice) ptr() *C.libusb_device {
	return (*C.libusb_device)(d)
}

// USBHandle maps directly to a libusb_device_handle struct
type USBHandle C.libusb_device_handle

// Close terminates the device session
func (h *USBHandle) Close() {
	C.libusb_close(h.ptr())
}

// StringDescriptorAscii returns a string matching the descriptor string index i
func (h *USBHandle) StringDescriptorAscii(i int) (string, error) {
	buf := make([]byte, USBStringDescMaxLen)
	if ret := C.libusb_get_string_descriptor_ascii(h.ptr(), C.uint8_t(i), (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf))); ret < 0 {
		return "", newLibUSBError(ret)
	}

	return string(buf), nil
}

// BulkTransfer sends/receives data to/from endpoint ep
func (h *USBHandle) BulkTransfer(ep int, data []byte, tout int) (int, error) {
	var err error
	var got C.int
	if ret := C.libusb_bulk_transfer(h.ptr(), C.uchar(ep), (*C.uchar)(unsafe.Pointer(&data[0])),
		C.int(len(data)), &got, C.uint(tout)); ret < 0 {
		err = newLibUSBError(ret)
	}
	return int(got), err
}

// ControlTransfer
func (h *USBHandle) ControlTransfer(typ, req, val, idx int, data []byte, tout int) (int, error) {
	var ret C.int
	var dataPtr *C.uchar

	if data != nil {
		dataPtr = (*C.uchar)(unsafe.Pointer(&data[0]))
	}
	if ret = C.libusb_control_transfer(h.ptr(), C.uint8_t(typ), C.uint8_t(req), C.uint16_t(val), C.uint16_t(idx),
		dataPtr, C.uint16_t(len(data)), C.uint(tout)); ret < 0 {
		return 0, newLibUSBError(ret)
	}
	return int(ret), nil
}

// ReleaseInterface
func (h *USBHandle) ReleaseInterface(i int) error {
	if ret := C.libusb_release_interface(h.ptr(), C.int(i)); ret < 0 {
		return newLibUSBError(ret)
	}
	return nil
}

func (d *USBHandle) ptr() *C.libusb_device_handle {
	return (*C.libusb_device_handle)(d)
}
