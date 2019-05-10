package squantified

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/stemplate"
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
func RenderReading(c *modulir.Context, db *sql.DB, viewsChanged bool,
	getLocals func(string, map[string]interface{}) map[string]interface{}) error {
	readings, err := getReadingsData(db)
	if err != nil {
		return err
	}

	readingsByYear := groupReadingsByYear(readings)

	readingsByYearXYears, readingsByYearYCounts, err :=
		getReadingsCountByYearData(db)
	if err != nil {
		return err
	}

	pagesByYearXYears, pagesByYearYCounts, err := getReadingsPagesByYearData(db)
	if err != nil {
		return err
	}

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
		c.TargetDir+"/reading/index.html", stemplate.GetAceOptions(viewsChanged), locals)
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

// Reading is a read book procured from Goodreads.
type Reading struct {
	// Author is the full name of the book's author.
	Author string

	// ISBN is the unique identifier for the book.
	ISBN string

	// NumPages are the number of pages in the book. If unavailable, this
	// number will be zero.
	NumPages int

	// OccurredAt is UTC time when the book was read.
	OccurredAt *time.Time

	// Rating is the rating that I assigned to the read book.
	Rating int

	// Title is the title of the book.
	Title string
}

// readingYear holds a collection of readings grouped by year.
type readingYear struct {
	Year     int
	Readings []*Reading
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

func getReadingsData(db *sql.DB) ([]*Reading, error) {
	var readings []*Reading

	if db == nil {
		return readings, nil
	}

	rows, err := db.Query(`
		SELECT
			metadata -> 'author',
			metadata -> 'isbn',
			-- not every book has a number of pages
			(COALESCE(NULLIF(metadata -> 'num_pages', ''), '0'))::int,
			occurred_at,
			(metadata -> 'rating')::int,
			metadata -> 'title'
		FROM events
		WHERE type = 'goodreads'
		ORDER BY occurred_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("Error selecting readings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reading Reading

		err = rows.Scan(
			&reading.Author,
			&reading.ISBN,
			&reading.NumPages,
			&reading.OccurredAt,
			&reading.Rating,
			&reading.Title,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning readings: %v", err)
		}

		readings = append(readings, &reading)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error iterating readings: %v", err)
	}

	return readings, nil
}

func getReadingsCountByYearData(db *sql.DB) ([]string, []int, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYCounts := []int{}

	if db == nil {
		return byYearXYears, byYearYCounts, nil
	}

	rows, err := db.Query(`
		SELECT date_part('year', occurred_at)::text AS year,
			COUNT(*)
		FROM events
		WHERE type = 'goodreads'
		GROUP BY date_part('year', occurred_at)
		ORDER BY date_part('year', occurred_at)
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting reading count by year: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var year string
		var count int

		err = rows.Scan(
			&year,
			&count,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("Error scanning reading count by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating reading count by year: %v", err)
	}

	return byYearXYears, byYearYCounts, nil
}

func getReadingsPagesByYearData(db *sql.DB) ([]string, []int, error) {
	// Give these arrays 0 elements (instead of null) in case no Black Swan
	// data gets loaded but we still need to render the page.
	byYearXYears := []string{}
	byYearYCounts := []int{}

	if db == nil {
		return byYearXYears, byYearYCounts, nil
	}

	rows, err := db.Query(`
		SELECT date_part('year', occurred_at)::text AS year,
			sum((metadata -> 'num_pages')::int)
		FROM events
		WHERE type = 'goodreads'
			AND metadata -> 'num_pages' <> ''
		GROUP BY date_part('year', occurred_at)
		ORDER BY date_part('year', occurred_at)
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("Error selecting reading pages by year: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var year string
		var count int

		err = rows.Scan(
			&year,
			&count,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("Error scanning reading pages by year: %v", err)
		}

		byYearXYears = append(byYearXYears, year)
		byYearYCounts = append(byYearYCounts, count)
	}
	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("Error iterating reading pages by year: %v", err)
	}

	return byYearXYears, byYearYCounts, nil
}

func groupReadingsByYear(readings []*Reading) []*readingYear {
	var year *readingYear
	var years []*readingYear

	for _, reading := range readings {
		if year == nil || year.Year != reading.OccurredAt.Year() {
			year = &readingYear{reading.OccurredAt.Year(), nil}
			years = append(years, year)
		}

		year.Readings = append(year.Readings, reading)
	}

	return years
}
