package util

import "strings"

func StringToSlug(in string) string {
	id := in
	for _, old := range []string{"#"} {
		id = strings.ReplaceAll(id, old, "")
	}

	id = strings.Trim(id, " \t")
	for _, old := range []string{" ", "_", ".", "/"} {
		id = strings.ReplaceAll(id, old, "-")
	}

	return strings.ToLower(id)
}
