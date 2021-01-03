package squantified

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/yosssi/ace"
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

// RenderReading renders the `/reading` page by fetching and processing data
// from an attached Black Swan database.
func RenderReading(c *modulir.Context, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{}) error {

	readings, err := getReadingsData(c, scommon.DataDir+"/goodreads.toml")
	if err != nil {
		return err
	}

	// Important: all these functions assume reverse chronological read at
	// order has already been applied.
	readingsByYear := groupReadingsByYear(readings)
	const maxYears = 10
	readingsByYearXYears, readingsByYearYCounts := getReadingsCountByYearData(readings, maxYears)
	pagesByYearXYears, pagesByYearYCounts := getReadingsPagesByYearData(readings, maxYears)

	locals := getLocals("Reading", map[string]interface{}{
		"NumReadings":    len(readings),
		"ReadingsByYear": readingsByYear,

		// chart: readings by year
		"ReadingsByYearXYears":  readingsByYearXYears,
		"ReadingsByYearYCounts": readingsByYearYCounts,

		// chart: pages by year
		"PagesByYearXYears":  pagesByYearXYears,
		"PagesByYearYCounts": pagesByYearYCounts,
	})

	return mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/reading/index.ace",
		c.TargetDir+"/reading/index.html", getAceOptions(viewsChanged), locals)
}

// RenderRuns renders the `/runs` page by fetching and processing data.
//
// This traditionally used a Black Swan database for run information, but I've
// deprecated that project, so to work again it needs to be converted over to
// use a qself flat file containing run information, like Goodreads and Twitter
// already do in this file.
func RenderRuns(c *modulir.Context, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{}) error {

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

// RenderTwitter renders the `/twitter` page by fetching and processing data
// from an attached Black Swan database.
func RenderTwitter(c *modulir.Context, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{}) error {

	tweetsWithReplies, err := getTwitterData(c, scommon.DataDir+"/twitter.toml")
	if err != nil {
		return err
	}

	var tweets []*Tweet
	for _, tweet := range tweetsWithReplies {
		if tweet.ReplyOrMention {
			continue
		}

		tweets = append(tweets, tweet)
	}

	optionsMatrix := map[string]bool{
		"index.html":   false,
		"with-replies": true,
	}

	for page, withReplies := range optionsMatrix {
		ts := tweets
		if withReplies {
			ts = tweetsWithReplies
		}

		tweetsByYearAndMonth := groupTwitterByYearAndMonth(ts)
		tweetCountXMonths, tweetCountYCounts := getTwitterByMonth(ts)

		locals := getLocals("Twitter", map[string]interface{}{
			"NumTweets":            len(tweets),
			"NumTweetsWithReplies": len(tweetsWithReplies),
			"TweetsByYearAndMonth": tweetsByYearAndMonth,
			"WithReplies":          withReplies,

			// chart: tweets by month
			"TweetCountXMonths": tweetCountXMonths,
			"TweetCountYCounts": tweetCountYCounts,
		})

		err = mace.RenderFile(c, scommon.MainLayout, scommon.ViewsDir+"/twitter/index.ace",
			c.TargetDir+"/twitter/"+page, getAceOptions(viewsChanged), locals)
		if err != nil {
			return err
		}
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

// readingYear holds a collection of readings grouped by year.
type readingYear struct {
	Year     int
	Readings []*Reading
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
	Type string `toml:"type"`
	URL  string `toml:"url"`
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

func getReadingsData(c *modulir.Context, target string) ([]*Reading, error) {
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
	}

	return readingDB.Readings, nil
}

func getReadingsCountByYearData(readings []*Reading, maxYears int) ([]int, []int) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []int{}
	byYearYCounts := []int{}

	for _, reading := range readings {
		year := reading.ReadAt.Year()

		if len(byYearXYears) == 0 || byYearXYears[len(byYearXYears)-1] != year {
			if len(byYearXYears) >= maxYears {
				break
			}

			byYearXYears = append(byYearXYears, year)
			byYearYCounts = append(byYearYCounts, 0)
		}

		byYearYCounts[len(byYearYCounts)-1]++
	}

	return byYearXYears, byYearYCounts
}

func getReadingsPagesByYearData(readings []*Reading, maxYears int) ([]int, []int) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []int{}
	byYearYCounts := []int{}

	for _, reading := range readings {
		year := reading.ReadAt.Year()

		if len(byYearXYears) == 0 || byYearXYears[len(byYearXYears)-1] != year {
			if len(byYearXYears) >= maxYears {
				break
			}

			byYearXYears = append(byYearXYears, year)
			byYearYCounts = append(byYearYCounts, 0)
		}

		byYearYCounts[len(byYearYCounts)-1] += reading.NumPages
	}

	return byYearXYears, byYearYCounts
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

// Needs to be converted to a qself flat file to work again.
/*
func getRunsByYearData(db *sql.DB) ([]string, []float64, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYDistances := []float64{}

	if db == nil {
		return byYearXYears, byYearYDistances, nil
	}

	rows, err := db.Query(`
		WITH runs AS (
			SELECT *,
				(metadata -> 'occurred_at_local')::timestamptz AS occurred_at_local,
				-- convert to distance in kilometers
				((metadata -> 'distance')::float / 1000.0) AS distance
			FROM events
			WHERE type = 'strava'
				AND metadata -> 'type' = 'Run'
		)

		SELECT date_part('year', occurred_at_local)::text AS year,
			SUM(distance)
		FROM runs
		GROUP BY year
		ORDER BY year DESC
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting runs by year: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var distance float64
		var year string

		err = rows.Scan(
			&year,
			&distance,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("Error scanning runs by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYDistances = append(byYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating runs by year: %v", err)
	}

	return byYearXYears, byYearYDistances, nil
}

func getRunsLastYearData(db *sql.DB) ([]string, []float64, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	lastYearXDays := []string{}
	lastYearYDistances := []float64{}

	if db == nil {
		return lastYearXDays, lastYearYDistances, nil
	}

	rows, err := db.Query(`
		WITH runs AS (
			SELECT *,
				(metadata -> 'occurred_at_local')::timestamptz AS occurred_at_local,
				-- convert to distance in kilometers
				((metadata -> 'distance')::float / 1000.0) AS distance
			FROM events
			WHERE type = 'strava'
				AND metadata -> 'type' = 'Run'
		),

		runs_days AS (
			SELECT date_trunc('day', occurred_at_local) AS day,
				SUM(distance) AS distance
			FROM runs
			WHERE occurred_at_local > NOW() - '180 days'::interval
			GROUP BY day
			ORDER BY day
		),

		-- generates a baseline series of every day in the last 180 days
		-- along with a zeroed distance which we will then add against the
		-- actual runs we extracted
		days AS (
			SELECT i::date AS day,
				0::float AS distance
			FROM generate_series(NOW() - '180 days'::interval,
				NOW(), '1 day'::interval) i
		)

		SELECT to_char(d.day, 'Mon') AS day,
			d.distance + COALESCE(rd.distance, 0::float)
		FROM days d
			LEFT JOIN runs_days rd ON d.day = rd.day
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting last year runs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var distance float64

		err = rows.Scan(
			&day,
			&distance,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("Error scanning last year runs: %v", err)
		}

		lastYearXDays = append(lastYearXDays, day)
		lastYearYDistances = append(lastYearYDistances, distance)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating last year runs: %v", err)
	}

	return lastYearXDays, lastYearYDistances, nil
}
*/

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

func getTwitterData(c *modulir.Context, target string) ([]*Tweet, error) {
	var tweetDB TweetDB

	err := retryOnce(c, func() error {
		return mtoml.ParseFile(c, target, &tweetDB)
	})
	if err != nil {
		return nil, err
	}

	for _, tweet := range tweetDB.Tweets {
		if tweet.Entities != nil {
			for _, mediaEntity := range tweet.Entities.Medias {
				if mediaEntity.Type == "photo" {
					tweet.ImageURLs = append(tweet.ImageURLs, mediaEntity.URL)
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

func groupReadingsByYear(readings []*Reading) []*readingYear {
	var year *readingYear
	var years []*readingYear

	for _, reading := range readings {
		if year == nil || year.Year != reading.ReadAt.Year() {
			year = &readingYear{reading.ReadAt.Year(), nil}
			years = append(years, year)
		}

		year.Readings = append(year.Readings, reading)
	}

	return years
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

// Matches links in a tweet (like protocol://link).
//
// Note that the last character isn't allowed to match a few extra characters
// in case the link was wrapped in parenthesis, ended a sentence, or the like.
var linkRE = regexp.MustCompile(`(^|[\n ])([\w]+?:\/\/[\w]+[^ "\n\r\t< ]*[^ "\n\r\t<. ])`)

// Matches tags in a tweet (like #mix11).
var tagRE = regexp.MustCompile(`([\s\(]|^)#(\w+)([\s\)]|$)`)

// Matches users in a tweet (like #brandur).
var userRE = regexp.MustCompile(`@(\w+)`)

func tweetTextToHTML(tweet *Tweet) template.HTML {
	content := tweet.Text
	tagMap := make(map[string]string)

	urlEntitiesMap := make(map[string]*TweetEntitiesURL)
	if tweet.Entities != nil && tweet.Entities.URLs != nil {
		for _, urlEntity := range tweet.Entities.URLs {
			urlEntitiesMap[urlEntity.URL] = urlEntity
		}
	}

	// links like protocol://link
	content = linkRE.ReplaceAllStringFunc(content, func(link string) string {
		matches := linkRE.FindStringSubmatch(link)

		//fmt.Printf("matches = %+v (len %v)\n", matches, len(matches))

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
		sum := sha1.Sum([]byte(href))
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
		content = strings.Replace(content, tag, link, -1)
	}

	// show newlines as line breaks
	content = strings.Replace(content, "\n", `<div class="tweet-linebreak">`, -1)

	return template.HTML(content)
}
