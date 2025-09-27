package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Xunop/e-oasis/internal/log"
	"github.com/Xunop/e-oasis/internal/model"
	"github.com/Xunop/e-oasis/internal/util"
	"github.com/Xunop/e-oasis/internal/util/parsers/epub"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Store) GetBook(find *model.FindBook) (*model.Book, error) {
	if find.BookID != nil {
		if cache, ok := s.BookCache.Load(*find.BookID); ok {
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

// RemoveBook removes book records from the database AND its corresponding file from storage.
func (s *Store) RemoveBook(find *model.FindBook) error {
	// Find the book to get its path before deleting the record.
	// We use ListBooks because it can handle all find criteria.
	books, err := s.ListBooks(find)
	if err != nil {
		return errors.Wrap(err, "failed to find book for deletion")
	}
	if len(books) == 0 {
		log.Warn("Attempted to delete a book that does not exist or does not match criteria", zap.Any("find", find))
		return nil // This is not an error, just nothing to do.
	}

	// Assuming we only delete the first book that matches the criteria.
	bookToDelete := books[0]
	bookID := bookToDelete.ID
	bookPath := bookToDelete.Path

	// Delete database records in a transaction.
	// metaDb transaction
	metaTx, err := s.metaDb.Begin()
	if err != nil {
		return err
	}
	defer metaTx.Rollback() // Rollback on error

	// The AFTER DELETE trigger on the 'books' table will handle cascading deletes
	// for most linked data in metaDb. We just need to delete from the books table.
	if _, err := metaTx.Exec("DELETE FROM books WHERE id = ?", bookID); err != nil {
		return errors.Wrap(err, "failed to delete from books table")
	}

	// appDb transaction
	appTx, err := s.appDb.Begin()
	if err != nil {
		return err
	}
	defer appTx.Rollback()

    // Delete all links to the book from the application database tables.
	tablesToClean := []string{
		"book_user_link",
		"book_shelf_link",
		"bookmark",
		"duration_info",
		"reading_status",
		"book_hash_link",
	}

	for _, table := range tablesToClean {
		stmt := fmt.Sprintf("DELETE FROM %s WHERE book_id = ?", table)
		if _, err := appTx.Exec(stmt, bookID); err != nil {
			return errors.Wrapf(err, "failed to delete from %s table", table)
		}
	}

	// Commit transactions for both databases.
	if err := metaTx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit metaDb transaction")
	}
	if err := appTx.Commit(); err != nil {
		// Note: At this point, metaDb is committed. If appDb fails to commit,
		// the data will be in an inconsistent state. A more complex system might require
		// a two-phase commit or compensation logic, but for now, we'll just return the error.
		return errors.Wrap(err, "failed to commit appDb transaction")
	}

	// If database operations were successful, delete the physical file.
	if bookPath != "" {
		// The book file is stored in its own directory, so we remove the entire directory.
		bookDir := filepath.Dir(bookPath)
		log.Debug("Deleting book directory", zap.String("path", bookDir))
		if err := os.RemoveAll(bookDir); err != nil {
			// This is a critical error as it results in an orphaned file. It needs to be logged.
			log.Error("CRITICAL: Failed to delete book directory after deleting DB record",
				zap.String("path", bookDir),
				zap.Error(err))
			return errors.Wrap(err, "failed to delete book directory")
		}
	}

	log.Info("Book deleted successfully", zap.Int("bookID", bookID))
	return nil
}

func (s *Store) ListBooks(find *model.FindBook) ([]*model.Book, error) {
	if v := find.UserID; v != nil {
		return s.ListBooksByUserID(*v)
	}

	where, args := []string{"1 = 1"}, []any{}

	if v := find.BookID; v != nil {
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

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	rows, err := tx.Query(stmt, args...)
	if err != nil {
		tx.Rollback()
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

	stmt = `
		DELETE FROM books
		WHERE ` + strings.Join(where, " AND ")
	if _, err = tx.Exec(stmt, args...); err != nil {
		tx.Rollback()
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
		tx.Rollback()
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

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newPublisher model.Publisher
	if err := tx.QueryRow(stmt, args...).Scan(
		&newPublisher.ID,
		&newPublisher.Name,
		&newPublisher.Sort,
	); err != nil {
		tx.Rollback()
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

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newLanguage model.Language
	if err := tx.QueryRow(stmt, args...).Scan(&newLanguage.ID, &newLanguage.LangCode); err != nil {
		tx.Rollback()
		return nil, err
	}
	return &newLanguage, nil
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

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var newLink model.BookUserLink
	if err := tx.QueryRow(stmt, args...).Scan(&newLink.ID, &newLink.BookID, &newLink.UserID); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newLink, nil
}

// UpsetBookStatus upset book status
// If the book status is not exist, insert new record
// If the book status is exist, update the record, except for book_id, user_id, page
func (s *Store) UpsetBookStatus(status *model.BookReadingStatusLink) (*model.BookReadingStatusLink, error) {
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

	s.appDbLock.Lock()
	defer s.appDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		log.Error("Failed to begin transaction", zap.Error(err))
		return nil, err
	}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	if _, err := tx.Exec(stmt, args...); err != nil {
		tx.Rollback()
		log.Error("Failed to scan book reading status link", zap.Error(err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// FIXME: status.ID is not set
	status.ID = -1
	return status, nil
}

func (s *Store) GetBookStatus(bookID, userID int) (*model.BookReadingStatusLink, error) {
	stmt := `
   		SELECT
   	        id,
   	        book_id,
   	        user_id,
   	        last_read_time,
   	        duration,
   	        cur_page,
   	        percentage,
   	        status,
   	        page
   	    FROM reading_status WHERE book_id = ? AND user_id = ?
	`
	args := []any{bookID, userID}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	var status model.BookReadingStatusLink
	if err := s.appDb.QueryRow(stmt, args...).Scan(
		&status.ID,
		&status.BookID,
		&status.UserID,
		&status.LastRead,
		&status.Duration,
		&status.CurPage,
		&status.Percentage,
		&status.Status,
		&status.Page,
	); err != nil {
		return nil, err
	}
	return &status, nil
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

// AddBookHashLink creates a link between a book ID and its unique content hash.
func (s *Store) AddBookHashLink(bookID int, hash string) error {
	stmt := `INSERT INTO book_hash_link (book_id, hash) VALUES (?, ?)`
	args := []any{bookID, hash}

	s.appDbLock.Lock()
	defer s.appDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		return err
	}

	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", stmt, args))

	if _, err := tx.Exec(stmt, args...); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Store) CheckBookHash(hash string) (int, bool) {
	stmt := `SELECT book_id FROM book_hash_link WHERE hash = ?`
	args := []any{hash}

	var bookID int
	if err := s.appDb.QueryRow(stmt, args...).Scan(&bookID); err != nil {
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

// ListAllTags retrieves all unique tags and their book counts.
func (s *Store) ListAllTags() ([]*model.Tag, error) {
	query := `
		SELECT
			t.id,
			t.name,
			COUNT(btl.book) as count
		FROM tags t
		JOIN books_tags_link btl ON t.id = btl.tag
		GROUP BY t.id, t.name
		ORDER BY t.name
	`
	rows, err := s.metaDb.Query(query)
	if err != nil {
		log.Error("Failed to query all tags", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.BookCount); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

// ListBooksByTag retrieves all books associated with a specific tag ID.
func (s *Store) ListBooksByTag(tagID int) ([]*model.Book, error) {
	query := `
		SELECT
			b.id, b.title, b.sort, b.timestamp, b.pubdate, b.series_index,
			b.author_sort, b.isbn, b.lccn, b.path, b.flags, b.uuid,
			b.has_cover, b.last_modified
		FROM books b
		JOIN books_tags_link btl ON b.id = btl.book
		WHERE btl.tag = ?
		ORDER BY b.sort
	`
	rows, err := s.metaDb.Query(query, tagID)
	if err != nil {
		log.Error("Failed to query books by tag", zap.Int("tagID", tagID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var books []*model.Book
	for rows.Next() {
		var book model.Book
		if err := rows.Scan(
			&book.ID, &book.Title, &book.SortTitle, &book.TimeStamp, &book.PublishDate,
			&book.SeriesIndex, &book.AuthorSort, &book.ISBN, &book.LCCN, &book.Path,
			&book.Flags, &book.UUID, &book.HasCover, &book.LastModified,
		); err != nil {
			return nil, err
		}
		books = append(books, &book)
	}
	return books, nil
}

// AddTagToBook associates a tag with a book.
// It finds the tag by name or creates it if it doesn't exist,
// then creates the link between the book and the tag.
// This is done in a transaction to ensure atomicity.
func (s *Store) AddTagToBook(bookID int, tagName string) error {
	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	tx, err := s.metaDb.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	// Defer a rollback in case of an error, it will be ignored if Commit() is called.
	defer tx.Rollback()

	// Find the tag's ID, or create the tag if it doesn't exist.
	tagID, err := s.findOrCreateTagTx(tx, tagName)
	if err != nil {
		return errors.Wrap(err, "failed to find or create tag")
	}

	// Create the link between the book and the tag.
	// "INSERT OR IGNORE" is used to silently fail if the link already exists,
	// preventing duplicate entry errors, which is desired behavior.
	stmt := `INSERT OR IGNORE INTO books_tags_link (book, tag) VALUES (?, ?)`
	if _, err := tx.Exec(stmt, bookID, tagID); err != nil {
		return errors.Wrap(err, "failed to insert into books_tags_link")
	}

	return tx.Commit()
}

// findOrCreateTagTx finds a tag by name or creates it if it doesn't exist.
// This helper must be called within an existing transaction.
func (s *Store) findOrCreateTagTx(tx *sql.Tx, tagName string) (int, error) {
	var tagID int

	// try to find the existing tag.
	query := `SELECT id FROM tags WHERE name = ?`
	err := tx.QueryRow(query, tagName).Scan(&tagID)

	if err != nil {
		// If no rows were found, the tag doesn't exist. Create it.
		if errors.Is(err, sql.ErrNoRows) {
			insertStmt := `INSERT INTO tags (name) VALUES (?) RETURNING id`
			if err := tx.QueryRow(insertStmt, tagName).Scan(&tagID); err != nil {
				return 0, errors.Wrap(err, "failed to insert new tag")
			}
			return tagID, nil
		}
		// For any other error, return it.
		return 0, errors.Wrap(err, "failed to query for existing tag")
	}

	// If we found the tag, return its ID.
	return tagID, nil
}

// ParseAndSaveBookMeta takes a file path, parses it, and saves the book and all
// its related metadata (author, publisher, links) in a single transaction.
func (s *Store) ParseAndSaveBookMeta(path string, userID int, tags []string) error {
	// Parse the book file to get its metadata.
	book, err := epub.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open epub for parsing")
	}
	defer book.Close()

	hasCover := false
	bookCover, _ := book.GetCover(filepath.Dir(path))
	if bookCover != "" {
		hasCover = true
		// Handle cover conversion in a separate goroutine so it doesn't slow down the import
		go func(coverPath string) {
			webpPath := util.ImageToWebp(coverPath, 75)
			if webpPath != "" { // If conversion succeeds, remove original
				os.Remove(coverPath)
			}
		}(bookCover)
	}

	s.metaDbLock.Lock()
	defer s.metaDbLock.Unlock()
	// Begin a database transaction.
	tx, err := s.metaDb.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback() // Rollback on any error

	// Create or find the author within the transaction.
	authorName := book.GetAuthor()
	authorSort := util.AuthorSort(authorName)
	authorID, err := s.findOrCreateAuthorTx(tx, authorName, authorSort)
	if err != nil {
		return err
	}

	// Create or find the publisher within the transaction.
	publisherName := book.GetPublisher()
	if strings.TrimSpace(publisherName) == "" {
		publisherName = "Unknown"
	}
	publisherID, err := s.findOrCreatePublisherTx(tx, publisherName)
	if err != nil {
		return err
	}

	// Insert the book record.
	bookToCreate := &model.Book{
		Title:        book.GetTitle(),
		SortTitle:    util.TitleSort(book.GetTitle()),
		PublishDate:  book.GetDate(),
		AuthorSort:   authorSort,
		ISBN:         book.GetISBN(),
		Path:         path,
		UUID:         book.GetUUID(),
		HasCover:     hasCover,
		LastModified: time.Now().UTC().Format(time.RFC3339),
	}

	bookInsertStmt := `INSERT INTO books (title, sort, pubdate, author_sort, isbn, path, uuid, has_cover, last_modified) 
					   VALUES (?,?,?,?,?,?,?,?,?) RETURNING id`
	var bookID int
	err = tx.QueryRow(bookInsertStmt, bookToCreate.Title, bookToCreate.SortTitle, bookToCreate.PublishDate,
		bookToCreate.AuthorSort, bookToCreate.ISBN, bookToCreate.Path, bookToCreate.UUID,
		bookToCreate.HasCover, bookToCreate.LastModified).Scan(&bookID)
	if err != nil {
		return errors.Wrap(err, "failed to insert book record")
	}

	// Link book to author and publisher.
	_, err = tx.Exec(`INSERT OR IGNORE INTO books_authors_link (book, author) VALUES (?, ?)`, bookID, authorID)
	if err != nil {
		return errors.Wrap(err, "failed to link book to author")
	}

	_, err = tx.Exec(`INSERT OR IGNORE INTO books_publishers_link (book, publisher) VALUES (?, ?)`, bookID, publisherID)
	if err != nil {
		return errors.Wrap(err, "failed to link book to publisher")
	}

	if len(tags) > 0 {
		for _, tagName := range tags {
			// Sanitize tag name
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Find or create the tag within the transaction
			tagID, err := s.findOrCreateTagTx(tx, tagName)
			if err != nil {
				// If one tag fails, the whole transaction for this book will be rolled back.
				return errors.Wrapf(err, "failed to find or create tag '%s'", tagName)
			}

			// Link the book to the tag
			_, err = tx.Exec(`INSERT OR IGNORE INTO books_tags_link (book, tag) VALUES (?, ?)`, bookID, tagID)
			if err != nil {
				return errors.Wrapf(err, "failed to link book to tag '%s'", tagName)
			}
		}
	}

	// Link the book to the user in the other database.
	// This can't be in the same transaction, but should happen before we commit.
	if _, err := s.AddBookUserLink(&model.BookUserLink{BookID: bookID, UserID: userID}); err != nil {
		return errors.Wrap(err, "failed to link book to user")
	}

	// Commit the transaction.
	log.Debug("Successfully imported and saved metadata", zap.String("book", book.GetTitle()))
	return tx.Commit()
}

// Helper function for finding/creating authors within a transaction.
func (s *Store) findOrCreateAuthorTx(tx *sql.Tx, name, sort string) (int, error) {
	var authorID int
	err := tx.QueryRow(`SELECT id FROM authors WHERE name = ?`, name).Scan(&authorID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`INSERT INTO authors (name, sort, link) VALUES (?, ?, ?) RETURNING id`, name, sort, "").Scan(&authorID)
	}
	return authorID, err
}

// Helper function for finding/creating publishers within a transaction.
func (s *Store) findOrCreatePublisherTx(tx *sql.Tx, name string) (int, error) {
	var publisherID int
	err := tx.QueryRow(`SELECT id FROM publishers WHERE name = ?`, name).Scan(&publisherID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`INSERT INTO publishers (name, sort) VALUES (?, ?) RETURNING id`, name, "").Scan(&publisherID)
	}
	return publisherID, err
}
