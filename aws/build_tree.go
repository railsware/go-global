package aws

import (
	"strings"

	"github.com/railsware/go-global/v2/tree"
)

// Builds a tree of parameters out of the array
func buildParamTree(params []param) *tree.Node {
	paramTree := new(tree.Node)

	for _, param := range params {
		pathParts := strings.Split(param.path, "/")
		destination := paramTree
		for _, part := range pathParts {
			if destination.Children == nil {
				destination.Children = make(map[string]*tree.Node)
			}
			newDestination, ok := destination.Children[part]
			if !ok {
				newDestination = &tree.Node{}
				destination.Children[part] = newDestination
			}
			destination = newDestination
		}
		destination.Value = param.value
	}
	return paramTree
}
