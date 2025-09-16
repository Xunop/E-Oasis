package epub

type Container struct {
	Rootfile Rootfile `xml:"rootfiles>rootfile" json:"rootfile"`
}

type Rootfile struct {
	Fullpath string `xml:"full-path,attr"`
	Type     string `xml:"media-type,attr"`
}
