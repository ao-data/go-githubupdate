package main

import (
	"fmt"

	"github.com/ao-data/go-githubupdate/updater"
)

var (
	version string
)

func main() {
	fmt.Printf("Hello to hello go version: %s\n", version)

	u := updater.NewUpdater(version, "ao-data", "go-githubupdate", "update-")
	if err := u.BackgroundUpdater(); err != nil {
		fmt.Println(err)
	}
}
