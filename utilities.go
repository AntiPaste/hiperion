package main

import (
	"log"
	"strconv"
)

func stringToInt(input string) int {
	result, err := strconv.Atoi(input)
	if err != nil {
		log.Fatal("conversion:", err)
	}

	return result
}
