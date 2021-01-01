package squantified

import (
	"testing"

	"github.com/brandur/sorg/modules/squantifiedtypes"
	assert "github.com/stretchr/testify/require"
)

func TestRenderTweet(t *testing.T) {
	// short link
	assert.Equal(t,
		`<a href="https://example.com" rel="nofollow">https://example.com</a>`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `https://example.com`})),
	)

	// link with whitespace and newlines
	assert.Equal(t,
		`content`+
			`<div class="tweet-linebreak"><div class="tweet-linebreak">`+
			`<a href="https://example.com" rel="nofollow">https://example.com</a>`+
			`<div class="tweet-linebreak"><div class="tweet-linebreak">`+
			`end`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `content

https://example.com

end`})),
	)

	// long link
	assert.Equal(t,
		`<a href="https://example.com/path/to/more/great/stuff/and/this/is/even/longer/now" rel="nofollow">https://example.com/path/to/more/great/stuff/and/t&hellip;</a>`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `https://example.com/path/to/more/great/stuff/and/this/is/even/longer/now`})),
	)

	// long with special characters
	assert.Equal(t,
		`<a href="https://example.com/w/Film_(2005)" rel="nofollow">https://example.com/w/Film_(2005)</a>.`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `https://example.com/w/Film_(2005).`})),
	)

	// html inclued in tweet
	assert.Equal(t,
		`not a &lt;video&gt; tag`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `not a <video> tag`})),
	)

	// tag
	assert.Equal(t,
		`<a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a>`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `#mix11`})),
	)

	// tag floating with space
	assert.Equal(t,
		` <a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a> `,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: ` #mix11 `})),
	)

	// tag in parenthesis
	assert.Equal(t,
		`(<a href="https://search.twitter.com/search?q=mix11" rel="nofollow">#mix11</a>)`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `(#mix11)`})),
	)

	// HTML entities with a pound should not be tags
	assert.Equal(t,
		`&amp;#39;`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `&#39;`})),
	)

	// user
	assert.Equal(t,
		`<a href="https://www.twitter.com/brandur" rel="nofollow">@brandur</a>`,
		string(tweetTextToHTML(&squantifiedtypes.Tweet{Text: `@brandur`})),
	)
}
