package main

import (
	"fmt"
	"github.com/Nickolas990/otus_hw/hw02_unpack_string/hw02unpackstring"
)

func main() {
	input := `qwe\4\5`
	result, err := hw02unpackstring.Unpack(input)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
