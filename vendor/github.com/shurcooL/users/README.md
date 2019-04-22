users
=====

[![Build Status](https://travis-ci.org/shurcooL/users.svg?branch=master)](https://travis-ci.org/shurcooL/users) [![GoDoc](https://godoc.org/github.com/shurcooL/users?status.svg)](https://godoc.org/github.com/shurcooL/users)

Package users provides a users service definition.

Installation
------------

```bash
go get -u github.com/shurcooL/users
```

Directories
-----------

| Path                                                               | Synopsis                                                                       |
|--------------------------------------------------------------------|--------------------------------------------------------------------------------|
| [asanaapi](https://godoc.org/github.com/shurcooL/users/asanaapi)   | Package asanaapi implements users.Service using Asana API client.              |
| [fs](https://godoc.org/github.com/shurcooL/users/fs)               | Package fs implements an in-memory users.Store backed by a virtual filesystem. |
| [githubapi](https://godoc.org/github.com/shurcooL/users/githubapi) | Package githubapi implements users.Service using GitHub API client.            |

License
-------

-	[MIT License](https://opensource.org/licenses/mit-license.php)
