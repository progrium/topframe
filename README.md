<img align="left" src="./topframe.png" />

# topframe
User programmable screen overlay using web technologies

[![Go Report Card](https://goreportcard.com/badge/github.com/progrium/topframe)](https://goreportcard.com/report/github.com/progrium/topframe)
<a href="https://twitter.com/progriumHQ" title="@progriumHQ on Twitter"><img src="https://img.shields.io/badge/twitter-@progriumHQ-55acee.svg" alt="@progriumHQ on Twitter"></a>
<a href="https://github.com/progrium/topframe/discussions" title="Project Forum"><img src="https://img.shields.io/badge/community-forum-ff69b4.svg" alt="Project Forum"></a>
<a href="https://github.com/sponsors/progrium" title="Sponsor Project"><img src="https://img.shields.io/static/v1?label=sponsor&message=%E2%9D%A4&logo=GitHub" alt="Sponsor Project" /></a>

---
* Display information or place always-on-top widgets
* Use HTML/JS/CSS to draw on your screen
* Great for screencasting or streaming overlays
* Edit source, hit save, screen will update
* Less than 300 lines of Go code
* Currently alpha, macOS only ...

## Getting Started

 * download or run shell command
 * run with edit flag `topframe --edit`

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

 * wiki

## Getting Help

 * forum

## Contributing

 * maintainers wanted
 * small PRs welcome
 * big PRs discuss first

## About

MIT Licensed