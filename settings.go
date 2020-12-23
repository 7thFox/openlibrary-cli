package main

import "flag"

type Settings struct {
	DefaultLookupKind string
	NWorkers          int
	Quiet             bool
	Color             bool
	Format            string
	SmartSpace        bool
}

var instance *Settings

const (
	formatMessage = ``
)

var (
	//inputFilename = flag.NewFlagSet()
	lookupKind = flag.String("n", "ISBN", "Default number provided for lookup without specifier")
	nWorkers   = flag.Int("w", 3, "Number of async workers")
	quiet      = flag.Bool("q", false, "Quiet - suppress error output")
	color      = flag.Bool("c", true, "Print errors with color")
	format     = flag.String("f", "[{classifications.lc_classifications.0}] {title} ({publish_date}) - {authors.*.name}", formatMessage)
	smartSpace = flag.Bool("s", true, "Add spaces around bindings 'as-needed'")
)

func GetSettings() *Settings {
	if instance == nil {
		flag.Parse()
		s := Settings{
			DefaultLookupKind: *lookupKind,
			NWorkers:          *nWorkers,
			Quiet:             *quiet,
			Color:             *color,
			Format:            *format,
			SmartSpace:        *smartSpace,
		}
		instance = &s
	}
	return instance
}
