events
======

[![Build Status](https://travis-ci.org/shurcooL/events.svg?branch=master)](https://travis-ci.org/shurcooL/events) [![GoDoc](https://godoc.org/github.com/shurcooL/events?status.svg)](https://godoc.org/github.com/shurcooL/events)

Package events provides an events service definition.

Installation
------------

```bash
go get -u github.com/shurcooL/events
```

Directories
-----------

| Path                                                                | Synopsis                                                             |
|---------------------------------------------------------------------|----------------------------------------------------------------------|
| [event](https://godoc.org/github.com/shurcooL/events/event)         | Package event defines event types.                                   |
| [fs](https://godoc.org/github.com/shurcooL/events/fs)               | Package fs implements events.Service using a virtual filesystem.     |
| [githubapi](https://godoc.org/github.com/shurcooL/events/githubapi) | Package githubapi implements events.Service using GitHub API client. |

License
-------

-	[MIT License](LICENSE)
