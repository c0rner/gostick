package gostick

//#include <libusb.h>
import "C"

// LibUSBError wraps a libusberror
type LibUSBError struct {
	Code int
	desc string
}

func newLibUSBError(id C.int) error {
	if id == 0 {
		return nil
	}
	err := LibUSBError{
		Code: int(id),
	}
	var errstr *C.char
	errstr = C.libusb_error_name(id)
	err.desc = C.GoString(errstr)
	return err
}

// Error returns a descriptive string of the error
func (e LibUSBError) Error() string {
	return e.desc
}
