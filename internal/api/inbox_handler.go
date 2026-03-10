package api

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	inboxv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/inbox/v1"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InboxHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewInboxHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *InboxHandler {
	return &InboxHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *InboxHandler) ListInboxItems(ctx context.Context, req *connect.Request[inboxv1.ListInboxItemsRequest]) (*connect.Response[inboxv1.ListInboxItemsResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pageSize := uint32(req.Msg.GetPageSize())
	if pageSize == 0 {
		pageSize = 20
	}

	var page *uint32
	if req.Msg.GetPageToken() != "" {
		p, err := strconv.ParseUint(req.Msg.GetPageToken(), 10, 32)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid page_token"))
		}
		pageVal := uint32(p)
		page = &pageVal
	} else {
		pageVal := uint32(1)
		page = &pageVal
	}

	result, err := h.baseService.Inbox().List(ctx, strconv.FormatInt(userID, 10), page, &pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*inboxv1.InboxItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = inboxItemToProto(item)
	}

	var nextPageToken string
	if uint32(len(result.Items)) == pageSize {
		nextPageToken = strconv.FormatUint(uint64(*page+1), 10)
	}

	return connect.NewResponse(&inboxv1.ListInboxItemsResponse{
		Items:         items,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *InboxHandler) GetInboxItem(ctx context.Context, req *connect.Request[inboxv1.GetInboxItemRequest]) (*connect.Response[inboxv1.GetInboxItemResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := h.baseService.Inbox().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}

	if item.UserID != strconv.FormatInt(userID, 10) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}

	return connect.NewResponse(&inboxv1.GetInboxItemResponse{
		Item: inboxItemToProto(item),
	}), nil
}

func (h *InboxHandler) CreateInboxItem(ctx context.Context, req *connect.Request[inboxv1.CreateInboxItemRequest]) (*connect.Response[inboxv1.CreateInboxItemResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := h.baseService.Inbox().Create(ctx, req.Msg.GetData(), strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&inboxv1.CreateInboxItemResponse{
		Item: inboxItemToProto(item),
	}), nil
}

func (h *InboxHandler) UpdateInboxItem(ctx context.Context, req *connect.Request[inboxv1.UpdateInboxItemRequest]) (*connect.Response[inboxv1.UpdateInboxItemResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Inbox().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}
	if existing.UserID != strconv.FormatInt(userID, 10) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}

	item, err := h.baseService.Inbox().Update(ctx, userID, req.Msg.GetId(), req.Msg.GetData(), req.Msg.GetAdditionalData())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inboxv1.UpdateInboxItemResponse{
		Item: inboxItemToProto(item),
	}), nil
}

func (h *InboxHandler) DeleteInboxItem(ctx context.Context, req *connect.Request[inboxv1.DeleteInboxItemRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Inbox().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}
	if existing.UserID != strconv.FormatInt(userID, 10) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}

	if err := h.baseService.Inbox().Delete(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *InboxHandler) ConvertToTodo(ctx context.Context, req *connect.Request[inboxv1.ConvertToTodoRequest]) (*connect.Response[inboxv1.ConvertToTodoResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Inbox().GetByID(ctx, req.Msg.GetInboxItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}
	if existing.UserID != strconv.FormatInt(userID, 10) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("inbox item not found"))
	}

	plannedDate := req.Msg.GetPlannedDate().AsTime()
	if plannedDate.IsZero() {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("planned_date is required"))
	}

	todoID, err := h.baseService.Inbox().ConvertToTodo(
		ctx,
		req.Msg.GetInboxItemId(),
		userID,
		req.Msg.GetTitle(),
		req.Msg.GetDescription(),
		plannedDate,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inboxv1.ConvertToTodoResponse{
		TodoId: todoID,
	}), nil
}

func (h *InboxHandler) SubscribeInbox(
	ctx context.Context,
	req *connect.Request[inboxv1.SubscribeInboxRequest],
	stream *connect.ServerStream[inboxv1.SubscribeInboxResponse],
) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	ch := h.baseService.Inbox().Broker().Subscribe()
	if ch == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription unavailable"))
	}
	defer h.baseService.Inbox().Broker().Unsubscribe(ch)

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
			if err := stream.Send(&inboxv1.SubscribeInboxResponse{}); err != nil {
				return err
			}
		}
	}
}
