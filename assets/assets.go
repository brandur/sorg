package assets

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
)

type Asset struct {
	URL    string
	Target string
}

func Fetch(assets []Asset) error {
	client := &http.Client{}

	for _, asset := range assets {
		err := fetchAsset(client, asset)
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchAsset(client *http.Client, asset Asset) error {
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
