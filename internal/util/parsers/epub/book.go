package epub // import "github.com/Xunop/e-oasis/internal/util/parsers/epub"

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// currentDir returns the current dirctory
// func (p *Book) currentDir() string {
//
// }

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
	if p.Opf.Metadata.Title != nil {
		return p.Opf.Metadata.Title[0]
	}
	return ""
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
	if p.Opf.Metadata.Language != nil {
		return p.Opf.Metadata.Language[0]
	}
	return ""
}

func (p *Book) GetDescription() string {
	if p.Opf.Metadata.Description != nil {
		return p.Opf.Metadata.Description[0]
	}
	return ""
}

func (p *Book) GetPublisher() string {
	if p.Opf.Metadata.Publisher != nil {
		return p.Opf.Metadata.Publisher[0]
	}
	return ""
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

// GetCover copy the cover image to the destination and return the path
func (p *Book) GetCover(dest string) (string, error) {
	var filename string
	for _, meta := range p.Opf.Metadata.Meta {
		if meta.Name == "cover" {
			filename = meta.Content
		}
	}

	// If metadata can't find cover, from manifest
	if filepath.Ext(filename) == "" {
		for _, m := range p.Opf.Manifest {
			if m.ID == filename && strings.HasPrefix(m.MediaType, "image/") {
				filename = m.Href
			}
		}
	}

	if filename == "" {
		fmt.Printf("Can't find cover: %s", p.GetTitle())
		return "", nil
	}

	var fileDest string
	for _, f := range p.fd.File {
		if filepath.Base(f.Name) == filepath.Base(filename) {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			// Make sure the dir exists
			if os.Stat(filepath.Dir(dest)); os.IsNotExist(err) {
				return "", fmt.Errorf("Please make sure dirctory exist")
			}
			coverExt := filepath.Ext(filename)
			fileDest = filepath.Join(dest, "cover"+coverExt)
			outFile, err := os.Create(fileDest)
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return "", err
			}
		}
	}
	return fileDest, nil
}

func (p *Book) GetDate() string {
	if p.Opf.Metadata.Date != nil {
		return p.Opf.Metadata.Date[0].Data
	}
	return ""
}

// GetContent returns the given href content
func (p *Book) GetContent(href string) (string, error) {
	for _, m := range p.Opf.Manifest {
		if m.Href == href {
			rc, err := p.open(p.filename(m.Href))
			if err != nil {
				return "", err
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	return "", fmt.Errorf("content not found: %s", href)
}
