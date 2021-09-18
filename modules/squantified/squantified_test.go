package squantified

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCombineAuthors(t *testing.T) {
	assert.Equal(t,
		"Alex",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
		}),
	)

	assert.Equal(t,
		"Alex & Kate",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
		}),
	)

	assert.Equal(t,
		"Alex, Kate & Scan",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
			{Name: "Scan"},
		}),
	)

	assert.Equal(t,
		"Alex, Kate, Scan & Will",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
			{Name: "Scan"},
			{Name: "Will"},
		}),
	)
}

func TestRenderTweet(t *testing.T) {
	// short link
	assert.Equal(t,
		`<a href="https://example.com" rel="nofollow">example.com</a>`,
		string(tweetTextToHTML(&Tweet{Text: `https://example.com`})),
	)

	// link with whitespace and newlines
	assert.Equal(t,
		`content`+
			`<div class="tweet-linebreak"><div class="tweet-linebreak">`+
			`<a href="https://example.com" rel="nofollow">example.com</a>`+
			`<div class="tweet-linebreak"><div class="tweet-linebreak">`+
			`end`,
		string(tweetTextToHTML(&Tweet{Text: `content

https://example.com

end`})),
	)

	// long link
	assert.Equal(t,
		`<a href="https://example.com/path/to/more/great/stuff/and/this/is/even/longer/now" `+
			`rel="nofollow">example.com/path/to/more/great/stuff/and/this/is/e&hellip;</a>`,
		string(tweetTextToHTML(&Tweet{Text: `https://example.com/path/to/more/great/stuff/and/this/is/even/longer/now`})),
	)

	// TODO: This needs fixing.
	/*
		// long with special characters
		assert.Equal(t,
			`<a href="https://example.com/w/Film_(2005)" rel="nofollow">example.com/w/Film_(2005)</a>.`,
			string(tweetTextToHTML(&Tweet{Text: `https://example.com/w/Film_(2005).`})),
		)

		assert.Equal(t,
			`(in quotes <a href="https://example.com/w/Film_(2005)" rel="nofollow">example.com/w/Film_(2005)</a>).`,
			string(tweetTextToHTML(&Tweet{Text: `(in quotes https://example.com/w/Film_(2005)).`})),
		)
	*/

	// with trailing parenthesis
	assert.Equal(t,
		`(in quotes <a href="https://example.com/" rel="nofollow">example.com/</a>).`,
		string(tweetTextToHTML(&Tweet{Text: `(in quotes https://example.com/).`})),
	)

	// with trailing dot and parenthesis
	assert.Equal(t,
		`(in quotes <a href="https://example.com/" rel="nofollow">example.com/</a>.)`,
		string(tweetTextToHTML(&Tweet{Text: `(in quotes https://example.com/.)`})),
	)

	// html inclued in tweet
	assert.Equal(t,
		`not a &lt;video&gt; tag`,
		string(tweetTextToHTML(&Tweet{Text: `not a <video> tag`})),
	)

	// tag
	assert.Equal(t,
		`<a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a>`,
		string(tweetTextToHTML(&Tweet{Text: `#mix11`})),
	)

	// tag floating with space
	assert.Equal(t,
		` <a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a> `,
		string(tweetTextToHTML(&Tweet{Text: ` #mix11 `})),
	)

	// tag in parenthesis
	assert.Equal(t,
		`(<a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a>)`,
		string(tweetTextToHTML(&Tweet{Text: `(#mix11)`})),
	)

	// HTML entities with a pound should not be tags
	assert.Equal(t,
		`&amp;#39;`,
		string(tweetTextToHTML(&Tweet{Text: `&#39;`})),
	)

	// user
	assert.Equal(t,
		`<a href="https://www.twitter.com/brandur" rel="nofollow">@brandur</a>`,
		string(tweetTextToHTML(&Tweet{Text: `@brandur`})),
	)
}
