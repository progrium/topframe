<img align="left" src="./topframe.png" />

# topframe
User programmable screen overlay using web technologies

[![Go Report Card](https://goreportcard.com/badge/github.com/progrium/topframe)](https://goreportcard.com/report/github.com/progrium/topframe)
<a href="https://twitter.com/progriumHQ" title="@progriumHQ on Twitter"><img src="https://img.shields.io/badge/twitter-@progriumHQ-55acee.svg" alt="@progriumHQ on Twitter"></a>
<a href="https://github.com/progrium/topframe/discussions" title="Project Forum"><img src="https://img.shields.io/badge/community-forum-ff69b4.svg" alt="Project Forum"></a>
<a href="https://github.com/sponsors/progrium" title="Sponsor Project"><img src="https://img.shields.io/static/v1?label=sponsor&message=%E2%9D%A4&logo=GitHub" alt="Sponsor Project" /></a>

---
* Display information and always-on-top widgets
* Use HTML/JS/CSS to draw on your screen
* Great for screencasting or streaming overlays
* Edit source, hit save, screen will update
* Less than 300 lines of Go code
* Currently alpha, macOS only ...

## Getting Started

First, [download Go](https://golang.org/dl/) or `brew install go`. If you have Go installed, make sure it is 
version 1.16 or greater:

```
$ go version
go version go1.16.2 darwin/amd64
```

Then use `go get` to download, build, and install the topframe binary into a `PATH` directory:

```
$ GOBIN=/usr/local/bin go get github.com/progrium/topframe
```

Currently, this is the preferred way to install as anything else requires a much more elaborate
release process with Apple code signing, etc. Specifying `GOBIN` is optional, but lets you specify
where to install the binary, ensuring it is put in a directory in your `PATH`. 

Running `topframe` will create a `~/.topframe` directory with a default `index.html` used for the
overlay. If you have an `EDITOR` specified, you can run with `-edit` to open this in your preferred editor
so you can start making changes to your topframe overlay immediately:

```
$ topframe -edit
```

### Launching on Startup

Topframe works with `launchd` to run as an agent on startup. You can generate
a plist file with `topframe -agent`, which you can write to a file and move
into `/Library/LaunchAgents`:

```
$ topframe -agent > com.progrium.Topframe.plist
$ sudo mv com.progrium.Topframe.plist /Library/LaunchAgents
```

The generated plist will use the current binary location, so make sure it's
in the right place before generating, or modify the plist file.


## Documentation

There is not a whole lot to topframe! I recommend [reading the source](https://github.com/progrium/topframe/blob/main/topframe.go) as its only a few hundred lines,
but otherwise there is a [wiki](https://github.com/progrium/topframe/wiki) ready to document anything else.

## Getting Help

If you're having trouble, be sure to check [issues](https://github.com/progrium/topframe/issues) to see if its a known issue. Otherwise, feel free to drop
a question into the [discussion forum](https://github.com/progrium/topframe/discussions).

## Contributing

Ideally, topframe will be kept small. Bug fixes and other small PRs are welcome and should be merged quickly.
If you happen to have a large PR that we haven't discussed, you should talk about it in the forum first. In order
to keep the project small, some features suggestions may be held back in favor of determining a good extension point to expose instead.

## About

Topframe started as a 130 line example for [progrium/macdriver](https://github.com/progrium/macdriver).

MIT Licensed