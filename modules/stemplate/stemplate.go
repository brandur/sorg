package stemplate

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brandur/modulir/modules/mtemplate"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap = template.FuncMap{
	"FormatTimeWithMinute":    formatTimeWithMinute,
	"FormatTimeYearMonth":     formatTimeYearMonth,
	"InKM":                    inKM,
	"LazyRetinaImage":         lazyRetinaImage,
	"LazyRetinaImageLightbox": lazyRetinaImageLightbox,
	"MonthName":               monthName,
	"NumberWithDelimiter":     numberWithDelimiter,
	"Pace":                    pace,
	"RandIntn":                randIntn,
	"RenderTweetContent":      renderTweetContent,
	"ToStars":                 toStars,
}

func formatTimeYearMonth(t *time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2006"))
}

func formatTimeWithMinute(t *time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2, 2006 15:04"))
}

// Produces a retina-compatible photograph that's lazy loaded. Largely used for
// the photographs and sequences sets.
func lazyRetinaImage(index int, path, slug string) string {
	return lazyRetinaImageLightboxMaybe(index, path, slug, false, false)
}

// Same as the above, but also allows the image to be clicked to get a
// lightbox.
func lazyRetinaImageLightbox(index int, path, slug string, portrait bool) string {
	return lazyRetinaImageLightboxMaybe(index, path, slug, portrait, true)
}

func lazyRetinaImageLightboxMaybe(index int, path, slug string, portrait, lightbox bool) string {
	slug = mtemplate.QueryEscape(slug)
	largePath := path + slug + "_large.jpg"
	largePathRetina := path + slug + "_large@2x.jpg"

	var standinPath string
	if portrait {
		// We only have one portrait standin currently (thus `% 1`).
		standinPath = fmt.Sprintf("/assets/images/standin_portrait_0%d.jpg", index%1)
	} else {
		standinPath = fmt.Sprintf("/assets/images/standin_0%d.jpg", index%5)
	}

	code := fmt.Sprintf(`<img class="lazy" src="%s" data-src="%s" data-srcset="%s 2x, %s 1x">`,
		standinPath, largePath, largePathRetina, largePath)

	if lightbox {
		code = fmt.Sprintf(`<a href="%s">%s</a>`, largePathRetina, code)
	}

	return code
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

func randIntn(bound int) int {
	return rand.Intn(bound)
}

// Matches links in a tweet (like protocol://link).
//
// Note that the last character isn't allowed to match a few extra characters
// in case the link was wrapped in parenthesis, ended a sentence, or the like.
var linkRE = regexp.MustCompile(`(^|[\n ])([\w]+?:\/\/[\w]+[^ "\n\r\t< ]*[^ "\n\r\t<. ])`)

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
		if len(href) > 50 {
			display = fmt.Sprintf("%s&hellip;", href[0:50])
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
	content = strings.Replace(content, "\n", `<div class="tweet-linebreak">`, -1)

	return content
}

// There is no "round" function built into Go :/
func round(f float64) float64 {
	return math.Floor(f + .5)
}

func toStars(n int) string {
	var stars string
	for i := 0; i < n; i++ {
		stars += "★ "
	}
	return toNonBreakingWhitespace(stars)
}
