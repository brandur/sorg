package talks

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	talk, err := Compile("../content", "../content/talks-drafts", "paradise-lost.md", true)
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
		if i == 0 {
			// The first slide should have presenter notes (i.e., content
			// trailing an `???`) because those get used as the talk's intro.
			assert.NotEmpty(t, slide.PresenterNotesRaw)
		} else {
			// And every other slide should have a caption because an empty one
			// next to the slide's photo wouldn't look very good.
			assert.NotEmpty(t, slide.CaptionRaw)
		}

		if slide.CaptionRaw != "" {
			assert.NotEmpty(t, slide.Caption)
		}

		if slide.PresenterNotesRaw != "" {
			assert.NotEmpty(t, slide.PresenterNotes)
		}

		assert.Equal(t, fmt.Sprintf("%03d", i+1), slide.Number)
		assert.NotEmpty(t, slide.ImagePath)
	}
}
