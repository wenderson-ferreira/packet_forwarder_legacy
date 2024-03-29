// +build dummy

package wrapper

import "math/rand"

const (
	StatusCRCOK  = uint8(0)
	StatusCRCBAD = uint8(1)
	StatusNOCRC  = uint8(2)

	ModulationLoRa = uint8(0)
	ModulationFSK  = uint8(1)

	NbMaxPackets = 8
)

// Randomly return 1 empty packet, once every 5000 times (since there's one query per 5 milliseconds)

func Receive() ([]Packet, error) {
	packets := make([]Packet, 0)
	if rand.Float64() <= 0.0002 {
		dummyPacket := Packet{
			Payload: make([]byte, 0),
		}
		packets = append(packets, dummyPacket)
	}
	return packets, nil
}

var datarateString = map[uint32]string{
	uint32(0): "SF7",
	uint32(1): "SF8",
	uint32(2): "SF9",
	uint32(3): "SF10",
	uint32(4): "SF11",
	uint32(5): "SF12",
}

var bandwidthString = map[uint8]string{
	uint8(0): "BW125",
	uint8(1): "BW250",
	uint8(2): "BW500",
}

var coderateString = map[uint8]string{
	uint8(4): "4/5",
	uint8(1): "4/6",
	uint8(2): "4/7",
	uint8(3): "4/8",
	0:        "OFF",
}
