#!/usr/bin/env bash

go build -o lib/visual-migration
./lib/visual-migration -collections="/content/collections/"
