package git

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// DiffNameOnlyToFSInclude gets a Git diff based on name (`git diff --name-only`) and
// returns as a FS compatible `Include` paths options.
// This can be used to only load the files changed in a git diff.
func DiffNameOnlyToFSInclude(diff io.Reader) []string {
	paths := []string{}

	sc := bufio.NewScanner(diff)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		line = strings.ReplaceAll(line, "/", `\/`)
		pathRegex := fmt.Sprintf(`.*\/%s$`, line)
		paths = append(paths, pathRegex)
	}

	return paths
}
