package main

import (
	"fmt"

	"github.com/broderickhyman/go-githubupdate/updater"
)

var (
	version string
)

func main() {
	fmt.Printf("Hello to hello go version: %s\n", version)

	u := updater.NewUpdater(version, "broderickhyman", "go-githubupdate", "update-")
	if err := u.BackgroundUpdater(); err != nil {
		fmt.Println(err)
	}
}
