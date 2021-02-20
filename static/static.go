package static

import (
	_ "embed" //nolint:gci
	"io/fs"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/embed"
)

//go:embed *
var static embed.FS //nolint:gochecknoglobals

func New(useEmbed bool) fs.ReadDirFS {
	if useEmbed {
		return static
	}
	return embed.DirFS("static")
}
