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
	usbStringDescMaxLen = (256 / 2) - 2
)

// FTDI request types
const (
	ftdiDeviceOutReqtype int = (C.LIBUSB_REQUEST_TYPE_VENDOR | C.LIBUSB_RECIPIENT_DEVICE | C.LIBUSB_ENDPOINT_OUT)
	ftdiDeviceInReqtype  int = (C.LIBUSB_REQUEST_TYPE_VENDOR | C.LIBUSB_RECIPIENT_DEVICE | C.LIBUSB_ENDPOINT_IN)
)

// Definitions for flow control
const (
	sioReset, sioResetRequest int = iota, iota
	_, _
	sioSetFlowCtrl, sioSetFlowCtrlRequest
	sioSetBaudrate, sioSetBaudrateRequest

	sioSetLatencyTimerRequest = 9
	sioSetBitmodeRequest      = 11
)

const (
	sioResetSIO int = iota
	sioResetPurgeRX
	sioResetPurgeTX
)

// usbContext maps directly to a libusb_context struct
type usbContext C.libusb_context

// New returns a new initialized libusb context
func newUSBContext() (*usbContext, error) {
	var ctx *C.struct_libusb_context
	if ret := C.libusb_init(&ctx); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	var c *usbContext = (*usbContext)(ctx)

	return c, nil
}

// Exit end the usb session
func (c *usbContext) exit() {
	C.libusb_exit(c.ptr())
}

// FundFunc is used to iterate connected USB devices
func (c *usbContext) findFunc(match func(d *usbDevice) bool) error {
	var devs **C.libusb_device
	if ret := C.libusb_get_device_list(c.ptr(), &devs); ret < 0 {
		return newLibUSBError(C.int(ret))
	}
	defer C.libusb_free_device_list(devs, 1)

	for usbdev := *devs; usbdev != nil; usbdev = C.next_device(&devs) {
		var dev *usbDevice = (*usbDevice)(usbdev)
		if match(dev) {
			break
		}
	}

	return nil
}

func (c *usbContext) ptr() *C.struct_libusb_context {
	return (*C.struct_libusb_context)(c)
}

// usbDevice maps directly to a libusb_device struct
type usbDevice C.libusb_device

// DeviceDescriptor returns the USB device descriptor
func (d *usbDevice) deviceDescriptor() (*C.struct_libusb_device_descriptor, error) {
	var desc C.struct_libusb_device_descriptor
	if ret := C.libusb_get_device_descriptor(d.ptr(), &desc); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	return &desc, nil
}

// Open returns a USB device handle after successfully opening the device
func (d *usbDevice) open() (*usbHandle, error) {
	var hdl *C.libusb_device_handle
	if ret := C.libusb_open(d.ptr(), &hdl); ret < 0 {
		return nil, newLibUSBError(ret)
	}
	var h *usbHandle = (*usbHandle)(hdl)
	return h, nil
}

// Reference increases the device reference count
func (d *usbDevice) reference() {
	C.libusb_ref_device(d.ptr())
}

// Reference decreases the device reference count
func (d *usbDevice) unreference() {
	C.libusb_unref_device(d.ptr())
}

func (d *usbDevice) ptr() *C.libusb_device {
	return (*C.libusb_device)(d)
}

// usbHandle maps directly to a libusb_device_handle struct
type usbHandle C.libusb_device_handle

// Close terminates the device session
func (h *usbHandle) close() {
	C.libusb_close(h.ptr())
}

// StringDescriptorAscii returns a string matching the descriptor string index i
func (h *usbHandle) stringDescriptorAscii(i int) (string, error) {
	buf := make([]byte, usbStringDescMaxLen)
	if ret := C.libusb_get_string_descriptor_ascii(h.ptr(), C.uint8_t(i), (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf))); ret < 0 {
		return "", newLibUSBError(ret)
	}

	return string(buf), nil
}

// BulkTransfer sends/receives data to/from endpoint ep
func (h *usbHandle) bulkTransfer(ep int, data []byte, tout int) (int, error) {
	var err error
	var got C.int
	if ret := C.libusb_bulk_transfer(h.ptr(), C.uchar(ep), (*C.uchar)(unsafe.Pointer(&data[0])),
		C.int(len(data)), &got, C.uint(tout)); ret < 0 {
		err = newLibUSBError(ret)
	}
	return int(got), err
}

// ControlTransfer
func (h *usbHandle) controlTransfer(typ, req, val, idx int, data []byte, tout int) (int, error) {
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
func (h *usbHandle) releaseInterface(i int) error {
	if ret := C.libusb_release_interface(h.ptr(), C.int(i)); ret < 0 {
		return newLibUSBError(ret)
	}
	return nil
}

func (d *usbHandle) ptr() *C.libusb_device_handle {
	return (*C.libusb_device_handle)(d)
}
