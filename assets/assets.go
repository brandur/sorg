package assets

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

// Number of simultaneous asset fetches that we should perform.
const numWorkers = 5

// Asset represents a site asset that should be fetch from a local URL and
// stored as a local file.
type Asset struct {
	URL    string
	Target string

	// Err is set if a problem occurred while fetching the asset.
	Err error
}

// Fetch performs an HTTP fetch for each given asset and stores them to their
// corresponding local targets.
func Fetch(assets []*Asset) error {
	var err error
	var numWorked int
	assetsChan := make(chan *Asset, len(assets))
	doneChan := make(chan *Asset, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go workAssets(assetsChan, doneChan)
	}

	for _, asset := range assets {
		assetsChan <- asset
	}

	for {
		select {
		case asset := <-doneChan:
			numWorked++

			// Quit when we're done or after encountering the first error.
			if numWorked == len(assets) || asset.Err != nil {
				err = asset.Err
				goto done
			}
		}
	}

done:
	// Signal workers to stop working.
	close(assetsChan)

	return err
}

func fetchAsset(client *http.Client, asset *Asset) error {
	if _, err := os.Stat(asset.Target); !os.IsNotExist(err) {
		log.Debugf("Skipping asset because local file exists: %v", asset.URL)
		return nil
	}

	log.Debugf("Fetching asset: %v", asset.URL)

	resp, err := client.Get(asset.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected status code %d while fetching: %v",
			resp.StatusCode, asset.URL)
	}

	f, err := os.Create(asset.Target)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// probably not needed
	defer w.Flush()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func workAssets(assetsChan chan *Asset, doneChan chan *Asset) {
	client := &http.Client{}

	// Note that this loop falls through when the channel is closed.
	for asset := range assetsChan {
		err := fetchAsset(client, asset)
		if err != nil {
			asset.Err = err
		}
		doneChan <- asset
	}
}
