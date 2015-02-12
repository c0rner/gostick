#include <usbhelper.h>

/* next_device - Iteration of libusb_device pointer arrays
 * This is a Go helper function that modifies input 'listptr'.
 */
libusb_device *next_device(libusb_device ***listptr)
{
	if (NULL == listptr || NULL == *listptr || NULL == **listptr) {
		return NULL;
	}
	(*listptr)++;
	return **listptr;
}
