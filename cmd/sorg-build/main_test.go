package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/brandur/sorg"
	"github.com/brandur/sorg/pool"
	_ "github.com/brandur/sorg/testing"
	_ "github.com/lib/pq"
	assert "github.com/stretchr/testify/require"
)

var db *sql.DB

func init() {
	conf.TargetDir = "./public"
	err := sorg.CreateOutputDirs(conf.TargetDir)
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("postgres", "postgres://localhost/sorg-test?sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func TestCompilePhotos(t *testing.T) {
	photos, err := compilePhotos(true)
	assert.NoError(t, err)
	assert.NotZero(t, len(photos))
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

func TestCompileRobots(t *testing.T) {
	dir, err := ioutil.TempDir("", "target")
	assert.NoError(t, err)
	path := path.Join(dir, "robots.txt")

	conf.Drafts = false
	err = compileRobots(path)
	assert.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err)

	conf.Drafts = true
	err = compileRobots(path)
	assert.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err)
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

func TestCompileSeq(t *testing.T) {
	err := compileSeq("sf", true)
	assert.NoError(t, err)
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

func TestEnsureSymlink(t *testing.T) {
	dir, err := ioutil.TempDir("", "symlink")
	assert.NoError(t, err)

	source := path.Join(dir, "source")
	err = ioutil.WriteFile(source, []byte("source"), 0755)
	assert.NoError(t, err)

	dest := path.Join(dir, "symlink-dest")

	//
	// Case 1: Symlink does not exist
	//

	err = ensureSymlink(source, dest)
	assert.NoError(t, err)

	actual, err := os.Readlink(dest)
	assert.Equal(t, source, actual)

	//
	// Case 2: Symlink does exist
	//
	// Consists solely of re-running the previous test case.
	//

	err = ensureSymlink(source, dest)
	assert.NoError(t, err)

	actual, err = os.Readlink(dest)
	assert.Equal(t, source, actual)

	//
	// Case 3: Symlink file exists, but source doesn't
	//

	err = os.RemoveAll(dest)
	assert.NoError(t, err)

	source = path.Join(dir, "source")
	err = ioutil.WriteFile(source, []byte("source"), 0755)
	assert.NoError(t, err)

	err = ensureSymlink(source, dest)
	assert.NoError(t, err)

	actual, err = os.Readlink(dest)
	assert.Equal(t, source, actual)
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

func TestRunTasks(t *testing.T) {
	conf.Concurrency = 3

	//
	// Success case
	//

	tasks := []*pool.Task{
		pool.NewTask(func() error { return nil }),
		pool.NewTask(func() error { return nil }),
		pool.NewTask(func() error { return nil }),
	}
	assert.Equal(t, true, runTasks(tasks))

	//
	// Failure case (1 error)
	//

	tasks = []*pool.Task{
		pool.NewTask(func() error { return nil }),
		pool.NewTask(func() error { return nil }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
	}
	assert.Equal(t, false, runTasks(tasks))

	//
	// Failure case (11 errors)
	//
	// Here we'll exit with a "too many errors" message.
	//

	tasks = []*pool.Task{
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
		pool.NewTask(func() error { return fmt.Errorf("error") }),
	}
	assert.Equal(t, false, runTasks(tasks))
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
