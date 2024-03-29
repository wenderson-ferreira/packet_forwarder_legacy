// +build dummy

package wrapper

import "github.com/TheThingsNetwork/go-utils/log"

func LoRaGPSEnable(TTYPath string) error {
	return nil
}

func GetGPSCoordinates() (GPSCoordinates, error) {
	return GPSCoordinates{}, nil
}

func UpdateGPSData(ctx log.Interface) error {
	return nil
}
