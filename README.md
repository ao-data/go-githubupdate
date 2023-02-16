go-githubupdate
===============

Self update with github releases. Inspired by [go-selfupdate](https://github.com/sanbornm/go-selfupdate)

## Features

* Checks github releases for a newer version and fetches binaries from there
* Should work on Mac, Linux, Arm and Windows (will get tested soon)

## Quickstart

### Enable your app to Self Update

```go
u := updater.NewUpdater(
    version,                // Current version
    "ao-data",       // Your organization or user
    "go-githubupdate",      // Your repo
    "update-",              // Prefix for the files, full name will be eg: update-linux-amd64.gz, update-windows-amd64.exe.gz
)

if err := u.BackgroundUpdater(); err != nil {
    fmt.Println(err)
}
```

### Upload gzip compressed binaries to github releases

An update is as easy as creating a new release on Github with the version Number as title
and the binaries named like:

- update-linux-amd64.gz
- update-darwin-amd64.gz
- update-windows-amd64.exe.gz


## License

MIT
