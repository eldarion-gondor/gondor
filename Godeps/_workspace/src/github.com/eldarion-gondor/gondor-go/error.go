package gondor

import (
	"fmt"
	"strings"
)

type APIError interface {
	Errors() []string
}

type ErrorList map[string][]string

type apiError struct {
	errList ErrorList
}

func (e apiError) Error() string {
	errs := e.Errors()
	if len(errs) == 0 {
		return "API error list is empty"
	} else if len(errs) == 1 {
		return errs[0]
	} else {
		delim := "\n\t * "
		return fmt.Sprintf("multiple issues reported:\n%s%s", delim, strings.Join(errs, delim))
	}
}

func (e apiError) Errors() []string {
	var res []string
	for key := range e.errList {
		for i := range e.errList[key] {
			var msg string
			if key == "non_field_errors" {
				msg = e.errList[key][i]
			} else {
				msg = fmt.Sprintf("%s: %s", key, e.errList[key][i])
			}
			res = append(res, msg)
		}
	}
	return res
}
