package plural

import "strings"

func Plural(name string) string {
	// stolen from Harshita's frontend code
	if len(name) < 2 {
		return name
	}

	if strings.HasSuffix(name, "y") {
		return name[:len(name)-1] + "ies"
	}

	return name + "s"
}
