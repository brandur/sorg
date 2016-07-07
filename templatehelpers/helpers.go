package templatehelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"time"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap template.FuncMap = template.FuncMap{
	"FormatTime":          formatTime,
	"InKM":                inKM,
	"MarshalJSON":         marshalJSON,
	"MonthName":           monthName,
	"NumberWithDelimiter": numberWithDelimiter,
	"Pace":                pace,
	"Round":               round,
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
func numberWithDelimiter(n int, sep rune) string {
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

func round(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
