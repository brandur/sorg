package stemplate

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

var testTime time.Time

func init() {
	const longForm = "2006/01/02 15:04"
	var err error
	testTime, err = time.Parse(longForm, "2016/07/03 12:34")
	if err != nil {
		panic(err)
	}
}

func TestDowncase(t *testing.T) {
	assert.Equal(t, "hello", downcase("HeLlO"))
}

func TestFormatTimeWithMinute(t *testing.T) {
	assert.Equal(t, "July 3, 2016 12:34", formatTimeWithMinute(&testTime))
}

func TestFormatTimeYearMonth(t *testing.T) {
	assert.Equal(t, "July 2016", formatTimeYearMonth(&testTime))
}

func TestInKM(t *testing.T) {
	assert.Equal(t, 2.342, inKM(2342.0))
}

func TestLazyRetinaImage(t *testing.T) {
	assert.Equal(t,
		`<img class="lazy" src="/assets/images/standin_00.jpg" data-src="/photographs/other/001_large.jpg" data-srcset="/photographs/other/001_large@2x.jpg 2x, /photographs/other/001_large.jpg 1x">`,
		lazyRetinaImage(0, "/photographs/other/", "001"),
	)
}

func TestLazyRetinaImageLightbox(t *testing.T) {
	assert.Equal(t,
		`<a href="/photographs/other/001_large@2x.jpg"><img class="lazy" src="/assets/images/standin_00.jpg" data-src="/photographs/other/001_large.jpg" data-srcset="/photographs/other/001_large@2x.jpg 2x, /photographs/other/001_large.jpg 1x"></a>`,
		lazyRetinaImageLightbox(0, "/photographs/other/", "001", false),
	)

	// Portrait
	assert.Equal(t,
		`<a href="/photographs/other/001_large@2x.jpg"><img class="lazy" src="/assets/images/standin_portrait_00.jpg" data-src="/photographs/other/001_large.jpg" data-srcset="/photographs/other/001_large@2x.jpg 2x, /photographs/other/001_large.jpg 1x"></a>`,
		lazyRetinaImageLightbox(0, "/photographs/other/", "001", true),
	)
}

func TestMonthName(t *testing.T) {
	assert.Equal(t, "July", monthName(time.July))
}

func TestNumberWithDelimiter(t *testing.T) {
	assert.Equal(t, "123", numberWithDelimiter(',', 123))
	assert.Equal(t, "1,234", numberWithDelimiter(',', 1234))
	assert.Equal(t, "12,345", numberWithDelimiter(',', 12345))
	assert.Equal(t, "123,456", numberWithDelimiter(',', 123456))
	assert.Equal(t, "1,234,567", numberWithDelimiter(',', 1234567))
}

func TestPace(t *testing.T) {
	d := time.Duration(60 * time.Second)

	// Easiest case: 1000 m ran in 60 seconds which is 1:00 per km.
	assert.Equal(t, "1:00", pace(1000.0, d))

	// Fast: 2000 m ran in 60 seconds which is 0:30 per km.
	assert.Equal(t, "0:30", pace(2000.0, d))

	// Slow: 133 m ran in 60 seconds which is 7:31 per km.
	assert.Equal(t, "7:31", pace(133.0, d))
}

func TestRandIntn(t *testing.T) {
	assert.Equal(t, 0, randIntn(1))
}

func TestRetinaImageAlt(t *testing.T) {
	assert.Equal(t,
		`<img alt="alt text" loading="lazy" src="/photographs/other/001.jpg" srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x">`,
		string(RetinaImageAlt("/photographs/other/001.jpg", "alt text")),
	)
}

func TestRound(t *testing.T) {
	assert.Equal(t, 0.0, round(0.2))
	assert.Equal(t, 1.0, round(0.8))
	assert.Equal(t, 1.0, round(0.5))
}

func TestToStars(t *testing.T) {
	assert.Equal(t, "", toStars(0))
	assert.Equal(t, "★ ", toStars(1))
	assert.Equal(t, "★ ★ ★ ★ ★ ", toStars(5))
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}
