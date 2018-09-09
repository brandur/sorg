package talks

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	talk, err := Compile("../content/talks-drafts", "paradise-lost.yaml", true)
	assert.NoError(t, err)

	assert.Equal(t, true, talk.Draft)
	assert.NotEmpty(t, talk.Intro)
	assert.NotEmpty(t, talk.IntroRaw)
	assert.NotEmpty(t, talk.Title)

	publishingInfo := talk.PublishingInfo()
	assert.Contains(t, publishingInfo, talk.Event)
	assert.Contains(t, publishingInfo, talk.Location)
	assert.Contains(t, publishingInfo, talk.Title)

	for i, slide := range talk.Slides {
		if slide.CaptionRaw != "" {
			assert.NotEmpty(t, slide.Caption)
		}

		assert.Equal(t, fmt.Sprintf("%03d", i+1), slide.Number)
	}
}
