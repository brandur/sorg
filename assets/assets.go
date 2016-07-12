package assets

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

// Number of simultaneous asset fetches that we should perform.
const numWorkers = 5

// Asset represents a site asset that should be fetch from a local URL and
// stored as a local file.
type Asset struct {
	URL    string
	Target string

	Err error
}

// Fetch performs an HTTP fetch for each given asset and stores them to their
// corresponding local targets.
func Fetch(assets []*Asset) error {
	var wg sync.WaitGroup
	wg.Add(len(assets))

	assetsChan := make(chan *Asset, len(assets))

	// Signal workers to stop looping and shut down.
	defer close(assetsChan)

	for i := 0; i < numWorkers; i++ {
		go workAssets(assetsChan, &wg)
	}

	for _, asset := range assets {
		assetsChan <- asset
	}

	wg.Wait()

	// This is not the greatest possible approach because we have to wait for
	// all assets to be processed, but practically problems should be
	// relatively rare. Implement fast timeouts so we can recover in
	// degenerate cases.
	for _, asset := range assets {
		if asset.Err != nil {
			return asset.Err
		}
	}

	return nil
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

func workAssets(assetsChan chan *Asset, wg *sync.WaitGroup) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Note that this loop falls through when the channel is closed.
	for asset := range assetsChan {
		err := fetchAsset(client, asset)
		if err != nil {
			asset.Err = err
		}
		wg.Done()
	}
}
