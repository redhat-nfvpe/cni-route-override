// +build dev

package assets

import (
	"go/build"
	"log"
	"net/http"

	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/httpfs/union"
)

// Assets contains assets for changes.
var Assets = union.New(map[string]http.FileSystem{
	"/script.js": gopherjs_http.Package("dmitri.shuralyov.com/app/changes/frontend"),
	"/assets":    http.Dir(importPathToDir("dmitri.shuralyov.com/app/changes/_data")),
})

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}
