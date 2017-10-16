package templatehelpers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap = template.FuncMap{
	"DistanceOfTimeInWords":        distanceOfTimeInWords,
	"DistanceOfTimeInWordsFromNow": distanceOfTimeInWordsFromNow,
	"FormatTime":                   formatTime,
	"FormatTimeWithMinute":         formatTimeWithMinute,
	"InKM":                         inKM,
	"MarshalJSON":                  marshalJSON,
	"MonthName":                    monthName,
	"NumberWithDelimiter":          numberWithDelimiter,
	"Pace":                         pace,
	"RenderTweetContent":           renderTweetContent,
	"RoundToString":                roundToString,
	"ToStars":                      toStars,
}

func distanceOfTimeInWords(to, from time.Time) string {
	d := from.Sub(to)
	min := int(round(d.Minutes()))

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
	return toNonBreakingWhitespace(t.Format("January 2, 2006"))
}

func formatTimeWithMinute(t *time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2, 2006 15:04"))
}

// This is a little tricky, but converts normal spaces to non-breaking spaces
// so that we can guarantee that certain strings will appear entirely on the
// same line. This is useful for a star count for example, because it's easy to
// misread a rating if it's broken up. See here for details:
//
// https://github.com/brandur/sorg/pull/60
func toNonBreakingWhitespace(str string) string {
	return strings.Replace(str, " ", " ", -1)
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

func monthName(m time.Month) string {
	return m.String()
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

// Matches links in a tweet (like protocol://link).
var linkRE = regexp.MustCompile(`(^|[\n ])([\w]+?:\/\/[\w]+[^ "\n\r\t< ]*)`)

// Matches tags in a tweet (like #mix11).
var tagRE = regexp.MustCompile(`#(\w+)`)

// Matches users in a tweet (like #mix11).
var userRE = regexp.MustCompile(`@(\w+)`)

// Renders the content of a tweet to HTML.
func renderTweetContent(content string) string {
	tagMap := make(map[string]string)

	// links like protocol://link
	content = linkRE.ReplaceAllStringFunc(content, func(link string) string {
		matches := linkRE.FindStringSubmatch(link)

		//fmt.Printf("matches = %+v (len %v)\n", matches, len(matches))

		whitespace := matches[1]
		href := matches[2]

		display := href
		if len(href) > 30 {
			display = fmt.Sprintf("%s&hellip;", href[0:30])
		}

		// replace with tags so links don't interfere with subsequent rules
		sum := sha1.Sum([]byte(href))
		tag := base64.URLEncoding.EncodeToString(sum[:])
		tagMap[tag] = fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, href, display)

		// make sure to preserve whitespace before the inserted tag
		return whitespace + tag
	})

	// user links (like @brandur)
	content = userRE.ReplaceAllString(content,
		`<a href="https://www.twitter.com/$1" rel="nofollow">@$1</a>`)

	// hash tag search (like #mix11) -- note like anyone would never use one of
	// these, lol
	content = tagRE.ReplaceAllString(content,
		`<a href="https://search.twitter.com/search?q=$1" rel="nofollow">#$1</a>`)

	// replace the stand-in tags for links generated earlier
	for tag, link := range tagMap {
		content = strings.Replace(content, tag, link, -1)
	}

	// show newlines as line breaks
	content = strings.Replace(content, "\n", "<br>", -1)

	return content
}

// There is no "round" function built into Go :/
func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundToString(f float64) string {
	return fmt.Sprintf("%.1f", f)
}

func toStars(n int) string {
	var stars string
	for i := 0; i < n; i++ {
		stars += "★ "
	}
	return toNonBreakingWhitespace(stars)
}
