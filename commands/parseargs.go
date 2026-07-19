package commands

import (
	"fmt"
	"strings"

	"deploycli/internal/static"
)

// parsedArgs is the result of splitting a command's argv into positional
// arguments and deploycli's flags. A hand-rolled parser is used instead of
// the stdlib flag package because deploycli's flags (--image, --port,
// --json) can appear *after* positional arguments, e.g.:
//
//	deploycli send user@ip:/path --image api:latest --port 2222
//
// and flag.Parse stops at the first non-flag argument, which would break on
// input like that.
type parsedArgs struct {
	positional []string
	images     []string
	port       string
	json       bool
}

func parseArgs(args []string) (parsedArgs, error) {
	p := parsedArgs{port: static.DefaultSSHPort}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			p.json = true
		case a == "--image":
			if i+1 >= len(args) {
				return parsedArgs{}, fmt.Errorf("--image requires a value")
			}
			p.images = append(p.images, args[i+1])
			i++
		case strings.HasPrefix(a, "--image="):
			val := strings.TrimPrefix(a, "--image=")
			if val == "" {
				return parsedArgs{}, fmt.Errorf("--image= requires a value")
			}
			p.images = append(p.images, val)
		case a == "--port":
			if i+1 >= len(args) {
				return parsedArgs{}, fmt.Errorf("--port requires a value")
			}
			p.port = args[i+1]
			i++
		case strings.HasPrefix(a, "--port="):
			p.port = strings.TrimPrefix(a, "--port=")
		default:
			p.positional = append(p.positional, a)
		}
	}
	return p, nil
}
