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
	if v := find.Isbn; v != nil {
		where, args = append(where, "isbn = ?"), append(args, *v)
	}
	if v := find.Iccn; v != nil {
		where, args = append(where, "iccn = ?"), append(args, *v)
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
            iccn,
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

	rows, err := s.db.Query(query, args...)
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
            &book.Isbn,
            &book.Iccn,
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
func (s *Store) AddBook(book *model.Book) error {
    return nil
}
