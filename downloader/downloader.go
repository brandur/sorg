package downloader

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Number of simultaneous file fetches that we should perform.
const numWorkers = 5

// File represents a site file that should be fetch from a local URL and
// stored as a local file.
type File struct {
	URL    string
	Target string

	Err error
}

// Fetch performs an HTTP fetch for each given file and stores them to their
// corresponding local targets.
func Fetch(files []*File) error {
	var wg sync.WaitGroup
	wg.Add(len(files))

	filesChan := make(chan *File, len(files))

	// Signal workers to stop looping and shut down.
	defer close(filesChan)

	for i := 0; i < numWorkers; i++ {
		go workFiles(filesChan, &wg)
	}

	for _, file := range files {
		filesChan <- file
	}

	wg.Wait()

	// This is not the greatest possible approach because we have to wait for
	// all files to be processed, but practically problems should be
	// relatively rare. Implement fast timeouts so we can recover in
	// degenerate cases.
	for _, file := range files {
		if file.Err != nil {
			return file.Err
		}
	}

	return nil
}

func fetchFile(client *http.Client, file *File) error {
	log.Debugf("Fetching file: %v", file.URL)

	resp, err := client.Get(file.URL)
	if err != nil {
		return fmt.Errorf("Error fetching %v: %v", file.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected status code fetching %v: %d",
			file.URL, resp.StatusCode)
	}

	f, err := os.Create(file.Target)
	if err != nil {
		return fmt.Errorf("Error creating %v: %v", file.Target, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// probably not needed
	defer w.Flush()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("Error copying to %v from HTTP response: %v", file.Target, err)
	}

	return nil
}

func workFiles(filesChan chan *File, wg *sync.WaitGroup) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Note that this loop falls through when the channel is closed.
	for file := range filesChan {
		err := fetchFile(client, file)
		if err != nil {
			file.Err = err
		}
		wg.Done()
	}
}
