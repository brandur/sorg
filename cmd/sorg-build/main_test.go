package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/brandur/sorg"
	_ "github.com/lib/pq"
	assert "github.com/stretchr/testify/require"
)

var db *sql.DB

func init() {
	// Move up into the project's root so that we in the right place relatively
	// to content/view/layout/etc. directories.
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}

	err = sorg.CreateTargetDirs()
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("postgres", "postgres://localhost/sorg-test?sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func TestCompilePhotos(t *testing.T) {
	//
	// No database
	//

	photos, err := compilePhotos(nil)
	assert.NoError(t, err)
	assert.Equal(t, []*Photo(nil), photos)

	//
	// With empty database
	//

	photos, err = compilePhotos(db)
	assert.NoError(t, err)
	assert.Equal(t, []*Photo(nil), photos)

	//
	// With results
	//

	// TODO: insert photos
	//photos, err = compilePhotos(db)
	//assert.NoError(t, err)
}

func TestCompileReading(t *testing.T) {
	//
	// No database
	//

	err := compileReading(nil)
	assert.NoError(t, err)

	//
	// With empty database
	//

	err = compileReading(db)
	assert.NoError(t, err)

	//
	// With results
	//

	// TODO: insert reading
	//err = compileReading(db)
	//assert.NoError(t, err)
}

func TestCompileRuns(t *testing.T) {
	//
	// No database
	//

	err := compileRuns(nil)
	assert.NoError(t, err)

	//
	// With empty database
	//

	err = compileRuns(db)
	assert.NoError(t, err)

	//
	// With results
	//

	// TODO: insert runs
	//err = compileRuns(db)
	//assert.NoError(t, err)
}

func TestCompileTwitter(t *testing.T) {
	//
	// No database
	//

	err := compileTwitter(nil)
	assert.NoError(t, err)

	//
	// With empty database
	//

	err = compileTwitter(db)
	assert.NoError(t, err)

	//
	// With results
	//

	now := time.Now()
	tweet := &Tweet{
		Content:    "Hello, world!",
		OccurredAt: &now,
		Slug:       "1234",
	}
	insertTweet(t, tweet, false)

	err = compileTwitter(db)
	assert.NoError(t, err)
}

func TestGetLocals(t *testing.T) {
	locals := getLocals("Title", map[string]interface{}{
		"Foo": "Bar",
	})

	assert.Equal(t, "Bar", locals["Foo"])
	assert.Equal(t, sorg.Release, locals["Release"])
	assert.Equal(t, "Title", locals["Title"])
}

func TestIsHidden(t *testing.T) {
	assert.Equal(t, true, isHidden(".gitkeep"))
	assert.Equal(t, false, isHidden("article"))
}

func TestSplitFrontmatter(t *testing.T) {
	frontmatter, content, err := splitFrontmatter(`---
foo: bar
---

other`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = splitFrontmatter(`other`)
	assert.NoError(t, err)
	assert.Equal(t, "", frontmatter)
	assert.Equal(t, "other", content)

	frontmatter, content, err = splitFrontmatter(`---
foo: bar
---
`)
	assert.NoError(t, err)
	assert.Equal(t, "foo: bar", frontmatter)
	assert.Equal(t, "", content)

	frontmatter, content, err = splitFrontmatter(`foo: bar
---
`)
	assert.Equal(t, errBadFrontmatter, err)
}

func insertTweet(t *testing.T, tweet *Tweet, reply bool) {
	_, err := db.Exec(`
		INSERT INTO events
			(content, occurred_at, metadata, slug, type)
		VALUES
			($1, $2, hstore('reply', $3), $4, $5)
	`, tweet.Content, tweet.OccurredAt, reply, tweet.Slug, "twitter")
	assert.NoError(t, err)
}
