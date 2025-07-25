package emv

import (
	"fmt"

	"github.com/moov-io/bertlv"
)

func ShowBerTLV(response []byte) {
	fmt.Println("Pretty printing BerTLV data...")
	// Remove the status word from the response
	data, err := bertlv.Decode(response[:len(response)-2])
	if err != nil {
		fmt.Println("Failed to decode response:", err)
		return
	}
	bertlv.PrettyPrint(data)
}
