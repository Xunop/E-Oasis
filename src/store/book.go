package store

import (
	"fmt"
	"strings"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"go.uber.org/zap"
)

func (s *Store) GetBook(find *model.FindBook) (*model.Book, error) {
	if find.ID != nil {
		if cache, ok := s.BookCache.Load(*find.ID); ok {
			return cache.(*model.Book), nil
		}
	}

	list, err := s.ListBooks(find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	book := list[0]
	s.BookCache.Store(book.ID, book)
	return book, nil
}

// TODO: RemoveBook need to remove link between user and book
// RemoveBook remove book from the store
func (s *Store) RemoveBook(find *model.FindBook) error {
	if find.UserID != nil {
		return s.RemoveBookByUserID(*find.UserID, *find.BookID)
	}
	stmt := `DELETE FROM books WHERE id = ?`
	args := []any{find.BookID}

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.metaDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	if _, err := tx.Exec(stmt, args...); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ListBooks(find *model.FindBook) ([]*model.Book, error) {
	if v := find.UserID; v != nil {
		return s.ListBooksByUserID(*v)
	}

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

func (s *Store) ListBooksByUserID(userID int) ([]*model.Book, error) {
	stmt := `
		SELECT
	        book_id
	    FROM book_user_link
	    WHERE user_id = ?
	`
	args := []any{userID}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	rows, err := s.appDb.Query(stmt, args...)
	if err != nil {
		log.Error("Failed to query books by user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	list := make([]string, 0)

	for rows.Next() {
		var bookID int
		if err := rows.Scan(&bookID); err != nil {
			log.Error("Failed to scan book ID", zap.Error(err))
			return nil, err
		}
		list = append(list, fmt.Sprintf("%d", bookID))
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
        WHERE id IN` + `(` + strings.Join(list, ",") + `)`

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", query, args))

	rows, err = s.metaDb.Query(query)
	if err != nil {
		log.Error("Failed to query books", zap.Error(err))
		return nil, err
	}

	listBooks := make([]*model.Book, 0)
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
		listBooks = append(listBooks, &book)
	}
	return listBooks, nil
}

func (s *Store) RemoveBookByUserID(userID int, bookID ...int) error {
	where, args := []string{"1 = 1"}, []any{}

	where, args = append(where, "user_id = ?"), append(args, userID)
	bookIDs := make([]string, 0)

	for _, id := range bookID {
		bookIDs = append(bookIDs, fmt.Sprintf("%d", id))
	}

	stmt := `
		DELETE FROM book_user_link
		WHERE user_id = ? AND book_id IN (` + strings.Join(bookIDs, ",") + `)
	    RETURNING book_id
	`

	s.appDbLock.Lock()
	defer s.appDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	rows, err := tx.Query(stmt, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	rmList := make([]int, 0)
	for rows.Next() {
		var bookID int
		if err := rows.Scan(&bookID); err != nil {
			return err
		}
		rmList = append(rmList, bookID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if len(rmList) == 0 {
		return nil
	}

	where, args = []string{"1 = 1"}, []any{}
	where, args = append(where, "id IN ("+strings.Join(bookIDs, ",")+")"), append(args, rmList)

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err = s.metaDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt = `
		DELETE FROM books
		WHERE ` + strings.Join(where, " AND ")
	if _, err = tx.Exec(stmt, args...); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) AddBook(book *model.Book) (*model.Book, error) {
	stmt := `
        INSERT INTO books (
            title,
            sort,
            pubdate,
            author_sort,
            isbn,
            lccn,
            path,
            uuid,
            has_cover,
            last_modified
        ) VALUES (?,?,?,?,?,?,?,?,?,?)
        RETURNING id, title, sort, pubdate, author_sort, isbn, lccn, path, uuid, has_cover, last_modified`
	args := []any{}

	args = append(args, book.Title)
	args = append(args, book.SortTitle)
	args = append(args, book.PublishDate)
	args = append(args, book.AuthorSort)
	args = append(args, book.ISBN)
	args = append(args, book.LCCN)
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
		&newBook.LCCN,
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
	if pID, ok := s.CheckPublisher(publisher.Name); ok {
		publisher.ID = pID
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

func (s *Store) AddAuthor(author *model.Author) (*model.Author, error) {
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
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newAuthor model.Author
	if err := tx.QueryRow(stmt, args...).Scan(&newAuthor.ID, &newAuthor.Name, &newAuthor.Sort, &newAuthor.Link); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newAuthor, nil
}

func (s *Store) AddBookUserLink(create *model.BookUserLink) (*model.BookUserLink, error) {
	stmt := `
		INSERT INTO book_user_link (
			book_id,
			user_id
		) VALUES (?,?)
		RETURNING id, book_id, user_id`
	args := []any{}

	args = append(args, create.BookID)
	args = append(args, create.UserID)

	s.appDbLock.Lock()
	defer s.appDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newLink model.BookUserLink
	if err := tx.QueryRow(stmt, args...).Scan(&newLink.ID, &newLink.BookID, &newLink.UserID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newLink, nil
}

// SetBookStatus set or update book reading status
func (s *Store) SetBookStatus(status *model.BookReadingStatusLink) (*model.BookReadingStatusLink, error) {
	// Insert or update
	stmt := `
	INSERT INTO reading_status (
	         book_id,
	         user_id,
	         last_read_time,
	         cur_page,
	         percentage,
	         duration,
	         page,
	         status
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(book_id, user_id) DO UPDATE
	SET
		last_read_time = EXCLUDED.last_read_time,
		cur_page = EXCLUDED.cur_page,
	    percentage = EXCLUDED.percentage,
	    duration = EXCLUDED.duration,
		status = EXCLUDED.status
	`
	args := []any{
		status.BookID,
		status.UserID,
		status.LastRead,
		status.CurPage,
		status.Percentage,
		status.Duration,
		status.Page,
		status.Status,
	}

	tx, err := s.appDb.Begin()
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback()

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	if _, err := tx.Exec(stmt, args...); err != nil {
		log.Error("Failed to scan book reading status link", zap.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return status, nil
}

func (s *Store) CheckBook(bookID int) bool {
	stmt := `
		SELECT EXISTS(SELECT 1 FROM books WHERE id = ?)
	`
	args := []any{bookID}

	var exists bool
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func (s *Store) CheckAuthor(name string) (int, bool) {
	stmt := `SELECT id FROM authors WHERE name = ?`
	args := []any{name}

	var authorID int
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&authorID); err != nil {
		return 0, false
	}
	return authorID, true
}

func (s *Store) CheckPublisher(name string) (int, bool) {
	stmt := `SELECT id FROM publishers WHERE name = ?`
	args := []any{name}

	var publisherID int
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&publisherID); err != nil {
		return 0, false
	}
	return publisherID, true
}

func (s *Store) CheckBookHash(hash string) (int, bool) {
	stmt := `SELECT book_id FROM book_hash_link WHERE hash = ?`
	args := []any{hash}

	var bookID int
	if err := s.metaDb.QueryRow(stmt, args...).Scan(&bookID); err != nil {
		return 0, false
	}
	return bookID, true
}

func (s *Store) CheckBookStatus(bookID, userID int) bool {
	stmt := `
	    SELECT EXISTS(SELECT 1 FROM book_reading_status_link WHERE book_id = ? AND user_id = ?)
	`
	args := []any{bookID, userID}

	var exists bool
	if err := s.appDb.QueryRow(stmt, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
