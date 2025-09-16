package epub

// Ncx is the struct that holds the information from the ncx file
type Ncx struct {
    Points []Point `xml:"navMap>navPoint" json:"points"`
}

// Point is the struct that holds the information about a point in the ncx file
type Point struct {
    Text string `xml:"navLabel>text" json:"text"`
    Content string `xml:"content" json:"content"`
    Points []Point `xml:"navPoint" json:"points"`
}

// Content is the struct that holds the information about the content of a point in the ncx file
type Content struct {
    Src string `xml:"src,attr" json:"src"`
}
