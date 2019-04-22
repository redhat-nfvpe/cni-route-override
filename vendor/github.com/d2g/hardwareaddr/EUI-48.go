package hardwareaddr

import (
	"math/rand"
	"net"
	"time"
)

/*
 * This function is for Generating a Random EUI-48 (MAC) Address
 * 00:00:00:00:00:00
 */

func GenerateEUI48() (net.HardwareAddr, error) {
	allowedCharacters := [16]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}
	var macAddress string

	for i := 0; i < 6; i++ {

		for j := 0; j < 2; j++ {
			macAddress = macAddress + allowedCharacters[rand.Intn(len(allowedCharacters)-1)]
		}

		if i < 5 {
			macAddress = macAddress + ":"
		}
	}

	return net.ParseMAC(macAddress)
}

func init() {
	rand.Seed(time.Now().Unix())
}
