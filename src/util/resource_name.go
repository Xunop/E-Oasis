package util

import "regexp"

var (
	UIDMatcher = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9-]{1,30}[a-zA-Z0-9])$")
    TitleSortMatcher, _ = regexp.Compile(`^(A|The|An|Der|Die|Das|Den|Ein|Eine|Einen|Dem|Des|Einem|Eines|Le|La|Les|L\'|Un|Une)\s+`)
)
