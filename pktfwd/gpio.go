package pktfwd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/stianeikeland/go-rpio"
)

const gpioTimeMargin = 100 * time.Millisecond

// ResetPin resets the specified pin
func ResetPin(pinNumber int) error {
	err := rpio.Open()
	if err != nil {
		return errors.Wrap(err, "couldn't get GPIO access")
	}

	pin := rpio.Pin(uint8(pinNumber))
	pin.Output()
	time.Sleep(gpioTimeMargin)
	pin.Low()
	time.Sleep(gpioTimeMargin)
	pin.High()
	time.Sleep(gpioTimeMargin)
	pin.Low()
	time.Sleep(gpioTimeMargin)

	return errors.Wrap(rpio.Close(), "couldn't close GPIO access")
}
