package visual

import (
	"os"
	"github.com/mmcdole/gofeed"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"net/url"
)

// How to get the post type.
//t := item.Extensions["wp"]["post_type"]

type Mapping map[string]*gofeed.Item

func ParseRSSFeed(filename string) (Mapping, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fp := gofeed.NewParser()
	visualFeed, err := fp.Parse(file)
	if err != nil {
		return nil, err
	}

	visualMapping := make(map[string]*gofeed.Item)

	for _, item := range visualFeed.Items {
		itemURL, err := url.Parse(item.Link)
		if err != nil {
			err := errors.New("failed to parse visual URL")
			log.Error(err, log.Data{"title": item.Title, "url": item.Link})
			return nil, err
		}

		if _, ok := visualMapping[itemURL.String()]; !ok {
			visualMapping[itemURL.String()] = item
		} else {
			err := errors.New("duplicate entry in visual rss xml")
			log.Error(err, log.Data{"title": item.Title, "url": item.Link})
			return nil, err
		}
	}
	return visualMapping, nil
}
