package horzmerge_test

import (
	"embed"
	"io/fs"

	_ "github.com/parrogo/horzmerge"
)

//go:embed fixtures
var fixtureRootFS embed.FS
var fixtureFS, _ = fs.Sub(fixtureRootFS, "fixtures")

// This example show how to use horzmerge.Func()
func ExampleMerge() {
	//fmt.Println(horzmerge.Merge())
	// Output: 42
}
