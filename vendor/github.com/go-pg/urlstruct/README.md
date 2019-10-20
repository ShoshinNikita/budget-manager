# urlstruct decodes url.Values into structs

[![Build Status](https://travis-ci.org/go-pg/urlstruct.svg?branch=master)](https://travis-ci.org/go-pg/urlstruct)
[![GoDoc](https://godoc.org/github.com/go-pg/urlstruct?status.svg)](https://godoc.org/github.com/go-pg/urlstruct)

## Example

Following example decodes URL query `?page=2&limit=100&author_id=123` into a struct and uses [go-pg](https://github.com/go-pg/pg) feature `WhereStruct` to autogenerate WHERE clause:

```go
package urlstruct_test

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/urlstruct"
)

type Book struct {
	ID        int64
	AuthorID  int64
	CreatedAt time.Time
}

type BookFilter struct {
	urlstruct.Pager
	AuthorID int64
}

func ExampleDecode_filter() {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "",
		Database: "postgres",
	})
	defer db.Close()

	values := url.Values{
		"author_id": {"123"},
		"page":      {"2"},
		"limit":     {"100"},
	}
	filter := new(BookFilter)
	err := urlstruct.Unmarshal(values, filter)
	if err != nil {
		panic(err)
	}

	filter.Pager.MaxLimit = 100     // default max limit is 1000
	filter.Pager.MaxOffset = 100000 // default max offset is 1000000

	// Following query generates:
	//
	// SELECT "book"."id", "book"."author_id", "book"."created_at"
	// FROM "books" AS "book"
	// WHERE author_id = 123
	// LIMIT 100 OFFSET 100

	var books []*Book
	_ = db.Model(&books).
		WhereStruct(filter).
		Limit(filter.Pager.GetLimit()).
		Offset(filter.Pager.GetOffset()).
		Select()

	fmt.Println("author", filter.AuthorID)
	fmt.Println("limit", filter.GetLimit())
	fmt.Println("offset", filter.GetLimit())
	// Output: author 123
	// limit 100
	// offset 100
}
```
