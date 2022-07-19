package main

import (
	"fmt"
	"isayevapps/sunposition/engine"
)

func main() {
	coordinates, err := engine.GetCoordinates(`43°18'02.6"N 76°56'32.1"E`)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(coordinates.Latitude)
		fmt.Println(coordinates.Longitude)
	}
}
