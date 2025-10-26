package urlsaver

import (
	"context"

	urlsv1 "github.com/nhassl3/url-saver-contracts/generated/go/urlsaver"
	urlshortener "github.com/nhassl3/url-saver/internals/clients/urlshortener/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NoIdentifier     = "None of several arguments were provided"
	UnknownAliasOrID = "An unknown Alias or ID was given"
)

type UrlSaver interface {
	Save(ctx context.Context, url, aliasReq string) (urlID int64, aliasRes string, err error)
	Get(ctx context.Context, aliasReq string) (url, aliasRes string, urlID int64, err error)
	UpdateByID(ctx context.Context, urlID int64, newURL, newAliasReq string) (success bool, newAliasRes string, err error)
	UpdateByAlias(ctx context.Context, alias, newURL, newAliasReq string) (success bool, newAliasRes string, err error)
	RemoveByID(ctx context.Context, urlID int64) (success bool, removedUrlID int64, err error)
	RemoveByAlias(ctx context.Context, aliasReq string) (success bool, removedUrlID int64, err error)
	List(ctx context.Context, pageToken string, pageSize int32) (URLs []*urlsv1.UrlItem, nextPageToken string, err error)
}

type ServerAPI struct {
	urlsv1.UnimplementedUrlSaverServer
	urlSaver     UrlSaver
	urlShortener *urlshortener.Client
}

// Register registration server through generated code from protobuf and use function
// registers UrlSaver server, but it's not only register this service
// and other services clients too
func Register(gRPC *grpc.Server, urlSaver UrlSaver, urlShortenerClient *urlshortener.Client) {
	urlsv1.RegisterUrlSaverServer(gRPC, &ServerAPI{urlSaver: urlSaver, urlShortener: urlShortenerClient})
}

func (api *ServerAPI) Save(ctx context.Context, in *urlsv1.SaveRequest) (*urlsv1.SaveResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	urlID, aliasRes, err := api.urlSaver.Save(ctx, in.GetUrl(), in.GetAlias())
	if err != nil {
		// TODO: implement not Internal error through condition error with errors.Is()
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlsv1.SaveResponse{
		UrlId: urlID,
		Alias: aliasRes,
	}, nil
}

func (api *ServerAPI) Get(ctx context.Context, in *urlsv1.GetRequest) (*urlsv1.GetResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	url, aliasRes, urlID, err := api.urlSaver.Get(ctx, in.GetAlias())
	if err != nil {
		// TODO: implement not Internal error through condition error with errors.Is()
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlsv1.GetResponse{
		Url:   url,
		Alias: aliasRes,
		UrlId: urlID,
	}, nil
}

func (api *ServerAPI) Update(ctx context.Context, in *urlsv1.UpdateRequest) (*urlsv1.UpdateResponse, error) {
	var (
		success              bool
		newAliasRes          string
		err                  error
		UrlShortenerResponse *urlshortener.ShortenResponse
	)

	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch v := in.GetIdentifier().(type) {
	case *urlsv1.UpdateRequest_UrlId:
		UrlShortenerResponse, err = api.urlShortener.ShortenURL(ctx, in.GetNewUrl(), in.GetNewAlias())
		if err != nil {
			// TODO: none Internal error too
			return nil, status.Error(codes.Internal, err.Error())
		}

		success, newAliasRes, err = api.urlSaver.UpdateByID(
			ctx, v.UrlId, UrlShortenerResponse.GetURL(), UrlShortenerResponse.GetAlias(),
		)
	case *urlsv1.UpdateRequest_Alias:
		UrlShortenerResponse, err = api.urlShortener.ShortenURL(ctx, in.GetNewUrl(), in.GetNewAlias())
		if err != nil {
			// TODO: none Internal error too
			return nil, status.Error(codes.Internal, err.Error())
		}

		success, newAliasRes, err = api.urlSaver.UpdateByAlias(
			ctx, v.Alias, UrlShortenerResponse.GetURL(), UrlShortenerResponse.GetAlias(),
		)
	case nil:
		return nil, status.Error(codes.InvalidArgument, NoIdentifier)
	default:
		return nil, status.Error(codes.InvalidArgument, UnknownAliasOrID)
	}

	if err != nil {
		// TODO: implement not Internal error through condition error with errors.Is()
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlsv1.UpdateResponse{
		Success:  success,
		NewAlias: newAliasRes,
	}, nil
}

func (api *ServerAPI) Remove(ctx context.Context, in *urlsv1.RemoveRequest) (*urlsv1.RemoveResponse, error) {
	var (
		success      bool
		removedUrlID int64
		err          error
	)

	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch v := in.GetIdentifier().(type) {
	case *urlsv1.RemoveRequest_UrlId:
		success, removedUrlID, err = api.urlSaver.RemoveByID(ctx, v.UrlId)
	case *urlsv1.RemoveRequest_Alias:
		success, removedUrlID, err = api.urlSaver.RemoveByAlias(ctx, v.Alias)
	case nil:
		return nil, status.Error(codes.InvalidArgument, NoIdentifier)
	default:
		return nil, status.Error(codes.InvalidArgument, UnknownAliasOrID)
	}

	if err != nil {
		// TODO: implement not Internal error through condition error with errors.Is()
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlsv1.RemoveResponse{
		Success:      success,
		RemovedUrlId: removedUrlID,
	}, nil
}

func (api *ServerAPI) List(ctx context.Context, in *urlsv1.ListRequest) (*urlsv1.ListResponse, error) {
	if err := in.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	URLs, nextPageToken, err := api.urlSaver.List(ctx, in.GetPageToken(), in.GetPageSize())
	if err != nil {
		// TODO: implement not Internal error through condition error with errors.Is()
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &urlsv1.ListResponse{
		Urls:          URLs,
		NextPageToken: nextPageToken,
	}, nil
}
