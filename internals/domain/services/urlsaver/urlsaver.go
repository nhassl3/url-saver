package urlsaver

import (
	"context"
	"errors"
	"log/slog"

	urlsv1 "github.com/nhassl3/url-saver-contracts/generated/go/urlsaver"
	"github.com/nhassl3/url-saver/internals/domain/entities"
	"github.com/nhassl3/url-saver/internals/lib/logger/sl"
	"github.com/nhassl3/url-saver/internals/storage"
)

const (
	opSave          = "services.urlsaver.Save"
	opGet           = "services.urlsaver.Get"
	opUpdateByID    = "services.urlsaver.UpdateByID"
	onUpdateByAlias = "services.urlsaver.UpdateByAlias"
	opRemoveByID    = "services.urlsaver.RemoveByID"
	opRemoveByAlias = "services.urlsaver.RemoveByAlias"
	opList          = "services.urlsaver.List"
)

var (
	ErrAliasExists = errors.New("alias already exists")
)

type UrlSaver struct {
	log         *slog.Logger
	urlSaver    SaverUrl
	urlProvider ProviderUrl
	urlUpdater  UpdaterUrl
}

func NewUrlSaver(
	log *slog.Logger,
	urlSaver SaverUrl,
	urlProvider ProviderUrl,
	urlUpdater UpdaterUrl,
) *UrlSaver {
	return &UrlSaver{
		log:         log,
		urlSaver:    urlSaver,
		urlProvider: urlProvider,
		urlUpdater:  urlUpdater,
	}
}

type SaverUrl interface {
	SaveUrl(ctx context.Context, url, alias string) (urlID int64, err error)
}

type ProviderUrl interface {
	Url(ctx context.Context, alias string) (url entities.URL, err error)
	UrlList(ctx context.Context, alias string) (urls []entities.URL, err error)
}

type UpdaterUrl interface {
	UpdateUrl(ctx context.Context, urlID int64, alias string) (err error)
	RemoveUrl(ctx context.Context, alias string) (err error)
}

func (u *UrlSaver) Save(ctx context.Context, url, aliasReq string) (urlID int64, aliasRes string, err error) {
	log := u.log.With(slog.String("op", opSave))
	// TODO: Remove from protobuf file returning 3th parameters. Only 2 or less must be returnable
	aliasRes = aliasReq

	urlID, err = u.urlSaver.SaveUrl(ctx, url, aliasReq)
	if err != nil {
		if errors.Is(err, storage.ErrAliasExists) {
			return 0, "", sl.ErrUpLevel(opSave, ErrAliasExists.Error())
		}
		log.Error("failed to save url", sl.Err(err))

		return 0, "", sl.ErrUpLevel(opSave, err.Error())
	}

	return
}

func (u *UrlSaver) Get(ctx context.Context, aliasReq string) (url, aliasRes string, urlID int64, err error) {
	// TODO: implement domain zone
	panic("implement me")
}

func (u *UrlSaver) UpdateByID(ctx context.Context, urlID int64, newURL, newAliasReq string) (success bool, newAliasRes string, err error) {
	// TODO: implement domain zone
	panic("implement me")
}

func (u *UrlSaver) UpdateByAlias(ctx context.Context, alias, newURL, newAliasReq string) (success bool, newAliasRes string, err error) {
	// TODO: implement domain zone
	panic("implement me")
}

func (u *UrlSaver) RemoveByID(ctx context.Context, urlID int64) (success bool, removedUrlID int64, err error) {
	// TODO: implement domain zone
	panic("implement me")
}

func (u *UrlSaver) RemoveByAlias(ctx context.Context, aliasReq string) (success bool, removedUrlID int64, err error) {
	// TODO: implement domain zone
	panic("implement me")
}

func (u *UrlSaver) List(ctx context.Context, pageToken string, pageSize int32) (URLs []*urlsv1.UrlItem, nextPageToken string, err error) {
	// TODO: implement domain zone
	panic("implement me")
}
