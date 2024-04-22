# E-Oasis 设计文档

## 项目背景

需要同步书籍信息，方便查看和管理。

## 功能需求

1. 书籍源文件同步（同步书籍信息）

- 书籍信息包括：书名、作者、出版社、出版日期、ISBN、简介、目录、封面图片

2. 阅读进度同步

- 书籍阅读进度同步
- 书签同步

3. 摘抄/笔记同步

- 可以直接查看对应书籍的摘抄/笔记

4. 书籍下载

- 通过 Z-library 下载书籍

5. 界面友好

## 需求实现分析

数据库使用 SQLite，后端使用 Go。

因为需要保存文件，所以需要考虑文件的存储问题，可能需要考虑文件的压缩存储。

**注意**！！！

目前是由两个数据库文件，分别是 `metadata.db` 和 `app.db`。
`metadata.db` 存储书籍的信息，在下面的书籍信息标题下的所有 SQL 都在这个文件下，其他的所有 SQL 都在 `app.db` 下包括书架，阅读时长，用户管理等。

### 书籍源文件同步

- 断点续传
- 大文件压缩

### 书籍信息

书籍信息存储在 SQLite 中。

目前的想法：

```sql
CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL DEFAULT 'Unknown' COLLATE NOCASE,
    sort TEXT COLLATE NOCASE,
    author_sort TEXT COLLATE NOCASE,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pubdate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    series_index TEXT NOT NULL DEFAULT '1.0',
    path TEXT NOT NULL DEFAULT '',
    has_cover SMALLINT DEFAULT 0,
    flags INTEGER NOT NULL DEFAULT 1,
    uuid TEXT,
    md5 TEXT,
    isbn TEXT DEFAULT '' COLLATE NOCASE,
    iccn TEXT DEFAULT '' COLLATE NOCASE,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    FOREIGN KEY (author_id) REFERENCES authors (id),
    UNIQUE(book_id, author_id)
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE,
    UNIQUE (name)
);

CREATE TABLE books_tags_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    tag_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (tag_id) REFERENCES tags (id),
    UNIQUE(book_id, tag_id)
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment TEXT,
    book_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    UNIQUE(book_id)
);

CREATE TABLE data (
    id INTEGER PRIMARY KEY,
    book_id INTEGER NOT NULL,
    format TEXT NOT NULL COLLATE NOCASE,
    uncompressed_size INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (book_id) REFERENCES books (id)
    UNIQUE(book_id, format)
);

CREATE TABLE series (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE,
    sort TEXT COLLATE NOCASE
    UNIQUE(name)
);

CREATE TABLE books_series_link (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER,
    series_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (series_id) REFERENCES series (id)
    UNIQUE(book_id)
);

CREATE TABLE publishers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE,
    sort TEXT COLLATE NOCASE,
    UNIQUE(name)
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
    type TEXT NOT NULL DEFAULT "isbn" COLLATE NOCASE,
    value TEXT,
    book_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id)
    UNIQUE(book_id, type)
);
```

以上表都是关于书籍信息，其中需要解释的点就是 sort 字段和 author_sort 字段，这些都是针对非中文书籍。

sort 存储需要将前面的冠词换到后面：

title_sort:

```py
    def update_title_sort(self, config, conn=None):
        # user defined sort function for calibre databases (Series, etc.)
        def _title_sort(title):
            # calibre sort stuff
            # ^(A|The|An|Der|Die|Das|Den|Ein|Eine|Einen|Dem|Des|Einem|Eines|Le|La|Les|L\'|Un|Une)\s+
            title_pat = re.compile(config.config_title_regex, re.IGNORECASE)
            match = title_pat.search(title)
            if match:
                prep = match.group(1)
                title = title[len(prep):] + ', ' + prep
            return title.strip()

        try:
            # sqlalchemy <1.4.24
            conn = conn or self.session.connection().connection.driver_connection
        except AttributeError:
            # sqlalchemy >1.4.24 and sqlalchemy 2.0
            conn = conn or self.session.connection().connection.connection
        try:
            conn.create_function("title_sort", 1, _title_sort)
        except sqliteOperationalError:
            pass
```

> 'The Missing README' -> 'Missing README, The'

author_sort 遵循以下代码，进行排序处理：

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

> 'Chris Riccomini' -> 'Riccomini, Chris'

```
sort_authors: Riccomini, Chris, input_authors: ['Chris Riccomini'], db_author: <Authors('Chris Riccomini,Riccomini, Chris')>, renamed_authors: []
```

**书籍语言**：

语言使用语言代码，根据 [ISO_639](https://zh.wikipedia.org/zh-cn/ISO_639)，其中采取 [ISO_639-3](https://zh.wikipedia.org/zh-cn/ISO_639-3)。

```py

    def _copy_fields(l):
        l.part1 = getattr(l, 'alpha_2', None)
        l.part3 = getattr(l, 'alpha_3', None)
        return l

    def get(name=None, part1=None, part3=None):
        if part3 is not None:
            return _copy_fields(pyc_languages.get(alpha_3=part3))
        if part1 is not None:
            return _copy_fields(pyc_languages.get(alpha_2=part1))
        if name is not None:
            return _copy_fields(pyc_languages.get(name=name))

def get_lang3(lang):
    try:
        if len(lang) == 2:
            ret_value = get(part1=lang).part3
        elif len(lang) == 3:
            ret_value = lang
        else:
            ret_value = ""
    except KeyError:
        ret_value = lang
    return ret_value

---
    lang = epub_metadata['language'].split('-', 1)[0].lower()
    # lang = zh
    log.debug('Language: {}'.format(lang))
    epub_metadata['language'] = isoLanguages.get_lang3(lang)
    # zho
```

### 解析书籍信息

需要从文件 (epub, pdf, etc) 中解析封面图以及书籍信息。

### 书籍源文件

考虑是否压缩存储。`data` 表。
关于书籍源文件需要考虑到书籍的大小，是否应该添加云盘（比如 onedrive）来存储书籍而不是存放到服务器上。

目前有这样一个想法：

E-Oasis 提供上传书籍和下载书籍的接口，这两个接口仅用于同步图书。
比如，用户可以通过向 E-Oasis 上传书籍，E-Oasis 提供一个接口给用户使用，
E-Oasis 通过这个接口将用户上传的书籍传到用户设备上。
同样的，也可以将用户设备上的书籍同步到 E-Oasis 中。

有个问题就是 E-Oasis 该提供存储功能吗？

如果不提供，而是通过云盘进行存储的话，E-Oasis 只需要存放书籍的信息，而不是书籍本身。
书籍本身可以向云盘获取。

### 书籍下载

通过 Z-library 下载书籍，需要考虑 API 的调用。

### 阅读信息同步

这些信息由用户提供，因为 `E-Oasis` 是没有阅读功能的。

### 书籍阅读进度

存储在 SQLite 中。

```sql
CREATE TABLE reading_status (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER NOT NULL,
    last_read_time DATETIME NOT NULL DEFAULT "2000-01-01 00:00:00+00:00",
    reading_duration INTEGER NOT NULL DEFAULT 0,
    cur_page INTEGER NOT NULL DEFAULT -1,
    status SMALLINT,
    FOREIGN KEY(user_id) REFERENCES user (id)
)
```

> 因为在不同的数据库文件下，所以不支持把 `bookt_id` 设置成外键。

字段 `staus` 分别为 0, 1, 2, 3，分别表示未看、想看、在看、看过。

### 最近阅读

最近阅读可以从 `reading_status` 表中查询 `status` 字段为 `2` 的数据。

### 书签

```sql
CREATE TABLE bookmark (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    tips TEXT,
    FOREIGN KEY(user_id) REFERENCES user (id),
    UNIQUE(position, book_id, user_id)
)
```

`position` 表示书签所在页；`tips` 表示书签所在页的提示信息，可以为空。

### 阅读时长

```sql
CREATE TABLE duration_info (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    book_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL DEFAULT "2000-01-01 00:00:00+00:00",
    read_duration INTEGER,
    percentage INTEGER,
    FOREIGN KEY(user_id) REFERENCES user (id)
)
```

这个表记录用户每次阅读的时长，但是不能保证所有用户都拥有这些数据，所以这个表只用与某些特定的功能。

获取阅读时长有两种方法，可以直接从 `reading_status` 表中的 `read_duration` 字段获取，也可以从 `duration_info` 计算得来。

### 摘抄/笔记同步

摘抄/笔记存储在数据库中

### 用户管理

```sql
CREATE TABLE user (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	name VARCHAR(64),
	email VARCHAR(120),
	role SMALLINT,
	password VARCHAR,
	view_settings JSON,
)
```

`role` 为管理员和用户。
`view_settings` 控制用户的配置。暂时不确定应该如何使用。

#### 用户登录

使用 session 控制用户登录信息：

```sql
CREATE TABLE user_session (
	id INTEGER NOT NULL,
	user_id INTEGER,
	session_key VARCHAR,
	PRIMARY KEY (id),
	FOREIGN KEY(user_id) REFERENCES user (id)
)
```

### 书架管理

支持多用户，多个用户可以创建书架/图书馆，并且可以公开给其他用户。
书架支持用户查看时本地排序，也可以由书架创立者将书架上的书排序之后分享出去，用户看到的就是创立者排序的。
也可以由创立者设置**视图公开**，视图公开则所有用户都可以更改排序切这个排序就是书架的排序。

> 排序操作可以手动排序。

```sql
CREATE TABLE shelf (
	id INTEGER NOT NULL,
	uuid VARCHAR,
	name VARCHAR NOT NULL,
	is_public SMALLINT DEFAULT 0,
	user_id INTEGER,
	created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 	last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 	display_order VARCHAR(20) NOT NULL COLLATE NOCASE,
	order_reverse SMALLINT DEFAULT 0,
	PRIMARY KEY (id),
	FOREIGN KEY(user_id) REFERENCES user (id)
)

CREATE TABLE book_shelf_link (
	id INTEGER NOT NULL,
	book_id INTEGER,
	position INTEGER NOT NULL DEFAULT 1,
	shelf_id INTEGER,
	date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	FOREIGN KEY(shelf_id) REFERENCES shelf (id)
)
```

shelf 表中的 `display_order` 有 `mannual`, `modified_time`, `title`, `authors`, `publish_time`。`order_reverse` 字段表示是否逆序。

book_shelf_link 表中的 `position` 表示当前这本书在 shelf 中的位置。用于手动排序。

### system_setting

系统设置，将配置存在 SQLite。

```sql
CREATE TABLE system_setting (
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    UNIQUE(name)
)
```

考虑应该把什么 setting 放置在这个表中。

目前考虑跟日志相关的配置可以存放在这里，以及之后想要的扩展 (plugins) 的一些配置。

将 `system_setting` 与服务相关的（不包括扩展）配置分为 `system_basic` 和 `system_general` 两种。

`system_basic` 是配置一些配置比如日志等级和邮件服务设置等。设想如下：

- 服务器配置：端口，可信主机
- 日志配置：日志级别，日志文件路径

`system_general` 主要配置功能方面：

- 用户密码策略
- 用户注册
- 上传书籍功能，允许上传的书籍格式
- 匿名浏览

`system_plugin`:

- 扩展功能 (plugins)，这里是指扩展功能的配置，而不是扩展的 token 等。
