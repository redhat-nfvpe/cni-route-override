// Package gofontwoff provides the Go font family in Web Open Font Format.
//
// It's a Go package that statically embeds Go font family WOFF data, exposing it via an http.FileSystem.
//
// These fonts were created by the Bigelow & Holmes foundry specifically for the
// Go project. See https://blog.golang.org/go-fonts for details.
package gofontwoff

//go:generate goexec -quiet "err := vfsgen.Generate(http.Dir(\"_data\"), vfsgen.Options{PackageName: \"gofontwoff\", VariableName: \"Assets\", VariableComment: \"Assets provides the Go font family WOFF data.\"}); if err != nil { log.Fatalln(err) }"
