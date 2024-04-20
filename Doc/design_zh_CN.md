# E-oasis 设计文档

## 项目背景

需要同步书籍信息，方便查看和管理。

## 功能需求

1. 书籍源文件同步（同步书籍信息）

- 书籍信息包括：书名、作者、出版社、出版日期、ISBN、价格、简介、目录、封面图片

2. 阅读进度同步

- 书籍阅读进度同步
- 书签同步

3. 摘抄/笔记同步

- 可以直接查看对应书籍的摘抄/笔记

4. 书籍下载

- 通过 Z-library 下载书籍

5. 界面友好

## 需求实现分析

数据库使用 SQLite。

因为需要保存文件，所以需要考虑文件的存储问题，可能需要考虑文件的压缩存储。

## 书籍源文件同步

- 断点续传
- 大文件压缩

### 书籍信息

书籍信息存储在 SQLite 中。

目前的想法：

```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL DEFAULT 'Unknown',
    sort TEXT,
    author_sort TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pubdate TIMESTAMP DEFAULT '0101-01-01 00:00:00',
    series_index TEXT NOT NULL DEFAULT '1.0',
    last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    path TEXT NOT NULL DEFAULT '',
    has_cover INTEGER DEFAULT 0,
    uuid TEXT,
    md5 TEXT,
    isbn TEXT DEFAULT '',
    flags INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE authors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
    sort TEXT,
    link TEXT DEFAULT ''
);

CREATE TABLE books_authors_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    author_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (author_id) REFERENCES authors (id)
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE books_tags_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    tag_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (tag_id) REFERENCES tags (id)
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment TEXT,
    book_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id)
);

CREATE TABLE data (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL,
    format TEXT NOT NULL,
    uncompressed_size INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (book_id) REFERENCES books (id)
);

CREATE TABLE series (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    sort TEXT
);

CREATE TABLE books_series_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    series_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (series_id) REFERENCES series (id)
);

CREATE TABLE publishers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE books_publishers_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    publisher_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (publisher_id) REFERENCES publishers (id)
);

CREATE TABLE identifiers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    identifier_type TEXT,
    value TEXT,
    book_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id)
);
```

以上表都是关于书籍信息，其中需要解释的点就是 sort 字段和 author_sort 字段，
sort 存储需要

```
sort_authors: Riccomini, Chris, input_authors: ['Chris Riccomini'], db_author: <Authors('Chris Riccomini,Riccomini, Chris')>, renamed_authors: []
```

author_sort:

```py
def get_sorted_author(value):
    value2 = None
    try:
        if ',' not in value:
            regexes = [r"^(JR|SR)\.?$", r"^I{1,3}\.?$", r"^IV\.?$"]
            # (^(JR|SR)\.?$)|(^I{1,3}\.?$)|(^IV\.?$)
            combined = "(" + ")|(".join(regexes) + ")"
            value = value.split(" ")
            if re.match(combined, value[-1].upper()):
                if len(value) > 1:
                    value2 = value[-2] + ", " + " ".join(value[:-2]) + " " + value[-1]
                else:
                    value2 = value[0]
            elif len(value) == 1:
                value2 = value[0]
            else:
                value2 = value[-1] + ", " + " ".join(value[:-1])
        else:
            value2 = value
    except Exception as ex:
        log.error("Sorting author %s failed: %s", value, ex)
        if isinstance(list, value2):
            value2 = value[0]
        else:
            value2 = value
    return value2
```

### 解析书籍信息

需要从文件 (epub, pdf, etc) 中解析封面图以及书籍信息。

### 书籍源文件

考虑是否压缩存储。

## 阅读信息同步

### 书籍阅读进度

存储在 SQLite 中。

```

```

### 书签

### 阅读时长

## 摘抄/笔记同步

1. 摘抄/笔记存储在数据库中

## 书籍下载

1. 通过 Z-library 下载书籍，需要考虑 API 的调用。
