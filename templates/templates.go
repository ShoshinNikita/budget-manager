package templates

import (
	_ "embed" //nolint:gci
	"io/fs"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/embed"
)

//go:embed *
var templates embed.FS //nolint:gochecknoglobals

func New(useEmbed bool) fs.ReadDirFS {
	if useEmbed {
		return templates
	}
	return embed.DirFS("templates")
}
