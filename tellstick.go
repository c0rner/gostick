package gostick

//#cgo pkg-config: libusb-1.0
//#include <libusb.h>
import "C"
import (
	"bytes"
	"errors"
	"io"
)

const (
	// TelldusVID USB Vendor ID
	TelldusVID = 0x1781
	// ClassicPID - Tellstick Classic Product ID
	ClassicPID = 0x0c30
	// DuoPID - Tellstick Duo Product ID
	DuoPID = 0x0c31
)

// Error messages
var (
	ErrDeviceUnavailable = errors.New("USB device unavailable")
	ErrNoDevice          = errors.New("no supported device found")
)

var (
	// Models array contains currently supported Tellstick models
	Models = []int{ClassicPID, DuoPID}
)

// Tellstick represents a Tellstick device
type Tellstick struct {
	usb *usbContext
	hdl *usbHandle

	// Tellstick model
	Model string
	// Tellstick serial number
	Serial string

	// Read buffer
	readBuf buffer

	// USB Product ID
	productID int
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

// New returns a new Tellstick device.
//
// A device will be selected on first found basis irrespective of it already being in use.
func New() (*Tellstick, error) {
	var err error
	var iProduct int
	var iSerial int

	// Prepare struct with known values
	stick := Tellstick{
		readBuf:    make(buffer, 0, 1024),
		timeRead:   5000,
		timeWrite:  5000,
		iface:      0,
		index:      1,
		epIn:       0x02,
		epOut:      0x81,
		maxPktSize: 64,
	}

	stick.usb, err = newUSBContext()
	if err != nil {
		return nil, err
	}

	// Find first connected Telldus devices
	var dev *usbDevice
	err = stick.usb.findFunc(func(d *usbDevice) bool {
		desc, err := d.deviceDescriptor()
		if err != nil {
			// TODO, maybe store error?
			return false
		}

		if int(desc.idVendor) != TelldusVID {
			return false
		}
		for _, id := range Models {
			if int(desc.idProduct) == id {
				// Tellstick device found. Store product ID
				// and increment device reference counter
				iProduct = int(desc.iProduct)
				iSerial = int(desc.iSerialNumber)
				stick.productID = id
				d.reference()
				dev = d
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

	// Open Tellstick device and decrement device reference counter
	// before finally claiming the interface.
	stick.hdl, err = dev.open()
	dev.unreference()
	if err != nil {
		stick.Close()
		return nil, err
	}
	err = stick.hdl.claimInterface(stick.iface)
	if err != nil {
		stick.Close()
		return nil, err
	}

	// Get Tellstick model and serial number
	stick.Model, err = stick.hdl.stringDescriptorASCII(iProduct)
	if err != nil {
		stick.Close()
		return nil, err
	}
	stick.Serial, err = stick.hdl.stringDescriptorASCII(iSerial)
	if err != nil {
		stick.Close()
		return nil, err
	}

	// Getting here means we have a working usb context and usb handle
	// enabling communication with the Tellstick device.
	// Continue initialising the device interface.

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

	// Set baud rate depending on device type. The baud rate needs to
	// be encoded for the specific FTDI chip in the device.
	switch stick.productID {
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
		t.hdl.releaseInterface(t.iface)
		t.hdl.close()
		t.hdl = nil
	}
	if t.usb != nil {
		t.usb.exit()
		t.usb = nil
	}
}

// Poll will read data from Tellstick device and return a
// array of string messages. String array may be empty
// and does not indicate a failure.
func (t *Tellstick) Poll() ([]string, error) {
	var err error
	var got int

	buf := t.readBuf.new()
	if buf != nil && len(buf) == 0 {
		// Something has gone terribly wrong if the read buffer
		// is full. Purge device and read buffers.
		err = t.ftdiPurgeBuffers()
		return nil, err
	}

	got, err = t.hdl.bulkTransfer(t.epOut, buf, t.timeWrite)
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
	t.readBuf.resize(got - (packets+1)*2)
	var messages []string
	for {
		idx := bytes.IndexByte(t.readBuf, 0x0a)
		if idx < 0 {
			break
		}
		messages = append(messages, string(t.readBuf[:idx-1]))
		t.readBuf.shift(idx + 1)
	}
	return messages, err
}

// SendRaw will transmit a raw data stream
//
// Sending malformed data may put the Tellstick device in a unstable state.
func (t *Tellstick) SendRaw(msg io.Reader) error {
	var err error
	buf := make([]byte, 512)
	count, err := msg.Read(buf)
	if count <= 0 || err != nil {
		return err
	}
	_, err = t.hdl.bulkTransfer(t.epIn, buf[:count], t.timeWrite)
	return err
}

func (t *Tellstick) ftdiSetLatencyTimer(l int) error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.controlTransfer(ftdiDeviceOutReqtype, sioSetLatencyTimerRequest, l, t.index, nil, t.timeWrite)
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
	t.readBuf.trunc(0)

	_, err := t.hdl.controlTransfer(ftdiDeviceOutReqtype, sioResetRequest, sioResetPurgeRX, t.index, nil, t.timeWrite)
	return err
}

// ftdiPurgeTXBuffers clears the write buffer on the chip.
func (t *Tellstick) ftdiPurgeTXBuffers() error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.controlTransfer(ftdiDeviceOutReqtype, sioResetRequest, sioResetPurgeTX, t.index, nil, t.timeWrite)
	return err
}

func (t *Tellstick) ftdiReset() error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	// Invalidate data in the readbuffer
	t.readBuf.trunc(0)

	_, err := t.hdl.controlTransfer(ftdiDeviceOutReqtype, sioResetRequest, sioResetSIO, t.index, nil, t.timeWrite)
	return err
}

// ftdiSetBaudrate sets the FTDI baudrate. Baudrate must already be encoded for the specific chip.
func (t *Tellstick) ftdiSetBaudrate(baud int) error {
	if t.hdl == nil {
		return ErrDeviceUnavailable
	}

	_, err := t.hdl.controlTransfer(ftdiDeviceOutReqtype, sioSetBaudrateRequest, baud, t.index, nil, t.timeWrite)
	return err
}
