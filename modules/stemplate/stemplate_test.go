package stemplate

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	_ "github.com/brandur/sorg/modules/stesting"
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

func TestFavicon(t *testing.T) {
	s := favicon("custom", "png")
	assert.Contains(t, s,
		`<link rel="shortcut icon" type="image/png" sizes="192x192" href="/assets/images/favicon/custom-192.png">`)
}

func TestFormatTimeWithMinute(t *testing.T) {
	assert.Equal(t, "July 3, 2016 12:34", formatTimeWithMinute(testTime))
}

func TestFormatTimeYearMonth(t *testing.T) {
	assert.Equal(t, "July 2016", formatTimeYearMonth(testTime))
}

func TestInKM(t *testing.T) {
	assert.Equal(t, 2.342, inKM(2342.0))
}

func TestLazyRetinaImage(t *testing.T) {
	assert.Equal(t,
		`<img class="lazy" src="/assets/images/standin_00.jpg" data-src="/photographs/other/001.jpg" `+
			`data-srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x">`,
		string(lazyRetinaImage(0, "/photographs/other/", "001", ".jpg")),
	)
}

func TestLazyRetinaImageLightbox(t *testing.T) {
	assert.Equal(t,
		`<a href="/photographs/other/001@2x.jpg">`+
			`<img class="lazy" src="/assets/images/standin_00.jpg" data-src="/photographs/other/001.jpg" `+
			`data-srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x"></a>`,
		string(lazyRetinaImageLightbox(0, "/photographs/other/", "001", ".jpg", false, "")),
	)

	// Portrait
	assert.Equal(t,
		`<a href="/photographs/other/001@2x.jpg">`+
			`<img class="lazy" src="/assets/images/standin_portrait_00.jpg" data-src="/photographs/other/001.jpg" `+
			`data-srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x"></a>`,
		string(lazyRetinaImageLightbox(0, "/photographs/other/", "001", ".jpg", true, "")),
	)

	// Link override
	assert.Equal(t,
		`<a href="/photographs/other/001">`+
			`<img class="lazy" src="/assets/images/standin_00.jpg" data-src="/photographs/other/001.jpg" `+
			`data-srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x"></a>`,
		string(lazyRetinaImageLightbox(0, "/photographs/other/", "001", ".jpg", false, "/photographs/other/001")),
	)
}

func TestMod(t *testing.T) {
	assert.Equal(t, 0, mod(2, 2))
	assert.Equal(t, 1, mod(3, 2))
}

func TestMonthName(t *testing.T) {
	assert.Equal(t, "July", monthName(time.July))
}

func TestNanoglyphSignup(t *testing.T) {
	t.Run("InEmail", func(t *testing.T) {
		str := nanoglyphSignup(true)
		assert.Equal(t, "", string(str))
	})

	t.Run("NotEmail", func(t *testing.T) {
		str := nanoglyphSignup(false)
		assert.Contains(t, str, "<form")
	})
}

func TestNumberWithDelimiter(t *testing.T) {
	assert.Equal(t, "123", numberWithDelimiter(',', 123))
	assert.Equal(t, "1,234", numberWithDelimiter(',', 1234))
	assert.Equal(t, "12,345", numberWithDelimiter(',', 12345))
	assert.Equal(t, "123,456", numberWithDelimiter(',', 123456))
	assert.Equal(t, "1,234,567", numberWithDelimiter(',', 1234567))
}

func TestPace(t *testing.T) {
	d := 60 * time.Second

	// Easiest case: 1000 m ran in 60 seconds which is 1:00 per km.
	assert.Equal(t, "1:00", pace(1000.0, d))

	// Fast: 2000 m ran in 60 seconds which is 0:30 per km.
	assert.Equal(t, "0:30", pace(2000.0, d))

	// Slow: 133 m ran in 60 seconds which is 7:31 per km.
	assert.Equal(t, "7:31", pace(133.0, d))
}

func TestRandIntN(t *testing.T) {
	assert.Equal(t, 0, randIntN(1))
}

func TestRetinaImageAlt(t *testing.T) {
	assert.Equal(t,
		`<img alt="alt text" loading="lazy" src="/photographs/other/001.jpg" `+
			`srcset="/photographs/other/001@2x.jpg 2x, /photographs/other/001.jpg 1x">`,
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

func TestURLBaseExt(t *testing.T) {
	assert.Equal(t, "", urlBaseExt("https://example.com/video"))
	assert.Equal(t, "jpg", urlBaseExt("https://example.com/image.JPG"))
	assert.Equal(t, "mp4", urlBaseExt("https://example.com/video.mp4"))
	assert.Equal(t, "webm", urlBaseExt("https://example.com/video.webm"))
}

func TestURLBaseFile(t *testing.T) {
	assert.Equal(t, "video", urlBaseFile("https://example.com/video"))
	assert.Equal(t, "image.JPG", urlBaseFile("https://example.com/image.JPG"))
	assert.Equal(t, "video.mp4", urlBaseFile("https://example.com/video.mp4"))
	assert.Equal(t, "video.webm", urlBaseFile("https://example.com/video.webm"))
}
