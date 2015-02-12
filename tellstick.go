package gostick

//#cgo pkg-config: libusb-1.0
//#include <libusb.h>
import "C"
import (
	"bytes"
	"errors"
)

const (
	// Telldus USB Vendor ID
	TelldusVID = 0x1781
	// Tellstick Classic Product ID
	ClassicPID = 0x0c30
	// Tellstick Duo Product ID
	DuoPID = 0x0c31
)

// Errors
var (
	ErrDeviceUnavailable = errors.New("USB device unavailable")
	ErrNoDevice          = errors.New("no supported device found")
)

var (
	// Supported models
	Models = []int{ClassicPID, DuoPID}
)

// Tellstick represents a Tellstick device
type Tellstick struct {
	usb *USBContext
	hdl *USBHandle

	// Tellstick model
	Model int
	// Tellstick serial number
	Serial string // Device serial number

	// Read buffer
	readBuf Buffer

	// USB Read/Write timeouts
	timeRead  int
	timeWrite int
	// USB interface
	iface int
	index int
	// USB Endpoint In
	epIn int
	// USB Endpoint Out
	epOut int
	// USB packet transfer size
	maxPktSize int
}

// New will return a Tellstick session to the first found device.
func New() (*Tellstick, error) {
	var err error

	// Prepare struct with known values
	stick := Tellstick{
		readBuf:    make(Buffer, 0, 1024),
		timeRead:   5000,
		timeWrite:  5000,
		iface:      0,
		index:      1,
		epIn:       0x02,
		epOut:      0x81,
		maxPktSize: 64,
	}

	stick.usb, err = NewUSBContext()
	if err != nil {
		return nil, err
	}

	// Find first connected Telldus devices
	var dev *USBDevice
	err = stick.usb.FindFunc(func(d *USBDevice) bool {
		desc, err := d.DeviceDescriptor()
		if err != nil {
			// TODO, maybe store error?
			return false
		}

		if int(desc.idVendor) != TelldusVID {
			return false
		}
		for _, id := range Models {
			if int(desc.idProduct) == id {
				stick.Model = id
				dev = d
				dev.Reference()
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	if dev == nil {
		return nil, ErrNoDevice
	}

	stick.hdl, err = dev.Open()
	dev.Unreference()
	if err != nil {
		stick.Close()
		return nil, err
	}

	// Getting here means we have a working usb context and usb handle
	// Now lets initialise the Tellstick device stick.
	err = stick.ftdiReset()
	if err != nil {
		stick.Close()
		return nil, err
	}

	err = stick.ftdiSetLatencyTimer(32)
	if err != nil {
		stick.Close()
		return nil, err
	}

	err = stick.ftdiPurgeBuffers()
	if err != nil {
		stick.Close()
		return nil, err
	}

	// Set baud rate depending on device type
	switch stick.Model {
	case ClassicPID:
		// Baudrate 4800 encodes to 0x0271 for this chip
		err = stick.ftdiSetBaudrate(0x0271)
	case DuoPID:
		// Baudrate 9600 encodes to 0x4138 for this chip
		err = stick.ftdiSetBaudrate(0x4138)
	}
	if err != nil {
		stick.Close()
		return nil, err
	}

	return &stick, nil
}

// Close will end the Tellstick session and should always be
// called before terminating the application.
func (t *Tellstick) Close() {
	if t.hdl != nil {
		t.hdl.Close()
		t.hdl = nil
	}
	if t.usb != nil {
		t.usb.Exit()
		t.usb = nil
	}
}

// Poll will read data from Tellstick device and return a
// array of string messages. String array may be empty
// and does not indicate a failure.
func (t *Tellstick) Poll() ([]string, error) {
	buf := t.readBuf.New()
	got, err := t.hdl.BulkTransfer(t.epOut, buf, t.timeWrite)
	if got <= 2 {
		return nil, err
	}

	// FTDI data will have 2 modem status bytes in every
	// transfer packet (64 or 512 bytes depending on chip).
	packets := got / t.maxPktSize
	final := got % t.maxPktSize
	for packet := 0; packet < packets; packet++ {
		copy(buf[packet*(t.maxPktSize-2):], buf[packet*t.maxPktSize+2:(packet+1)*t.maxPktSize])
	}
	if final > 2 {
		copy(buf[packets*(t.maxPktSize-2):], buf[packets*t.maxPktSize+2:packets*t.maxPktSize+final])
	}

	// Adjust the "real" buffer and start extracting strings
	t.readBuf.Grow(got - (packets+1)*2)
	var messages []string
	for {
		idx := bytes.IndexByte(t.readBuf, 0x0a)
		if idx < 0 {
			break
		}
		messages = append(messages, string(t.readBuf[:idx]))
		t.readBuf.Shift(idx + 1)
	}
	return messages, err
}

func (t *Tellstick) ftdiSetLatencyTimer(l int) error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.ControlTransfer(FTDIDeviceOutReqtype, SIOSetLatencyTimerRequest, l, t.index, nil, t.timeWrite)
	return err
}

// ftdiPurgeBuffers clears the buffers on the chip and the internal read buffer.
func (t *Tellstick) ftdiPurgeBuffers() error {
	err := t.ftdiPurgeRXBuffers()
	if err != nil {
		return err
	}
	return t.ftdiPurgeTXBuffers()
}

// ftdiPurgeRXBuffers clears the read buffer on the chip and the internal read buffer.
func (t *Tellstick) ftdiPurgeRXBuffers() error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	// Invalidate data in the readbuffer
	t.readBuf.Trunc(0)

	_, err := t.hdl.ControlTransfer(FTDIDeviceOutReqtype, SIOResetRequest, SIOResetPurgeRX, t.index, nil, t.timeWrite)
	return err
}

// ftdiPurgeTXBuffers clears the write buffer on the chip.
func (t *Tellstick) ftdiPurgeTXBuffers() error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.ControlTransfer(FTDIDeviceOutReqtype, SIOResetRequest, SIOResetPurgeTX, t.index, nil, t.timeWrite)
	return err
}

func (t *Tellstick) ftdiReset() error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	// Invalidate data in the readbuffer
	t.readBuf.Trunc(0)

	_, err := t.hdl.ControlTransfer(FTDIDeviceOutReqtype, SIOResetRequest, SIOResetSIO, t.index, nil, t.timeWrite)
	return err
}

// ftdiSetBaudrate sets the FTDI baudrate. Baudrate must already be encoded for the specific chip.
func (t *Tellstick) ftdiSetBaudrate(baud int) error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.ControlTransfer(FTDIDeviceOutReqtype, SIOSetBaudrateRequest, baud, t.index, nil, t.timeWrite)
	return err
}
