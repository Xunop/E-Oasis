package store

import (
	"fmt"

	"github.com/Xunop/e-oasis/internal/log"
	"github.com/Xunop/e-oasis/internal/model"
)

func (s *Store) AddAuthor(author *model.Author) (*model.Author, error) {
	// Get Author ID if author is exist
	// FIXME: Author may have same name
	if aID, ok := s.CheckAuthor(author.Name); ok {
		author.ID = aID
		return author, nil
	}
	stmt := `
	    INSERT INTO authors (
	    name, sort, link
	    ) VALUES (?,?,?) RETURNING id,name,sort,link`
	args := []any{}

	args = append(args, author.Name)
	args = append(args, author.Sort)
	args = append(args, author.Link)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()

	tx, err := s.metaDb.Begin()
	if err != nil {
		return nil, err
	}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newAuthor model.Author
	if err := tx.QueryRow(stmt, args...).Scan(&newAuthor.ID, &newAuthor.Name, &newAuthor.Sort, &newAuthor.Link); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newAuthor, nil
}

func (s *Store) UpdateAuthor(author *model.Author) (*model.Author, error) {
    stmt := `
        UPDATE authors SET
        name = ?, sort = ?, link = ?
        WHERE id = ? RETURNING id,name,sort,link`
    args := []any{}

    args = append(args, author.Name)
    args = append(args, author.Sort)
    args = append(args, author.Link)
    args = append(args, author.ID)

    s.metaDbLock.Lock()
    defer s.metaDbLock.Unlock()

    tx, err := s.metaDb.Begin()
    if err != nil {
        return nil, err
    }

    log.Debug("SQL query and args:")
    log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

    var newAuthor model.Author
    if err := tx.QueryRow(stmt, args...).Scan(&newAuthor.ID, &newAuthor.Name, &newAuthor.Sort, &newAuthor.Link); err != nil {
        tx.Rollback()
        return nil, err
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }
    return &newAuthor, nil
}

func (s *Store) AddBookAuthorLink(link *model.BookAuthorLink) (*model.BookAuthorLink, error) {
	stmt := `
		INSERT INTO books_authors_link (
			book,
			author
		) VALUES (?,?)
	    RETURNING id, book, author
	`
	args := []any{}

	args = append(args, link.BookID)
	args = append(args, link.AuthorID)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.metaDb.Begin()
	if err != nil {
		return nil, err
	}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	authorLink := model.BookAuthorLink{}
	if err := tx.QueryRow(stmt, args...).Scan(&authorLink.ID, &authorLink.BookID, &authorLink.AuthorID); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &authorLink, nil
}
