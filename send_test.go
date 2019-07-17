package main

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestMatchNewsletter(t *testing.T) {
	{
		info, draft := matchNewsletter("./content/nanoglyphs-drafts/999-placeholder.md")
		assert.Equal(t, newsletterInfoNanoglyphs, info)
		assert.True(t, draft)
	}

	{
		info, draft := matchNewsletter("./content/nanoglyphs/001-first.md")
		assert.Equal(t, newsletterInfoNanoglyphs, info)
		assert.False(t, draft)
	}

	{
		info, draft := matchNewsletter("./content/passages-drafts/999-placeholder.md")
		assert.Equal(t, newsletterInfoPassages, info)
		assert.True(t, draft)
	}

	{
		info, draft := matchNewsletter("./content/passages/001-first.md")
		assert.Equal(t, newsletterInfoPassages, info)
		assert.False(t, draft)
	}

	{
		info, _ := matchNewsletter("./content/articles/article.md")
		assert.Equal(t, (*newsletterInfo)(nil), info)
	}
}

func TestRenderAndSend(t *testing.T) {
	oldMailgunKey := conf.MailgunAPIKey
	defer func() {
		conf.MailgunAPIKey = oldMailgunKey
	}()

	conf.MailgunAPIKey = ""

	{
		err := renderAndSend(nil, "./content/passages/001-first.md", true, false)
		assert.Error(t, err, "MAILGUN_API_KEY must be configured in the environment")
	}

	conf.MailgunAPIKey = "key"

	{
		err := renderAndSend(nil, "./content/articles/article.md", true, false)
		assert.Error(t, err,
			"'./content/articles/article.md' does not appear to be a known newsletter (check its path)")
	}

	{
		err := renderAndSend(nil, "./content/passages-drafts/999-placeholder.md", true, false)
		assert.Error(t, err,
			"refusing to send a draft newsletter to a live list")
	}
}
