package migration

import (
	"os"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/go-ns/log"
	"net/url"
)

type Mapping map[string]*gofeed.Item

func ParseRSSFeed(filename string) (Mapping, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	log.Info("attempting to parse RSS export file", nil)
	fp := gofeed.NewParser()
	visualFeed, err := fp.Parse(file)
	if err != nil {
		return nil, Error{"failed to parse visual RSS feed", err, nil}
	}

	visualMapping := make(map[string]*gofeed.Item)

	log.Info("mapping visual posts by post url", nil)
	for _, item := range visualFeed.Items {
		itemURL, err := url.Parse(item.Link)
		if err != nil {
			return nil, Error{"failed to parse visual URL", err, log.Data{"title": item.Title, "url": item.Link}}
		}

		if _, ok := visualMapping[itemURL.String()]; !ok {
			visualMapping[itemURL.String()] = item
		} else {
			return nil, Error{"duplicate entry in visual RSS xmL", err, log.Data{"title": item.Title, "url": item.Link}}
		}
	}
	log.Info("mapping generated successfully", nil)
	return visualMapping, nil
}
