package main

import (
	"fmt"
	"strings"
)

var addRemoveAccessScript string

// change it to a map and then everything falls into place
// eventually save this script so that it can be accessed more easily
func initAddRemoveAccessScript() {
	// update all categories and tags at the same time.
	addRemoveAccessScript = strings.ReplaceAll(fmt.Sprintf(`
	for (int i = 0; i < access.length; i++) {
		bool cont = true;
		if (access[i].type != null) {
			if (access[i].type == '%s') {
				if (ctx._source.access[access[i].id] != null) {
					ctx._source.access[access[i].id].remove(access[i].id);
				}
				cont = false;
			} else {
				if (ctx._source.access[access[i].id] != null) {
					ctx._source.access[access[i].id].type = access[i].type;
				} else {
					ctx._source.access[access[i].id] = {
						'type': access[i].type
					}
				}
			}
		}
		if (cont) {
			if (access[userIDString].categories != null) {
				ctx._source.access[userIDString].categories = categories
			}
			if (access[userIDString].tags != null) {
				ctx._source.access[userIDString].tags = tags
			}
		}
	}
	`, noAccessLevel), "\n", "")
}
