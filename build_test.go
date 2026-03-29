package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}

func TestExtCanonical(t *testing.T) {
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.JPG"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg?query"))
}

func TestExtImageTarget(t *testing.T) {
	require.Equal(t, ".jpg", extImageTarget(".jpg"))
	require.Equal(t, ".webp", extImageTarget(".heic"))
}

func TestLexicographicBase32(t *testing.T) {
	// Should only incorporate lower case characters.
	require.Equal(t, lexicographicBase32, strings.ToLower(lexicographicBase32))

	// All characters in the encoding set should be lexicographically ordered.
	{
		// This can be replaced with `strings.Clone` come Go 1.20.
		b := make([]byte, len(lexicographicBase32))
		copy(b, lexicographicBase32)

		slices.Sort(b)
		require.Equal(t, lexicographicBase32, string(b))
	}
}

func TestPagePathKey(t *testing.T) {
	require.Equal(t, "about", pagePathKey("./pages/about.ace"))
	require.Equal(t, "about", pagePathKey("./pages-drafts/about.ace"))

	require.Equal(t, "deep/about", pagePathKey("./pages/deep/about.ace"))
	require.Equal(t, "deep/about", pagePathKey("./pages-drafts/deep/about.ace"))

	require.Equal(t, "really/deep/about", pagePathKey("./pages/really/deep/about.ace"))
	require.Equal(t, "really/deep/about", pagePathKey("./pages-drafts/really/deep/about.ace"))
}

func TestSimplifyMarkdownForSummary(t *testing.T) {
	require.Equal(t, "check that links are removed", simplifyMarkdownForSummary("check that [links](/link) are removed"))
	require.Equal(t, "double new lines are gone", simplifyMarkdownForSummary("double new\n\nlines are gone"))
	require.Equal(t, "single new lines are gone", simplifyMarkdownForSummary("single new\nlines are gone"))
	require.Equal(t, "space is trimmed", simplifyMarkdownForSummary(" space is trimmed "))
}

func TestBuildContributionGrid(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		grid := buildContributionGrid(nil)
		require.NotNil(t, grid)
		require.Greater(t, len(grid.Weeks), 50)
		require.NotEmpty(t, grid.MonthLabels)

		// All cells should have count 0.
		for _, week := range grid.Weeks {
			for _, day := range week.Days {
				require.Equal(t, 0, day.Count)
				require.Equal(t, 0, day.Level)
			}
		}
	})

	t.Run("SingleAtom", func(t *testing.T) {
		t.Parallel()
		atoms := []*Atom{
			{PublishedAt: time.Now().Add(-24 * time.Hour)},
		}
		grid := buildContributionGrid(atoms)

		total := 0
		for _, week := range grid.Weeks {
			for _, day := range week.Days {
				total += day.Count
			}
		}
		require.Equal(t, 1, total)
	})

	t.Run("MultipleAtomsSameDay", func(t *testing.T) {
		t.Parallel()
		// Use yesterday noon UTC to avoid timezone boundary issues.
		yesterday := time.Now().AddDate(0, 0, -1)
		noon := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 12, 0, 0, 0, time.UTC)
		atoms := []*Atom{
			{PublishedAt: noon},
			{PublishedAt: noon.Add(-1 * time.Hour)},
			{PublishedAt: noon.Add(-2 * time.Hour)},
		}
		grid := buildContributionGrid(atoms)

		dateStr := noon.Format("2006-01-02")
		var found bool
		for _, week := range grid.Weeks {
			for _, day := range week.Days {
				if day.Date == dateStr {
					require.Equal(t, 3, day.Count)
					require.Positive(t, day.Level)
					found = true
				}
			}
		}
		require.True(t, found, "expected cell for %s not found in grid", dateStr)
	})

	t.Run("OldAtomsExcluded", func(t *testing.T) {
		t.Parallel()
		atoms := []*Atom{
			{PublishedAt: time.Now().AddDate(-2, 0, 0)},
		}
		grid := buildContributionGrid(atoms)

		total := 0
		for _, week := range grid.Weeks {
			for _, day := range week.Days {
				total += day.Count
			}
		}
		require.Equal(t, 0, total, "atoms older than 52 weeks should not appear")
	})

	t.Run("MonthLabelsPresent", func(t *testing.T) {
		t.Parallel()
		grid := buildContributionGrid(nil)
		require.GreaterOrEqual(t, len(grid.MonthLabels), 12)
		for _, m := range grid.MonthLabels {
			require.NotEmpty(t, m.Name)
			require.GreaterOrEqual(t, m.OffsetPct, 0.0)
			require.Less(t, m.OffsetPct, 100.0)
		}
	})
}

func TestTruncateString(t *testing.T) {
	require.Equal(t, "Short string unchanged.", truncateString("Short string unchanged.", 100))

	exactly100Length := strings.Repeat("s", 100)
	require.Equal(t, exactly100Length, truncateString(exactly100Length, 100))

	require.Equal(t,
		"This is a longer string that's going to need truncation and which will be truncated by ending it w …",
		truncateString("This is a longer string that's going to need truncation and which will be truncated by ending it with a space and an ellipsis.", 100),
	)
}
