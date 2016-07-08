package templatehelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"time"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap template.FuncMap = template.FuncMap{
	"DistanceOfTimeInWords":        distanceOfTimeInWords,
	"DistanceOfTimeInWordsFromNow": distanceOfTimeInWordsFromNow,
	"FormatTime":                   formatTime,
	"InKM":                         inKM,
	"MarshalJSON":                  marshalJSON,
	"MonthName":                    monthName,
	"NumberWithDelimiter":          numberWithDelimiter,
	"Pace":                         pace,
	"RoundToString":                roundToString,
}

func distanceOfTimeInWords(to, from time.Time) string {
	fmt.Printf("to:   %+v\n", to)
	fmt.Printf("from: %+v\n", from)
	d := from.Sub(to)

	min := int(round(d.Minutes()))
	fmt.Printf("resulting duration: %+v\n", d)
	fmt.Printf("resulting minutes: %+v\n", min)

	if min == 0 {
		return "less than 1 minute"
	} else if min >= 1 && min <= 44 {
		return fmt.Sprintf("%d minutes", min)
	} else if min >= 45 && min <= 89 {
		return "about 1 hour"
	} else if min >= 90 && min <= 1439 {
		return fmt.Sprintf("about %d hours", int(round(d.Hours())))
	} else if min >= 1440 && min <= 2519 {
		return "about 1 day"
	} else if min >= 2520 && min <= 43199 {
		return fmt.Sprintf("%d days", int(round(d.Hours()/24.0)))
	} else if min >= 43200 && min <= 86399 {
		return "about 1 month"
	} else if min >= 86400 && min <= 525599 {
		return fmt.Sprintf("%d months", int(round(d.Hours()/24.0/30.0)))
	}

	return ""
}

func distanceOfTimeInWordsFromNow(to time.Time) string {
	return distanceOfTimeInWords(to, time.Now())
}

func formatTime(t *time.Time) string {
	return t.Format("January 2, 2006")
}

func inKM(m float64) float64 {
	return m / 1000.0
}

// Note that I thought I needed this to encode Javascript data in HTML
// templates, but I don't actually appear to need it so we can probably remove
// it.
func marshalJSON(o interface{}) string {
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func monthName(t *time.Time) string {
	return t.Format("January")
}

// Changes a number to a string and uses a separator for groups of three
// digits. For example, 1000 changes to "1,000".
func numberWithDelimiter(sep rune, n int) string {
	s := strconv.Itoa(n)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {
		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

// pace calculates the pace of a run in time per kilometer. This comes out as a
// "clock" time like 4:52 which translates to "4 minutes and 52 seconds" per
// kilometer.
func pace(distance float64, duration time.Duration) string {
	speed := float64(duration.Seconds()) / inKM(distance)
	min := int64(speed / 60.0)
	sec := int64(speed) % 60
	return fmt.Sprintf("%v:%02d", min, sec)
}

// There is no "round" function built into Go :/
func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundToString(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
