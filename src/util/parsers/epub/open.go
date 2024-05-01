package epub

import (
	"archive/zip"
	"fmt"
)

func Open(f string) (*Book, error) {
	fd, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}

	b := &Book{fd: fd}
	m, err := b.readBytes("mimetype")
	if err != nil {
		return nil, err
	}
	b.Mimetype = string(m)
	if b.Mimetype != "application/epub+zip" {
		return nil, fmt.Errorf("epub: invalid mimetype: %s", b.Mimetype)
	}

	if err := b.readXML("META-INF/container.xml", &b.Container); err != nil {
		return nil, err
	}

	if err := b.readXML(b.Container.Rootfile.Fullpath, &b.Opf); err != nil {
		return nil, err
	}

	for _, mf := range b.Opf.Manifest {
		if mf.MediaType == "application/x-dtbncx+xml" {
			if err := b.readXML(b.filename(mf.Href), &b.Ncx); err != nil {
				return nil, err
			}
		}
	}

	return b, nil
}
