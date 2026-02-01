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
	updated, err := u.BackgroundUpdater()
	if err != nil {
		fmt.Println(err)
	}
	if updated {
		fmt.Println("Application was updated, please restart.")
	}
}
