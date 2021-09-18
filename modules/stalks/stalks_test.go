package stalks

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/brandur/sorg/modules/stesting"
)

func TestRender(t *testing.T) {
	// stesting.init changes the working directory to root, which is why these
	// relative paths work.
	talk, err := Render(
		stesting.NewContext(),
		"./content",
		"./content/talks-drafts",
		"paradise-lost.md",
	)
	assert.NoError(t, err)

	assert.Equal(t, true, talk.Draft)
	assert.NotEmpty(t, talk.Intro)
	assert.NotEmpty(t, talk.IntroRaw)
	assert.NotEmpty(t, talk.Title)

	publishingInfo := talk.PublishingInfo()
	assert.Contains(t, talk.Event, publishingInfo["Event"])
	assert.Contains(t, talk.Location, publishingInfo["Location"])
	assert.Contains(t, talk.Title, publishingInfo["Talk"])

	for i, slide := range talk.Slides {
		if i == 0 {
			// The first slide should have presenter notes (i.e., content
			// trailing an `???`), but we'll have moved them to `talk.Intro`
			// and emptied their contents.
			assert.Empty(t, slide.PresenterNotesRaw)
		} else {
			// All other slides should have presenter notes.
			assert.NotEmpty(t, slide.PresenterNotesRaw)
		}

		if slide.ContentRaw != "" {
			assert.NotEmpty(t, slide.Content)
		}

		if slide.PresenterNotesRaw != "" {
			assert.NotEmpty(t, slide.PresenterNotes)
		}

		assert.Equal(t, fmt.Sprintf("%03d", i+1), slide.Number)
		assert.NotEmpty(t, slide.ImagePath)
	}
}
