package epub // import "github.com/Xunop/e-oasis/util/parsers/epub"

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
)

// Book is the main struct that holds all the information about the epub file
type Book struct {
	Ncx       Ncx       `json:"ncx"`
	Opf       Opf       `json:"opf"`
	Container Container `json:"container"`
	Mimetype  string    `json:"mimetype"`

	fd *zip.ReadCloser
}

// Open opens the epub file
func (p *Book) Open(file string) (io.ReadCloser, error) {
	return p.open(p.filename(file))
}

// Files returns a list of all the files in the epub
func (p *Book) Files() []string {
	var files []string
	for _, f := range p.fd.File {
		files = append(files, f.Name)
	}
	return files
}

// Close closes the epub file
func (p *Book) Close() error {
	return p.fd.Close()
}

// readXML reads the xml file with the given name and unmarshals it into the given interface
func (p *Book) readXML(n string, v interface{}) error {
	rc, err := p.open(n)
	if err != nil {
		return err
	}
	defer rc.Close()
	return xml.NewDecoder(rc).Decode(v)
}

// readBytes reads the file with the given name and returns its content as a byte slice
func (p *Book) readBytes(n string) ([]byte, error) {
    rc, err := p.open(n)
    if err != nil {
        return nil, err
    }
    defer rc.Close()
    return io.ReadAll(rc)
}

// filename returns the full path of the file
func (p *Book) filename(n string) string {
	return path.Join(path.Dir(p.Container.Rootfile.Fullpath), n)
}

// open opens the file with the given name
func (p *Book) open(n string) (io.ReadCloser, error) {
	for _, f := range p.fd.File {
		if f.Name == n {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", n)
}

func (p *Book) GetTitle() string {
	return p.Opf.Metadata.Title[0]
}

func (p *Book) GetAuthor() string {
	for _, author := range p.Opf.Metadata.Creator {
		if author.Role == "aut" {
			return author.Data
		} else if author.Role == "" {
			return author.Data
		}
	}
	return ""
}

func (p *Book) GetLanguage() string {
	return p.Opf.Metadata.Language[0]
}

func (p *Book) GetDescription() string {
	return p.Opf.Metadata.Description[0]
}

func (p *Book) GetPublisher() string {
	return p.Opf.Metadata.Publisher[0]
}

func (p *Book) GetISBN() string {
	for _, identifier := range p.Opf.Metadata.Identifier {
		if identifier.Scheme == "ISBN" {
			return identifier.Data
		} else if identifier.Scheme == "" {
			return identifier.Data // Fallback to default
		}
	}
	return ""
}

func (p *Book) GetUUID() string {
	for _, identifier := range p.Opf.Metadata.Identifier {
		if identifier.Scheme == "UUID" {
			return identifier.Data
		} else if identifier.Scheme == "" {
			return identifier.Data // Fallback to default
		}
	}
	return ""
}

// GetCover returns the path to the cover image
func (p *Book) GetCover() string {
	for _, meta := range p.Opf.Metadata.Meta {
		if meta.Name == "cover" {
			filename := meta.Content
			return p.filename(filename)
		}
	}
	return ""
}

func (p *Book) GetDate() string {
	return p.Opf.Metadata.Date[0].Data
}
