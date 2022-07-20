package engine

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Date struct {
	Day,
	Month,
	Year int
}

type Time struct {
	Hours,
	Minutes int
}

type SunPosition struct {
	Azimuth, Altitude float64
}

type Coordinates struct {
	Latitude, Longitude float64
}

func GetSunPosition(_coordinates, _date, _time string, gmt float64) (SunPosition, error) {
	date, err := GetDate(_date)
	if err != nil {
		return SunPosition{}, err
	}
	time, err := GetTime(_time)
	if err != nil {
		return SunPosition{}, err
	}
	coordinates, err := GetCoordinates(_coordinates)
	if err != nil {
		return SunPosition{}, err
	}
	day := date.Day
	month := date.Month
	year := date.Year
	hour := time.Hours
	minute := time.Minutes
	LAT := coordinates.Latitude
	LON := coordinates.Longitude

	ut := float64(hour) + float64(minute)/60.0 - gmt
	d := float64(367*year - (7*(year+((month+9)/12)))/4 + (275*month)/9 + day - 730530)     //преобразование даты в нужное числовое значение
	w := rev(282.9404 + 4.70935e-5*d)                                                       //долгота перигелия
	e := 0.016709 - 1.151e-9*d                                                              //эксцентриситет
	M := rev(356.0470 + 0.9856002585*d)                                                     //средняя аномалия
	oblecl := rev(23.4393 - 3.563e-7*d)                                                     //наклон эклиптики
	L := rev(w + M)                                                                         //средняя долгота Солнца
	E := M + (180.0/float64(math.Pi))*e*math.Sin(toRadians(M))*(1+e*math.Cos(toRadians(M))) //эксцентрическая аномалия
	x := math.Cos(toRadians(E)) - e
	y := math.Sin(toRadians(E)) * math.Sqrt(1-e*e)
	r := math.Sqrt(x*x + y*y)             //расстояние
	v := rev(toDegrees(math.Atan2(y, x))) //истинная аномалия
	lon := rev(v + w)
	x = r * math.Cos(toRadians(lon))
	y = r * math.Sin(toRadians(lon))
	xequat := x
	yequat := y * math.Cos(toRadians(oblecl))
	zequat := y * math.Sin(toRadians(oblecl))
	//r = math.Sqrt(xequat*xequat + yequat*yequat + zequat*zequat)
	RA := math.Atan2(yequat, xequat)
	Decl := math.Atan2(zequat, math.Sqrt(xequat*xequat+yequat*yequat))

	////////////////////Звездное время.Высота и азимут/////////////////////
	GMST0 := L/15 + 12               //звездное время на гринвичском меридиане в 00:00 прямо сейчас
	SIDTIME := GMST0 + ut + LON/15   //местное звездное время
	HA := SIDTIME*15 - toDegrees(RA) //часовой угол
	x = math.Cos(toRadians(HA)) * math.Cos(Decl)
	y = math.Sin(toRadians(HA)) * math.Cos(Decl)
	z := math.Sin(Decl)
	xhor := x*math.Sin(toRadians(LAT)) - z*math.Cos(toRadians(LAT))
	yhor := y
	zhor := x*math.Cos(toRadians(LAT)) + z*math.Sin(toRadians(LAT))
	azimuth := toDegrees(math.Atan2(yhor, xhor)) + 180
	altitude := toDegrees(math.Asin(zhor))
	return SunPosition{azimuth, altitude}, nil
}

func GetDate(_date string) (Date, error) {
	err := errors.New("wrong date")
	matched, _ := regexp.MatchString(`^\d{2}.\d{2}.\d{4}$`, _date)
	if !matched {
		return Date{}, err
	}
	day, _ := strconv.Atoi(_date[:2])
	month, _ := strconv.Atoi(_date[3:5])
	year, _ := strconv.Atoi(_date[6:])
	if day == 0 || day > 31 {
		return Date{}, err
	}
	if month > 12 || month == 0 {
		return Date{}, err
	}
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		if day > 31 {
			return Date{}, err
		}
	case 2:
		if year%4 == 0 {
			if day > 29 {
				return Date{}, err
			}
		} else if day > 28 {
			return Date{}, err
		}
	case 4, 6, 9, 11:
		if day > 30 {
			return Date{}, err
		}
	}
	return Date{day, month, year}, nil
}

func GetTime(_time string) (Time, error) {
	err := errors.New("wrong time")
	matched, _ := regexp.MatchString(`^\d{2}:\d{2}$`, _time)
	if !matched {
		return Time{}, err
	}
	hours, _ := strconv.Atoi(_time[:2])
	minutes, _ := strconv.Atoi(_time[3:])
	if hours > 23 || hours < 0 {
		return Time{}, err
	}
	if minutes > 59 || minutes < 0 {
		return Time{}, err
	}
	time := Time{hours, minutes}
	return time, nil
}

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

func rev(x float64) float64 {
	return x - math.Floor(x/360.0)*360.0
}

func toRadians(deg float64) float64 {
	return deg * float64(math.Pi) / 180.0
}

func toDegrees(rad float64) float64 {
	return rad * 180.0 / float64(math.Pi)
}
