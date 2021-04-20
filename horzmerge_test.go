package horzmerge

import (
	"bufio"
	"bytes"
	"embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures
var fixtureRootFS embed.FS
var fixtureFS, _ = fs.Sub(fixtureRootFS, "fixtures")

func TestMerge(t *testing.T) {
	t.Run("checkHeaders", func(t *testing.T) {
		require := require.New(t)

		source1, err := fixtureFS.Open("andre")
		require.NoError(err)

		err = checkHeaders(bufio.NewReader(source1), []string{
			"  name", " gender", " age", " f1", " f2", " f3",
		})
		require.NoError(err)

		source1.Close()
		source1, err = fixtureFS.Open("andre")
		require.NoError(err)

		err = checkHeaders(bufio.NewReader(source1), []string{
			"  name", "  gender", " age", " f1", " f2", " f3",
		})
		require.EqualError(err, "field header 1 differs: expected `  gender`, got ` gender`")

		source1.Close()
		source1, err = fixtureFS.Open("andre")
		require.NoError(err)

		err = checkHeaders(bufio.NewReader(source1), []string{
			" age", " f1", " f2", " f3",
		})
		require.EqualError(err, "headers len differs: expected 4, got 6")

	})
	t.Run("readHeaders", func(t *testing.T) {
		require := require.New(t)

		source1, err := fixtureFS.Open("andre")
		require.NoError(err)

		headers, err := readHeaders(bufio.NewReader(source1))
		require.NoError(err)
		require.Equal([]string{
			"  name", " gender", " age", " f1", " f2", " f3",
		}, headers)
	})

	t.Run("Merge", func(t *testing.T) {
		require := require.New(t)

		source1, err := fixtureFS.Open("andre")
		require.NoError(err)

		source2, err := fixtureFS.Open("tati")
		require.NoError(err)
		var buf bytes.Buffer

		Merge(Options{
			Empty:  "a",
			Target: &buf,
		}, source1, source2)

		require.NoError(err)
		actual := buf.String()
		expected :=
			/*    */ "  name gender age f1 f2 f3\n" +
				/**/ " andre      m  45  X  b  c\n"

		require.Equal(expected, actual)
	})

}
