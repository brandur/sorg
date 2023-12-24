package squantified

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yosssi/ace"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/modulir/modules/mmarkdown"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/scommon"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

// ReadTwitterData reads Twitter data from a TOML data file and does a little
// bit of post-processing to add some convenience properties and an HTML
// rendering of each tweet.
func ReadTwitterData(c *modulir.Context, source string) ([]*Tweet, error) {
	var tweetDB TweetDB

	err := retryOnce(c, func() error {
		return mtoml.ParseFile(c, source, &tweetDB)
	})
	if err != nil {
		return nil, err
	}

	for _, tweet := range tweetDB.Tweets {
		if tweet.Entities != nil {
			for _, media := range tweet.Entities.Medias {
				if media.Type == "photo" {
					// Hot-linked original Twitter version of photos (i.e. not
					// the one we download and cache locally)
					//
					// tweet.ImageURLs = append(tweet.ImageURLs, media.URL)

					ext := filepath.Ext(media.URL)
					tweet.ImageURLs = append(tweet.ImageURLs,
						fmt.Sprintf("/photographs/twitter/%v-%v%s", tweet.ID, media.ID, ext))
				}
			}
		}

		if tweet.Reply != nil || strings.HasPrefix(tweet.Text, "@") {
			tweet.ReplyOrMention = true
		}

		tweet.TextHTML = tweetTextToHTML(tweet)
	}

	return tweetDB.Tweets, nil
}

// RenderRuns renders the `/runs` page by fetching and processing data.
//
// This traditionally used a Black Swan database for run information, but I've
// deprecated that project, so to work again it needs to be converted over to
// use a qself flat file containing run information, like Goodreads and Twitter
// already do in this file.
func RenderRuns(c *modulir.Context, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{},
) error {
	runs, err := getRunsData(c, scommon.DataDir+"/strava.toml")
	if err != nil {
		return err
	}

	var lastYearXDays []string
	var lastYearYDistances []float64

	var byYearXYears []string
	var byYearYDistances []float64

	// Needs to be converted to a qself flat file to work again.
	/*
		lastYearXDays, lastYearYDistances, err := getRunsLastYearData(db)
		if err != nil {
			return err
		}

		byYearXYears, byYearYDistances, err := getRunsByYearData(db)
		if err != nil {
			return err
		}
	*/

	locals := getLocals("Running", map[string]interface{}{
		"Runs": runs,

		// chart: runs over last year
		"LastYearXDays":      lastYearXDays,
		"LastYearYDistances": lastYearYDistances,

		// chart: run distance by year
		"ByYearXYears":     byYearXYears,
		"ByYearYDistances": byYearYDistances,
	})

	return mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/runs/index.ace",
		c.TargetDir+"/runs/index.html", getAceOptions(viewsChanged), locals)
}

// RenderTwitter renders the `/twitter` page.
func RenderTwitter(c *modulir.Context, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{},
	tweets []*Tweet, withReplies bool,
) error {
	tweetsWithoutReplies := make([]*Tweet, 0, len(tweets))
	for _, tweet := range tweets {
		if tweet.ReplyOrMention {
			continue
		}

		tweetsWithoutReplies = append(tweetsWithoutReplies, tweet)
	}

	target := "index.html"
	ts := tweets
	if withReplies {
		target = "with-replies"
	} else {
		ts = tweetsWithoutReplies
	}

	tweetsByYearAndMonth := groupTwitterByYearAndMonth(ts)
	tweetCountXMonths, tweetCountYCounts := getTwitterByMonth(ts)

	locals := getLocals("Twitter", map[string]interface{}{
		"NumTweets":            len(tweetsWithoutReplies),
		"NumTweetsWithReplies": len(tweets),
		"TweetsByYearAndMonth": tweetsByYearAndMonth,
		"WithReplies":          withReplies,

		// chart: tweets by month
		"TweetCountXMonths": tweetCountXMonths,
		"TweetCountYCounts": tweetCountYCounts,
	})

	err := mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/twitter/index.ace",
		c.TargetDir+"/twitter/"+target, getAceOptions(viewsChanged), locals)
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private types
//
//
//
//////////////////////////////////////////////////////////////////////////////

//
// Goodreads
//

// Reading is a single Goodreads book stored to a TOML file.
type Reading struct {
	Authors       []*ReadingAuthor `toml:"authors"`
	ID            int              `toml:"id"`
	ISBN          string           `toml:"isbn"`
	ISBN13        string           `toml:"isbn13"`
	NumPages      int              `toml:"num_pages"`
	PublishedYear int              `toml:"published_year"`
	ReadAt        time.Time        `toml:"read_at"`
	Rating        int              `toml:"rating"`
	Review        string           `toml:"review"`
	ReviewHTML    template.HTML    `toml:"-"`
	ReviewID      int              `toml:"review_id"`
	Title         string           `toml:"title"`

	// AuthorsDisplay is just the names of all authors combined together for
	// display on a page.
	AuthorsDisplay string `toml:"-"`
}

// ReadingAuthor is a single Goodreads author stored to a TOML file.
type ReadingAuthor struct {
	ID   int    `toml:"id"`
	Name string `toml:"name"`
}

// ReadingDB is a database of Goodreads readings stored to a TOML file.
type ReadingDB struct {
	Readings []*Reading `toml:"readings"`
}

//
// Strava
//

// Run is a run as downloaded from Strava.
type Run struct {
	// Distance is the distance traveled for the run in meters.
	Distance float64

	// ElevationGain is the total gain in elevation in meters.
	ElevationGain float64

	// LocationCity is the closest city to which the run occurred. It may be
	// an empty string if Strava wasn't able to match anything.
	LocationCity string

	// MovingTime is the amount of time that the run took.
	MovingTime time.Duration

	// OccurredAt is the local time in which the run occurred. Note that we
	// don't use UTC here so as to not make runs in other timezones look to
	// have occurred at crazy times.
	OccurredAt *time.Time
}

// RunDB is a database of runs stored to a TOML file.
type RunDB struct {
	Runs []*Run `toml:"runs"`
}

//
// Twitter
//

// TweetDB is a database of tweets stored to a TOML file.
type TweetDB struct {
	Tweets []*Tweet `toml:"tweets"`
}

// Tweet is a single tweet stored to a TOML file.
type Tweet struct {
	CreatedAt     time.Time      `toml:"created_at"`
	Entities      *TweetEntities `toml:"entities"`
	FavoriteCount int            `toml:"favorite_count"`
	ID            int64          `toml:"id"`
	Reply         *TweetReply    `toml:"reply"`
	Retweet       *TweetRetweet  `toml:"retweet"`
	RetweetCount  int            `toml:"retweet_count"`
	Text          string         `toml:"text"`

	// ImageURLs are the URLs of all images associated with a given tweet, if
	// any.
	ImageURLs []string `toml:"-"`

	// ReplyOrMention is assigned to tweets which are either a direct mention
	// or reply, and which therefore don't go in the main timeline. It gives us
	// an easy way to access this information from a template.
	ReplyOrMention bool `toml:"-"`

	// TextHTML is Text rendered to HTML using a variety of special Twitter
	// rules. It's rendered once and added to the struct so that it can be
	// reused across multiple pages.
	TextHTML template.HTML `toml:"-"`
}

// TweetEntities contains various multimedia entries that may be contained in a
// tweet.
type TweetEntities struct {
	Medias       []*TweetEntitiesMedia       `toml:"medias"`
	URLs         []*TweetEntitiesURL         `toml:"urls"`
	UserMentions []*TweetEntitiesUserMention `toml:"user_mentions"`
}

// TweetEntitiesMedia is an image or video stored in a tweet.
type TweetEntitiesMedia struct {
	ID   int64  `toml:"id"`
	Type string `toml:"type"`
	URL  string `toml:"url"`

	// Internal
	ext string `toml:"-"`
}

func extCanonical(originalURL string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		panic(err)
	}

	return strings.ToLower(filepath.Ext(u.Path))
}

func (p *TweetEntitiesMedia) OriginalExt() string {
	if p.ext != "" {
		return p.ext
	}

	p.ext = extCanonical(p.URL)
	return p.ext
}

// TweetEntitiesURL is a URL referenced in a tweet.
type TweetEntitiesURL struct {
	DisplayURL  string `toml:"display_url"`
	ExpandedURL string `toml:"expanded_url"`
	URL         string `toml:"url"`
}

// TweetEntitiesUserMention is another user being mentioned in a tweet.
type TweetEntitiesUserMention struct {
	User   string `toml:"user"`
	UserID int64  `toml:"user_id"`
}

// TweetReply is populated with reply information for when a tweet is a reply.
type TweetReply struct {
	StatusID int64  `toml:"status_id"`
	User     string `toml:"user"`
	UserID   int64  `toml:"user_id"`
}

// TweetRetweet is populated with retweet information for when a tweet is a
// retweet.
type TweetRetweet struct {
	StatusID int64  `toml:"status_id"`
	User     string `toml:"user"`
	UserID   int64  `toml:"user_id"`
}

// tweetMonth holds a collection of tweets grouped by year.
type tweetMonth struct {
	Month  time.Month
	Tweets []*Tweet
}

// tweetYear holds a collection of tweetMonths grouped by year.
type tweetYear struct {
	Year   int
	Months []*tweetMonth
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

func combineAuthors(authors []*ReadingAuthor) string {
	if len(authors) == 0 {
		return ""
	}

	if len(authors) == 1 {
		return authors[0].Name
	}

	display := ""

	for i, author := range authors {
		if i == len(authors)-1 {
			display += " & "
		} else if i > 0 {
			display += ", "
		}

		display += author.Name
	}

	return display
}

// getAceOptions gets a good set of default options for Ace template rendering
// for the project.
func getAceOptions(dynamicReload bool) *ace.Options {
	options := &ace.Options{FuncMap: scommon.HTMLTemplateFuncMap}

	if dynamicReload {
		options.DynamicReload = true
	}

	return options
}

func GetReadingsData(c *modulir.Context, target string) ([]*Reading, error) {
	var readingDB ReadingDB

	err := retryOnce(c, func() error {
		return mtoml.ParseFile(c, target, &readingDB)
	})
	if err != nil {
		return nil, err
	}

	// Sort in reverse chronological order. Books should be roughly sorted
	// like this already, but they're sorted by review ID, which may be out
	// of order compared to the read date.
	sort.Slice(readingDB.Readings, func(i, j int) bool {
		return readingDB.Readings[i].ReadAt.After(readingDB.Readings[j].ReadAt)
	})

	for _, reading := range readingDB.Readings {
		reading.AuthorsDisplay = combineAuthors(reading.Authors)

		// Empty reviews written before 2020. These are poorly written (more
		// than usual even) and often contained spoilers since I used them like
		// notes.
		if reading.ReadAt.Year() < 2020 {
			reading.Review = ""
		} else {
			reading.ReviewHTML = template.HTML(string(mmarkdown.Render(c, []byte(reading.Review))))
		}
	}

	return readingDB.Readings, nil
}

func getRunsData(c *modulir.Context, target string) ([]*Run, error) {
	var runDB RunDB

	err := retryOnce(c, func() error {
		return mtoml.ParseFile(c, target, &runDB)
	})
	if err != nil {
		return nil, err
	}

	return runDB.Runs, nil
}

func getTwitterByMonth(tweets []*Tweet) ([]string, []int) {
	tweetCountXMonths := []string{}
	tweetCountYCounts := []int{}

	var currentMonth time.Month
	var currentYear int

	// Tweets are in reverse chronological order. Iterate backwards so we get
	// chronological order.
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]

		if currentYear == 0 || currentYear != tweet.CreatedAt.Year() || currentMonth != tweet.CreatedAt.Month() {
			tweetCountXMonths = append(tweetCountXMonths, tweet.CreatedAt.Format("Jan 2006"))
			tweetCountYCounts = append(tweetCountYCounts, 1)

			currentMonth = tweet.CreatedAt.Month()
			currentYear = tweet.CreatedAt.Year()
		} else {
			tweetCountYCounts[len(tweetCountYCounts)-1]++
		}
	}

	return tweetCountXMonths, tweetCountYCounts
}

func groupTwitterByYearAndMonth(tweets []*Tweet) []*tweetYear {
	var month *tweetMonth
	var year *tweetYear
	var years []*tweetYear

	for _, tweet := range tweets {
		if year == nil || year.Year != tweet.CreatedAt.Year() {
			year = &tweetYear{tweet.CreatedAt.Year(), nil}
			years = append(years, year)
			month = nil
		}

		if month == nil || month.Month != tweet.CreatedAt.Month() {
			month = &tweetMonth{tweet.CreatedAt.Month(), nil}
			year.Months = append(year.Months, month)
		}

		month.Tweets = append(month.Tweets, tweet)
	}

	return years
}

// Data files (especially Twitter's) can be quite large, and if we having
// something like Vim writing to one, our file watcher may notice the change
// before Vim is finished its write. This causes ioutil to read only a
// partially written file, and the TOML unmarshal below it to subsequently
// fail.
//
// Do some hacky protection against this by retrying once when we encounter an
// error. The process of trying to decode TOML the first time should take
// easily enough time to let Vim finish writing, so we'll pick up the full file
// on the second pass.
//
// Note that this only ever a problem on incremental rebuilds and will never be
// needed otherwise.
func retryOnce(c *modulir.Context, f func() error) error {
	var err error
	for i := 0; i < 2; i++ {
		err = f()
		if err != nil {
			c.Log.Errorf("Errored, but retrying once: %v", err)
			continue
		}
		break
	}
	return err
}

// Match a t.co shortlink at the end of a tweet. These tend to be added by
// Twitter for tweets with media embeds, and aren't really needed for anything
// as the media is already embedded inline.
var endTcoShortLinkRE = regexp.MustCompile(` https://t\.co/\w{5,}$`)

// Matches links in a tweet (like protocol://link).
//
// Note that the last character isn't allowed to match a few extra characters
// in case the link was wrapped in parenthesis, ended a sentence, or the like.
var linkRE = regexp.MustCompile(`(^|[\n ])([\w]+?:\/\/[\w]+[^ "\n\r\t< ]*[^ "\n\r\t<. )])`)

// Matches tags in a tweet (like #mix11).
var tagRE = regexp.MustCompile(`([\s\(]|^)#(\w+)([\s\)]|$)`)

// Matches users in a tweet (like #brandur).
var userRE = regexp.MustCompile(`@(\w+)`)

func tweetTextToHTML(tweet *Tweet) template.HTML {
	content := tweet.Text
	tagMap := make(map[string]string)

	// When tweet media is embedded, Twitter adds one last shortlink back to
	// the original tweet, which we prune here.
	if tweet.Entities != nil && tweet.Entities.Medias != nil {
		content = endTcoShortLinkRE.ReplaceAllString(content, "")
	}

	urlEntitiesMap := make(map[string]*TweetEntitiesURL)
	if tweet.Entities != nil && tweet.Entities.URLs != nil {
		for _, urlEntity := range tweet.Entities.URLs {
			urlEntitiesMap[urlEntity.URL] = urlEntity
		}
	}

	// links like protocol://link
	content = linkRE.ReplaceAllStringFunc(content, func(link string) string {
		matches := linkRE.FindStringSubmatch(link)

		// fmt.Printf("matches = %+v (len %v)\n", matches, len(matches))

		var display string
		whitespace := matches[1]
		href := matches[2]

		// Twitter ships URL entity information from its API. Use it if
		// available to produce a shortened "display" URL and the original
		// expanded URL. Otherwise, just do our own version of it.
		if urlEntity, ok := urlEntitiesMap[href]; ok {
			display = urlEntity.DisplayURL
			href = urlEntity.ExpandedURL
		} else {
			display = href
			display = strings.TrimPrefix(display, "http://")
			display = strings.TrimPrefix(display, "https://")
			if len(display) > 50 {
				display = fmt.Sprintf("%s&hellip;", display[0:50])
			}
		}

		// replace with tags so links don't interfere with subsequent rules
		sum := sha256.Sum224([]byte(href))
		tag := base64.URLEncoding.EncodeToString(sum[:])
		tagMap[tag] = fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, href, display)

		// make sure to preserve whitespace before the inserted tag
		return whitespace + tag
	})

	// URL escape (so HTML etc. isn't rendered)
	content = html.EscapeString(content)

	// user links (like @brandur)
	content = userRE.ReplaceAllString(content,
		`<a href="https://www.twitter.com/$1" rel="nofollow">@$1</a>`)

	// hash tag search (like #mix11) -- note like anyone would never use one of
	// these, lol
	content = tagRE.ReplaceAllString(content,
		`$1<a href="https://search.twitter.com/search?q=$2" rel="nofollow">#$2</a>$3`)

	// replace the stand-in tags for links generated earlier
	for tag, link := range tagMap {
		content = strings.ReplaceAll(content, tag, link)
	}

	// show newlines as line breaks
	content = strings.ReplaceAll(content, "\n", `<div class="tweet-linebreak">`)

	return template.HTML(content)
}
