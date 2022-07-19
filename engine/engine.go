package engine

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

/* type Date struct {
	day,
	month,
	year int
}

type Time struct {
	hours,
	minutes int
}

type SunPosition struct {
	Azimuth, Altitude float64
} */

type Coordinates struct {
	latitude, longitude float64
}

/* func getSunPosition(coordinates, date, time string, gmt float64) (SunPosition, error){

} */

/* func GetDate(d string) (Date, error) {

}
*/
func GetCoordinates(s string) (Coordinates, error) {
	c := strings.Split(s, " ")
	lat := c[0]
	lon := c[1]
	lat = strings.TrimRight(lat, ",")

	var latitude, longitude float64

	if matched, _ := regexp.MatchString(`^-{0,1}\d{0,1}\d\.\d*$`, lat); matched {
		latitude, _ = strconv.ParseFloat(lat, 64)
	} else if matched, _ := regexp.MatchString(`^-{0,1}\d{0,1}\d°\d{0,1}\d'\d{0,1}\d.{0,1}\d*"[NS]$`, lat); matched {
		deg_str := ""
		minute_str := ""
		second_str := ""
		for _, v := range lat {
			if v == '°' {
				break
			}
			deg_str += string(v)
		}

		for i := len(deg_str) + 2; lat[i] != '\''; i++ {
			minute_str += string(lat[i])
		}

		for i := len(deg_str+minute_str) + 3; lat[i] != '"'; i++ {
			second_str += string(lat[i])
		}

		degs, _ := strconv.Atoi(deg_str)
		if degs > 90 || degs < -90 {
			return Coordinates{}, errors.New("wrong latitude")
		}

		minutes, _ := strconv.Atoi(minute_str)
		if minutes < 0 || minutes > 59 {
			return Coordinates{}, errors.New("wrong latitude")
		}

		seconds, _ := strconv.ParseFloat(second_str, 64)
		if seconds < 0 || seconds > 59.9999999 {
			return Coordinates{}, errors.New("wrong latitude")
		}

		latitude = float64(degs) + float64(minutes)/60.0 + seconds/3600
		if lat[len(lat)-1] == 'S' {
			latitude = -latitude
		}
	} else {
		return Coordinates{}, errors.New("wrong latitude")
	}

	if matched, _ := regexp.MatchString(`^-{0,1}\d{0,1}\d{0,1}\d\.\d*$`, lon); matched {
		longitude, _ = strconv.ParseFloat(lon, 64)
	} else if matched, _ := regexp.MatchString(`^-{0,1}\d{0,1}\d{0,1}\d°\d{0,1}\d'\d{0,1}\d.{0,1}\d*"[EW]$`, lon); matched {
		deg_str := ""
		minute_str := ""
		second_str := ""
		for _, v := range lon {
			if v == '°' {
				break
			}
			deg_str += string(v)
		}

		for i := len(deg_str) + 2; lon[i] != '\''; i++ {
			minute_str += string(lon[i])
		}

		for i := len(deg_str+minute_str) + 3; lon[i] != '"'; i++ {
			second_str += string(lon[i])
		}

		degs, _ := strconv.Atoi(deg_str)
		if degs > 180 || degs < -180 {
			return Coordinates{}, errors.New("wrong longitude")
		}

		minutes, _ := strconv.Atoi(minute_str)
		if minutes < 0 || minutes > 59 {
			return Coordinates{}, errors.New("wrong longitude")
		}

		seconds, _ := strconv.ParseFloat(second_str, 64)
		if seconds < 0 || seconds > 59.9999999 {
			return Coordinates{}, errors.New("wrong longitude")
		}

		longitude = float64(degs) + float64(minutes)/60.0 + seconds/3600
		if lon[len(lon)-1] == 'W' {
			longitude = -longitude
		}
	} else {
		return Coordinates{}, errors.New("wrong longitude")
	}

	return Coordinates{latitude, longitude}, nil
}
