# E-oasis Design Document

## Requirements Analysis

### Project Background

The need to synchronize book information for easy viewing and management.

### Functional Requirements

1. Synchronization of book source files (synchronize book information)
   - Book information includes: title, author, publisher, publication date, ISBN, price, summary, table of contents, cover image.

2. Synchronization of reading progress
   - Synchronize book reading progress
   - Synchronize bookmarks

3. Synchronization of excerpts/notes
   - Direct viewing of corresponding book excerpts/notes

4. Book downloading
   - Download books via Z-library or others.

5. User-friendly interface

### Implementation Analysis

SQLite is used for the database.

Due to the need to save files, storage of files needs to be considered, possibly requiring compressed storage for files.

#### Synchronization of Book Source Files

1. Book information is stored in the database (SQLite)
   > Need to parse cover images and book information from files (epub, pdf, etc.).
2. Book source files are stored on the server (consider compressed storage).

### Synchronization of Reading Progress

1. Book reading progress is stored in the database.
2. Bookmarks are stored in the database.

### Synchronization of Excerpts/Notes

1. Excerpts/notes are stored in the database.

### Book Downloading

1. Download books via Z-library, considering API calls.
