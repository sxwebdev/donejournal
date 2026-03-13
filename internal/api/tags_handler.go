package api

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	tagsv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/tags/v1"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/tags"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_tags"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TagsHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewTagsHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *TagsHandler {
	return &TagsHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *TagsHandler) ListTags(ctx context.Context, req *connect.Request[tagsv1.ListTagsRequest]) (*connect.Response[tagsv1.ListTagsResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := repo_tags.FindParams{
		UserID: userID,
	}

	if req.Msg.GetPageSize() > 0 {
		ps := uint32(req.Msg.GetPageSize())
		params.PageSize = &ps
	}

	if req.Msg.GetPageToken() != "" {
		var p uint32
		if _, err := fmt.Sscanf(req.Msg.GetPageToken(), "%d", &p); err == nil {
			params.Page = &p
		}
	} else {
		p := uint32(1)
		params.Page = &p
	}

	if req.Msg.Search != nil {
		params.Search = req.Msg.Search
	}

	result, err := h.baseService.Tags().Find(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbTags := make([]*tagsv1.Tag, len(result.Items))
	for i, t := range result.Items {
		pbTags[i] = tagToProto(t)
	}

	var nextPageToken string
	if params.Page != nil && params.PageSize != nil && uint32(len(result.Items)) == *params.PageSize {
		nextPageToken = fmt.Sprintf("%d", *params.Page+1)
	}

	return connect.NewResponse(&tagsv1.ListTagsResponse{
		Tags:          pbTags,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *TagsHandler) CreateTag(ctx context.Context, req *connect.Request[tagsv1.CreateTagRequest]) (*connect.Response[tagsv1.CreateTagResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	tag, err := h.baseService.Tags().Create(ctx, userID, req.Msg.GetName(), req.Msg.GetColor())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&tagsv1.CreateTagResponse{
		Tag: tagToProto(tag),
	}), nil
}

func (h *TagsHandler) UpdateTag(ctx context.Context, req *connect.Request[tagsv1.UpdateTagRequest]) (*connect.Response[tagsv1.UpdateTagResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Tags().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
	}

	params := tags.UpdateParams{}
	if req.Msg.Name != nil {
		params.Name = req.Msg.Name
	}
	if req.Msg.Color != nil {
		params.Color = req.Msg.Color
	}

	tag, err := h.baseService.Tags().Update(ctx, userID, req.Msg.GetId(), params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&tagsv1.UpdateTagResponse{
		Tag: tagToProto(tag),
	}), nil
}

func (h *TagsHandler) DeleteTag(ctx context.Context, req *connect.Request[tagsv1.DeleteTagRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Tags().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
	}

	if err := h.baseService.Tags().Delete(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *TagsHandler) SubscribeTags(
	ctx context.Context,
	req *connect.Request[tagsv1.SubscribeTagsRequest],
	stream *connect.ServerStream[tagsv1.SubscribeTagsResponse],
) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	ch := h.baseService.Tags().Broker().Subscribe()
	if ch == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription unavailable"))
	}
	defer h.baseService.Tags().Broker().Unsubscribe(ch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if event.UserID != userID {
				continue
			}
			if err := stream.Send(&tagsv1.SubscribeTagsResponse{}); err != nil {
				return err
			}
		}
	}
}
