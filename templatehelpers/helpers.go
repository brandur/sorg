package templatehelpers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap template.FuncMap = template.FuncMap{
	"FormatTime":  formatTime,
	"InKM":        inKM,
	"MarshalJSON": marshalJSON,
	"MonthName":   monthName,
	"Pace":        pace,
	"Round":       round,
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
