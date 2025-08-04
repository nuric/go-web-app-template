package static

import "embed"

/* By embedding the static files, we can serve them directly from the binary.
 * This simplifies the deployment process and avoids copying static files
 * separately. */

//go:embed *.css
var FS embed.FS
