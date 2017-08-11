package main

import (
	"fmt"

	"github.com/pcdummy/go-githubupdate/updater"
)

var (
	version string
)

func main() {
	fmt.Printf("Hello to hello go version: %s\n", version)

	u := updater.NewUpdater(version, "pcdummy", "go-githupupdate", "update-")
	if err := u.BackgroundUpdater(); err != nil {
		fmt.Println(err)
	}
}
