package api

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	workspacesv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/workspaces/v1"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/workspaces"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_workspaces"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type WorkspacesHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewWorkspacesHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *WorkspacesHandler {
	return &WorkspacesHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *WorkspacesHandler) ListWorkspaces(ctx context.Context, req *connect.Request[workspacesv1.ListWorkspacesRequest]) (*connect.Response[workspacesv1.ListWorkspacesResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := repo_workspaces.FindParams{
		UserID:          userID,
		IncludeArchived: req.Msg.GetIncludeArchived(),
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

	result, err := h.baseService.Workspaces().Find(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbWorkspaces := make([]*workspacesv1.WorkspaceStats, len(result.Items))
	for i, w := range result.Items {
		stats, err := h.baseService.Workspaces().GetStats(ctx, w.ID, userID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pbWorkspaces[i] = workspaceStatsToProto(stats)
	}

	var nextPageToken string
	if params.Page != nil && params.PageSize != nil && uint32(len(result.Items)) == *params.PageSize {
		nextPageToken = fmt.Sprintf("%d", *params.Page+1)
	}

	return connect.NewResponse(&workspacesv1.ListWorkspacesResponse{
		Workspaces:    pbWorkspaces,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *WorkspacesHandler) GetWorkspace(ctx context.Context, req *connect.Request[workspacesv1.GetWorkspaceRequest]) (*connect.Response[workspacesv1.GetWorkspaceResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	workspace, err := h.baseService.Workspaces().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	if workspace.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	return connect.NewResponse(&workspacesv1.GetWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}), nil
}

func (h *WorkspacesHandler) CreateWorkspace(ctx context.Context, req *connect.Request[workspacesv1.CreateWorkspaceRequest]) (*connect.Response[workspacesv1.CreateWorkspaceResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	workspace, err := h.baseService.Workspaces().Create(ctx, userID, req.Msg.GetName(), req.Msg.GetDescription())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workspacesv1.CreateWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}), nil
}

func (h *WorkspacesHandler) UpdateWorkspace(ctx context.Context, req *connect.Request[workspacesv1.UpdateWorkspaceRequest]) (*connect.Response[workspacesv1.UpdateWorkspaceResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Workspaces().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	params := workspaces.UpdateParams{}
	if req.Msg.Name != nil {
		params.Name = req.Msg.Name
	}
	if req.Msg.Description != nil {
		params.Description = req.Msg.Description
	}

	workspace, err := h.baseService.Workspaces().Update(ctx, userID, req.Msg.GetId(), params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workspacesv1.UpdateWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}), nil
}

func (h *WorkspacesHandler) DeleteWorkspace(ctx context.Context, req *connect.Request[workspacesv1.DeleteWorkspaceRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Workspaces().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	if err := h.baseService.Workspaces().Delete(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *WorkspacesHandler) ArchiveWorkspace(ctx context.Context, req *connect.Request[workspacesv1.ArchiveWorkspaceRequest]) (*connect.Response[workspacesv1.ArchiveWorkspaceResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Workspaces().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	workspace, err := h.baseService.Workspaces().Archive(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workspacesv1.ArchiveWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}), nil
}

func (h *WorkspacesHandler) UnarchiveWorkspace(ctx context.Context, req *connect.Request[workspacesv1.UnarchiveWorkspaceRequest]) (*connect.Response[workspacesv1.UnarchiveWorkspaceResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Workspaces().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	workspace, err := h.baseService.Workspaces().Unarchive(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workspacesv1.UnarchiveWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}), nil
}

func (h *WorkspacesHandler) GetWorkspaceStats(ctx context.Context, req *connect.Request[workspacesv1.GetWorkspaceStatsRequest]) (*connect.Response[workspacesv1.GetWorkspaceStatsResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := h.baseService.Workspaces().GetStats(ctx, req.Msg.GetId(), userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	if stats.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("workspace not found"))
	}

	return connect.NewResponse(&workspacesv1.GetWorkspaceStatsResponse{
		Stats: workspaceStatsToProto(stats),
	}), nil
}

func (h *WorkspacesHandler) SubscribeWorkspaces(
	ctx context.Context,
	req *connect.Request[workspacesv1.SubscribeWorkspacesRequest],
	stream *connect.ServerStream[workspacesv1.SubscribeWorkspacesResponse],
) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	ch := h.baseService.Workspaces().Broker().Subscribe()
	if ch == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription unavailable"))
	}
	defer h.baseService.Workspaces().Broker().Unsubscribe(ch)

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
			if err := stream.Send(&workspacesv1.SubscribeWorkspacesResponse{}); err != nil {
				return err
			}
		}
	}
}
