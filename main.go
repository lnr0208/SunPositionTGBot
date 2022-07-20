package main

import (
	"fmt"
	"isayevapps/sunposition/engine"
	"os"
)

func main() {
	coordinates := `55°45'0"N 37°37'0"E`
	date := "20.07.2022"
	time := "17:52"
	gmt := 4.0
	sunPosition, err := engine.GetSunPosition(coordinates, date, time, gmt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(sunPosition)
}
