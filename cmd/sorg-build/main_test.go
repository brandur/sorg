package main

import (
	"database/sql"
	"io/ioutil"
	"testing"
	"time"

	"github.com/brandur/sorg"
	_ "github.com/brandur/sorg/testing"
	_ "github.com/lib/pq"
	assert "github.com/stretchr/testify/require"
)

var db *sql.DB

func init() {
	err := sorg.CreateTargetDirs()
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("postgres", "postgres://localhost/sorg-test?sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func TestCompileJavascripts(t *testing.T) {
	dir, err := ioutil.TempDir("", "javascripts")

	file1 := dir + "/file1.js"
	file2 := dir + "/file2.js"
	file3 := dir + "/file3.js"
	out := dir + "/app.js"

	err = ioutil.WriteFile(file1, []byte(`function() { return "file1" }`), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file2, []byte(`function() { return "file2" }`), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file3, []byte(`function() { return "file3" }`), 0755)
	assert.NoError(t, err)

	err = compileJavascripts([]string{file1, file2, file3}, out)
	assert.NoError(t, err)

	actual, err := ioutil.ReadFile(out)
	assert.NoError(t, err)

	expected := `/* file1.js */

(function() {

function() { return "file1" }

}).call(this);

/* file2.js */

(function() {

function() { return "file2" }

}).call(this);

/* file3.js */

(function() {

function() { return "file3" }

}).call(this);

`
	assert.Equal(t, expected, string(actual))
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

func TestCompileStylesheets(t *testing.T) {
	dir, err := ioutil.TempDir("", "stylesheets")

	file1 := dir + "/file1.sass"
	file2 := dir + "/file2.sass"
	file3 := dir + "/file3.css"
	out := dir + "/app.css"

	// The syntax of the first and second files is GCSS and the third is in
	// CSS.
	err = ioutil.WriteFile(file1, []byte("p\n  margin: 10px"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file2, []byte("p\n  padding: 10px"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file3, []byte("p {\n  border: 10px;\n}"), 0755)
	assert.NoError(t, err)

	err = compileStylesheets([]string{file1, file2, file3}, out)
	assert.NoError(t, err)

	actual, err := ioutil.ReadFile(out)
	assert.NoError(t, err)

	// Note that the first two files have no spacing in the output because they
	// go through the GCSS compiler.
	expected := `/* file1.sass */

p{margin:10px;}

/* file2.sass */

p{padding:10px;}

/* file3.css */

p {
  border: 10px;
}

`
	assert.Equal(t, expected, string(actual))
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
