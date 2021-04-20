package horzmerge_test

import (
	"embed"
	"io/fs"
)

//go:embed fixtures
var fixtureRootFS embed.FS
var fixtureFS, _ = fs.Sub(fixtureRootFS, "fixtures")

// This example show how to use horzmerge.Func()
func ExampleFunc() {
	//fmt.Println(horzmerge.Func())
	// Output: 42
}
