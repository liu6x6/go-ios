package main

import (
	"fmt"
)

// hexdump function to dump the contents of a byte slice
func hexdump(data []byte) {
	buffer := make([]byte, 16)
	offset := 0

	for len(data) > 0 {
		// Determine the number of bytes to process in this line
		n := copy(buffer, data)
		data = data[n:]

		// Print the offset
		fmt.Printf("%08x  ", offset)

		// Print the hex values
		for i := 0; i < n; i++ {
			fmt.Printf("%02x ", buffer[i])
		}

		// Print padding for incomplete lines
		for i := n; i < 16; i++ {
			fmt.Print("   ")
		}

		// Print the ASCII values
		fmt.Print(" |")
		for i := 0; i < n; i++ {
			if buffer[i] >= 32 && buffer[i] <= 126 {
				fmt.Printf("%c", buffer[i])
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")

		offset += n
	}
}

// func main() {
//     // Sample byte array to dump
//     data := []byte("Hello, this is a sample byte array for hexdump!")

//     // Call the hexdump function
//     hexdump(data)
// }
