package templatehelpers

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

var testTime time.Time

func init() {
	const longForm = "2006/01/02"
	var err error
	testTime, err = time.Parse(longForm, "2016/07/03")
	if err != nil {
		panic(err)
	}
}

func TestFormatTime(t *testing.T) {
	assert.Equal(t, "July 3, 2016", formatTime(&testTime))
}

func TestInKM(t *testing.T) {
	assert.Equal(t, 2.342, inKM(2342.0))
}

func TestMarshalJSON(t *testing.T) {
	str := marshalJSON(map[string]string{})
	assert.Equal(t, "{}", str)

	str = marshalJSON(7)
	assert.Equal(t, "7", str)

	str = marshalJSON([]int{1, 2, 3})
	assert.Equal(t, "[1,2,3]", str)
}

func TestMonthName(t *testing.T) {
	assert.Equal(t, "July", monthName(&testTime))
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

func TestRound(t *testing.T) {
	assert.Equal(t, "1.2", round(1.234))
	assert.Equal(t, "1.0", round(1))
}
