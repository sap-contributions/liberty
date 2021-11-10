package main

import (
	"os"

	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/open-liberty/openliberty"
)

func main() {
	libpak.Main(
		openliberty.Detect{
			Logger: bard.NewLogger(os.Stdout),
		},
		openliberty.Build{
			Logger: bard.NewLogger(os.Stdout),
		},
	)
}
