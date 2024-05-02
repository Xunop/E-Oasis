package store

import (
	"fmt"
	"strings"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"go.uber.org/zap"
)

func (s *Store) ListBooks(find *model.FindBook) ([]*model.Book, error) {
	where, args := []string{"1 = 1"}, []any{}

	if v := find.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := find.Title; v != nil {
		where, args = append(where, "title = ?"), append(args, *v)
	}
	if v := find.AuthorSort; v != nil {
		where, args = append(where, "author_sort = ?"), append(args, *v)
	}
	if v := find.ISBN; v != nil {
		where, args = append(where, "isbn = ?"), append(args, *v)
	}
	if v := find.LCCN; v != nil {
		where, args = append(where, "lccn = ?"), append(args, *v)
	}

	// Default order by title
	orderBy := []string{"title"}
	if find.OrderBy != nil {
		orderBy = append(orderBy, *find.OrderBy)
	}
	if find.Random {
		orderBy = []string{"RANDOM()"}
	}

	query := `
        SELECT
            id,
            title,
            sort,
            timestamp,
            pubdate,
            series_index,
            author_sort,
            isbn,
            lccn,
            path,
            flags,
            uuid,
            has_cover,
            last_modified
        FROM books
    WHERE ` + strings.Join(where, " AND ") + ` ORDER BY ` + strings.Join(orderBy, ", ")
	if v := find.Limit; v != nil {
		query += fmt.Sprintf(" LIMIT %d", *v)
	}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", query, args))

	rows, err := s.metaDb.Query(query, args...)
	if err != nil {
		log.Error("Failed to query books", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.Book, 0)
	for rows.Next() {
		var book model.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.SortTitle,
			&book.TimeStamp,
			&book.PublishDate,
			&book.SeriesIndex,
			&book.AuthorSort,
			&book.ISBN,
			&book.LCCN,
			&book.Path,
			&book.Flags,
			&book.UUID,
			&book.HasCover,
			&book.LastModified,
		); err != nil {
			log.Error("Failed to scan book", zap.Error(err))
			return nil, err
		}
		list = append(list, &book)
	}
	return list, nil
}

// TODO: Implement AddBook
func (s *Store) AddBook(book *model.Book) (*model.Book, error) {
	stmt := `
        INSERT INTO books (
            title,
            sort,
            pubdate,
            author_sort,
            isbn,
            path,
            uuid,
            has_cover,
            last_modified
        ) VALUES (?,?,?,?,?,?,?,?,?)
        RETURNING id, title, sort, pubdate, author_sort, isbn, path, uuid, has_cover, last_modified`
	args := []any{}

	args = append(args, book.Title)
	args = append(args, book.SortTitle)
	args = append(args, book.PublishDate)
	args = append(args, book.AuthorSort)
	args = append(args, book.ISBN)
	args = append(args, book.Path)
	args = append(args, book.UUID)
	args = append(args, book.HasCover)
	args = append(args, book.LastModified)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.metaDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newBook model.Book
	if err := tx.QueryRow(stmt, args...).Scan(
		&newBook.ID,
		&newBook.Title,
		&newBook.SortTitle,
		&newBook.PublishDate,
		&newBook.AuthorSort,
		&newBook.ISBN,
		&newBook.Path,
		&newBook.UUID,
		&newBook.HasCover,
		&newBook.LastModified,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newBook, nil
}

func (s *Store) AddPublisher(publisher *model.Publisher) (*model.Publisher, error) {
	exists, err := s.CheckPublisher(publisher.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return publisher, nil
	}

	stmt := `
        INSERT INTO publishers (
            name,
            sort
        ) VALUES (?, ?)
        RETURNING id, name, sort`
	args := []any{}

	args = append(args, publisher.Name)
	args = append(args, publisher.Sort)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.metaDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newPublisher model.Publisher
	if err := tx.QueryRow(stmt, args...).Scan(
		&newPublisher.ID,
		&newPublisher.Name,
		&newPublisher.Sort,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newPublisher, nil
}

func (s *Store) AddLanguage(code string) (*model.Language, error) {
	stmt := `
        INSERT INTO languages (
            lang_code,
        ) VALUES (?) RETURNING id, lang_code`
	args := []any{}

	args = append(args, code)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newLanguage model.Language
	if err := tx.QueryRow(stmt, args...).Scan(&newLanguage.ID, &newLanguage.LangCode); err != nil {
		return nil, err
	}
	return &newLanguage, nil
}

func (s *Store) CheckBook(title string) (bool, error) {
	stmt := `
		SELECT EXISTS(SELECT 1 FROM books WHERE title = ?)
	`
	args := []any{title}

	var exists bool
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) CheckPublisher(name string) (bool, error) {
	name = strings.TrimSpace(name)
	stmt := `
		SELECT EXISTS(SELECT 1 FROM publishers WHERE name = ?)
	`
	args := []any{name}

	var exists bool
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
