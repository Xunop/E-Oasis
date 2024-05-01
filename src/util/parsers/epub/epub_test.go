package epub

import (
	"log"
	"os"
	"testing"

	epub2 "github.com/go-shiori/go-epub"
)

func createEpub(n string) error {
	// Create a new EPUB
	e, err := epub2.NewEpub("Test title")
	if err != nil {
		log.Println(err)
	}

	// Set the author
	e.SetAuthor("Test author")

	// Add a section
	// section1Body := `<h1>Section 1</h1>
	//    <p>This is a paragraph.</p>`
	// section1Path, err := e.AddSection(section1Body, "Section 1", "", "")
	// if err != nil {
	// 	log.Println(err)
	// }

	// Write the EPUB
	err = e.Write(n)
	if err != nil {
		return err
	}
	return nil
}

func TestEpub(t *testing.T) {
	withBook := func(f string, fn func(*Book)) {
		err := createEpub(f)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Open(f)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()
		fn(b)
		os.Remove(f)
	}

	t.Run("TestOpen", func(t *testing.T) {
		withBook("test.epub", func(b *Book) {
			if b.Mimetype != "application/epub+zip" {
				t.Errorf("invalid mimetype: %s", b.Mimetype)
			}
			if b.Container.Rootfile.Fullpath != "EPUB/package.opf" {
				t.Errorf("invalid rootfile: %s", b.Container.Rootfile.Fullpath)
			}
		})
	})

	t.Run("TestFiles", func(t *testing.T) {
		withBook("test.epub", func(b *Book) {
			files := b.Files()
			if len(files) != 5 {
				t.Errorf("expected 5 files, got %d", len(files))
			}
		})
	})

	t.Run("TestReadXML", func(t *testing.T) {
		withBook("test.epub", func(b *Book) {
			var opf Opf
			err := b.readXML(b.Container.Rootfile.Fullpath, &opf)
			if err != nil {
				t.Fatalf("failed to read XML: %v", err)
			}
			if len(opf.Manifest) != 2 {
				t.Errorf("expected 2 manifest items, got %d", len(opf.Manifest))
			}
		})
	})

    t.Run("TestBookMetadata", func(t *testing.T) {
        withBook("test.epub", func(b *Book) {
            if b.Opf.Metadata.Title[0] != "Test title" {
                t.Errorf("expected title 'Test title', got '%s'", b.Opf.Metadata.Title[0])
            }
            if b.Opf.Metadata.Creator[0].Data != "Test author" {
                t.Errorf("expected author 'Test author', got '%s'", b.Opf.Metadata.Creator[0].Data)
            }
        })
    })
}
