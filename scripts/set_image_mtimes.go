package main

//
// UPDATE: This whole script turned out to not be needed because I found the
// `--size-only` options for `aws s3 sync` which ignores mtimes and doesn't
// exactly the right thing as far as image assets are concerned. I've left this
// script here for historical interest only.
//
// sorg pushes to S3 automatically from its GitHub Actions build so that it has
// an autodeploy mechanism on every push.
//
// Syncing is accomplished via the AWS CLI's `aws s3 sync`, which should be
// adequate except for the fact that it decides what to sync up using a file's
// modification time, and Git doesn't preserve modification times. So every
// build has the effect of cloning down a repository, having every file get a
// new mtime, then syncing everything from scratch.
//
// At some point I realized that every build was pushing 100 MB+ of images and
// running on cron every hour, which was getting expensive -- not *hugely*
// expensive, but on the order of single dollar digits a month, which is more
// than the value I was getting out of it.
//
// This script solves at least part of the problem by looking at every image in
// the repository, checking when its last commit was, and then changing the
// modification time of the file to that commit's timestamp. This has the
// effect of giving the file a stable mtime so that it's not pushed to S3 over
// and over again.
//
// Unfortunately it has a downside, which is that `git log` is not very fast,
// and there's no way I can find of efficiently batching up a lot of these
// commands for multiple files at once. As I write this, the script takes just
// over a minute to iterate every file and get its commit time.
//
// A better answer to this might be to stop storing images in the repository
// (which will be unsustainable eventually anyway) and instead put them in
// their own S3 bucket like which is already done for photographs.
//

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

const (
	imagePath = "./content/images/"

	// Number of parallel workers to runw which extract commit timestamps and
	// set modtimes on files.
	//
	// There's a very definite diminishing return to increasing this number,
	// but it does help. Some rough numbers from Mac OS:
	//
	//     1  -->   62s
	//     5  -->   28s
	//     10 -->   21s
	//     20 -->   21s
	//     100 -->  20s
	//
	parallelWorkers = 10
)

//
// ---
//

func main() {
	allImages, err := getAllImages()
	abortOnErr(err)

	imageBatches := batchImages(allImages)

	var wg sync.WaitGroup
	wg.Add(parallelWorkers)

	for i := 0; i < parallelWorkers; i++ {
		workerNum := i

		go func() {
			for _, path := range imageBatches[workerNum] {
				lastCommitTime, err := getLastCommitTime(path)
				abortOnErr(err)

				fmt.Printf("%v --> %v\n", path, lastCommitTime)
				err = os.Chtimes(path, *lastCommitTime, *lastCommitTime)
				abortOnErr(err)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

//
// ---
//

func abortOnErr(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Error encountered: %v", err)
	os.Exit(1)
}

// Breaks a set of images into groups for N parallel workers roughly evenly.
func batchImages(allImages []string) [][]string {
	batches := make([][]string, parallelWorkers)
	imagesPerWorker := int(math.Ceil(float64(len(allImages)) / float64(parallelWorkers-1)))

	for i := 0; i < parallelWorkers; i++ {
		startIndex := i * imagesPerWorker
		endIndex := minInt((i+1)*imagesPerWorker, len(allImages))

		// fmt.Printf("worker %v: start = %v end = %v (per worker = %v, total = %v)\n",
		// 	i, startIndex, endIndex, imagesPerWorker, len(allImages))

		// Thanks for our ceiling math, for cases where there are many workers
		// compared to the amount of work needing to be done, it's possible for
		// the last worker to be beyond the limits of the slice.
		if startIndex >= len(allImages) {
			break
		}

		batches[i] = allImages[startIndex:endIndex]
	}

	return batches
}

// Gets a list of all image paths by using a `git ls-tree` command on the
// target directory.
func getAllImages() ([]string, error) {
	out, err := runCommand("git", "ls-tree", "-r", "--name-only", "HEAD", imagePath)
	if err != nil {
		return nil, xerrors.Errorf("error getting images with `git ls-tree`: %w", err)
	}

	return strings.Split(out, "\n"), nil
}

// Gets the last commit time on a particular image path by using a `git log`
// command.
func getLastCommitTime(path string) (*time.Time, error) {
	out, err := runCommand("git", "log", "--max-count=1", `--pretty=format:%aI`, path)
	if err != nil {
		return nil, xerrors.Errorf("error getting commit time for '%s': %w", path, err)
	}

	lastCommitTime, err := time.Parse("2006-01-02T15:04:05-07:00", out)
	if err != nil {
		return nil, err
	}

	return &lastCommitTime, nil
}

func minInt(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func runCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", xerrors.Errorf("error executing command '%s': %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}
