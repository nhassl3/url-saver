package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nhassl3/url-saver/internals/domain/entities"
	"github.com/nhassl3/url-saver/internals/lib/logger/sl"
	"github.com/nhassl3/url-saver/internals/storage"
)

const (
	opNewStorage = "sqlite.NewStorage"
	opSaveUrl    = "sqlite.SaveUrl"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, sl.ErrUpLevel(opNewStorage, err.Error())
	}
	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveUrl(ctx context.Context, url, alias string) (urlID int64, err error) {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO urls (user_id, url, alias) VALUES (?, ?, ?)")
	if err != nil {
		return 0, sl.ErrUpLevel(opSaveUrl, err.Error())
	}
	defer stmt.Close()

	var sqliteErr sqlite3.Error
	// TODO: insert correct user id
	res, err := stmt.ExecContext(ctx, url, 1, alias)
	if err != nil {
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, sl.ErrUpLevel(opSaveUrl, storage.ErrAliasExists.Error())
		}
		return 0, sl.ErrUpLevel(opSaveUrl, err.Error())
	}

	urlID, err = res.LastInsertId()
	if err != nil {
		return 0, sl.ErrUpLevel(opSaveUrl, err.Error())
	}

	return
}

func (s *Storage) Url(ctx context.Context, alias string) (url entities.URL, err error) {
	// TODO: implement orm project system
	panic("implement me")
}

func (s *Storage) UrlList(ctx context.Context, alias string) (urls []entities.URL, err error) {
	// TODO: implement orm project system
	panic("implement me")
}

func (s *Storage) UpdateUrl(ctx context.Context, urlID int64, alias string) (err error) {
	// TODO: implement orm project system
	panic("implement me")
}

func (s *Storage) RemoveUrl(ctx context.Context, alias string) (err error) {
	// TODO: implement orm project system
	panic("implement me")
}
