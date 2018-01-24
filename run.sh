#!/usr/bin/env bash

go build -o lib/visual-migration
./lib/visual-migration -collectionsDir="/content/collections"
#./lib/visual-migration -collectionsDir="/Users/dave/Desktop/zebedee-data/content/zebedee/collections"
