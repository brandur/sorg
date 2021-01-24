package stemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/brandur/modulir/modules/mtemplate"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap = template.FuncMap{
	"Downcase":                downcase,
	"FormatTimeLocal":         formatTimeLocal,
	"FormatTimeWithMinute":    formatTimeWithMinute,
	"FormatTimeYearMonth":     formatTimeYearMonth,
	"InKM":                    inKM,
	"LazyRetinaImage":         lazyRetinaImage,
	"LazyRetinaImageLightbox": lazyRetinaImageLightbox,
	"MonthName":               monthName,
	"NumberWithDelimiter":     numberWithDelimiter,
	"Pace":                    pace,
	"RandIntn":                randIntn,
	"RenderPublishingInfo":    renderPublishingInfo,
	"RetinaImageAlt":          RetinaImageAlt,
	"ToStars":                 toStars,
}

var localLocation *time.Location

func init() {
	var err error
	localLocation, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
}

func downcase(s string) string {
	return strings.ToLower(s)
}

func formatTimeLocal(t *time.Time) string {
	return toNonBreakingWhitespace(t.In(localLocation).Format("January 2, 2006"))
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

func renderPublishingInfo(info map[string]string) template.HTML {
	s := ""

	for k, v := range info {
		s += fmt.Sprintf("<p><strong>%s</strong><br>%s</p>", k, v)
	}

	return template.HTML(s)
}

// RetinaImageAlt is a shortcut for creating an image with
// `mtemplate.ImgAndAlt` and rendering it with `mteplate.RenderHTML`. This is
// mostly for backwards compatibility as the interface was changed around a
// bit.
func RetinaImageAlt(src, alt string) template.HTML {
	return mtemplate.HTMLRender(
		mtemplate.ImgSrcAndAlt(src, alt),
	)
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
