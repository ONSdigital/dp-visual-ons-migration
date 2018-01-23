#!/usr/bin/env bash

go build -o lib/visual-migration
./lib/visual-migration -collections="/content/zebedee"
#./lib/visual-migration -collectionsDir="/Users/dave/Desktop/zebedee-data/content/zebedee/collections"
