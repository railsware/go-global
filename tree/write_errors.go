package tree

import (
	"fmt"
	"strings"

	"github.com/railsware/go-global/v2"
)

type writeError struct {
	msg         string
	path        string
	isPathError bool
}

type WriteErrors struct {
	errors []writeError
}

func newWriteErrors(msg string, isPathError bool) WriteErrors {
	return WriteErrors{[]writeError{{msg: msg, isPathError: isPathError}}}
}

func (we *WriteErrors) Present() bool {
	return len(we.errors) > 0
}

func (we *WriteErrors) append(err writeError) {
	we.errors = append(we.errors, err)
}

func (we *WriteErrors) merge(newErrors WriteErrors) {
	we.errors = append(we.errors, newErrors.errors...)
}

func (we *WriteErrors) mergeChildErrors(childName string, childErrors WriteErrors) {
	for _, childErr := range childErrors.errors {
		if childErr.path == "" {
			childErr.path = childName
		} else {
			childErr.path = fmt.Sprintf("%s/%s", childName, childErr.path)
		}
		we.errors = append(we.errors, childErr)
	}
}

func (we *WriteErrors) Join() global.Error {
	var msgs []string
	isWarning := true
	for _, err := range we.errors {
		msg := fmt.Sprintf("%s: %s", err.path, err.msg)
		msgs = append(msgs, msg)
		if !err.isPathError {
			isWarning = false
		}
	}

	msg := fmt.Sprintf("global: %s", strings.Join(msgs, ", "))

	if isWarning {
		return global.NewWarning(msg)
	} else {
		return global.NewError(msg)
	}
}
