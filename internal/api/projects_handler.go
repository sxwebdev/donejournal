package api

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	projectsv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/projects/v1"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/projects"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_projects"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProjectsHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewProjectsHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *ProjectsHandler {
	return &ProjectsHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *ProjectsHandler) ListProjects(ctx context.Context, req *connect.Request[projectsv1.ListProjectsRequest]) (*connect.Response[projectsv1.ListProjectsResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := repo_projects.FindParams{
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

	result, err := h.baseService.Projects().Find(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbProjects := make([]*projectsv1.ProjectStats, len(result.Items))
	for i, p := range result.Items {
		stats, err := h.baseService.Projects().GetStats(ctx, p.ID, userID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pbProjects[i] = projectStatsToProto(stats)
	}

	var nextPageToken string
	if params.Page != nil && params.PageSize != nil && uint32(len(result.Items)) == *params.PageSize {
		nextPageToken = fmt.Sprintf("%d", *params.Page+1)
	}

	return connect.NewResponse(&projectsv1.ListProjectsResponse{
		Projects:      pbProjects,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *ProjectsHandler) GetProject(ctx context.Context, req *connect.Request[projectsv1.GetProjectRequest]) (*connect.Response[projectsv1.GetProjectResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := h.baseService.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	if project.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	return connect.NewResponse(&projectsv1.GetProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectsHandler) CreateProject(ctx context.Context, req *connect.Request[projectsv1.CreateProjectRequest]) (*connect.Response[projectsv1.CreateProjectResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	project, err := h.baseService.Projects().Create(ctx, userID, req.Msg.GetName(), req.Msg.GetDescription())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectsv1.CreateProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectsHandler) UpdateProject(ctx context.Context, req *connect.Request[projectsv1.UpdateProjectRequest]) (*connect.Response[projectsv1.UpdateProjectResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	params := projects.UpdateParams{}
	if req.Msg.Name != nil {
		params.Name = req.Msg.Name
	}
	if req.Msg.Description != nil {
		params.Description = req.Msg.Description
	}

	project, err := h.baseService.Projects().Update(ctx, userID, req.Msg.GetId(), params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectsv1.UpdateProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectsHandler) DeleteProject(ctx context.Context, req *connect.Request[projectsv1.DeleteProjectRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	if err := h.baseService.Projects().Delete(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *ProjectsHandler) ArchiveProject(ctx context.Context, req *connect.Request[projectsv1.ArchiveProjectRequest]) (*connect.Response[projectsv1.ArchiveProjectResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	project, err := h.baseService.Projects().Archive(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectsv1.ArchiveProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectsHandler) UnarchiveProject(ctx context.Context, req *connect.Request[projectsv1.UnarchiveProjectRequest]) (*connect.Response[projectsv1.UnarchiveProjectResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := h.baseService.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	project, err := h.baseService.Projects().Unarchive(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectsv1.UnarchiveProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ProjectsHandler) GetProjectStats(ctx context.Context, req *connect.Request[projectsv1.GetProjectStatsRequest]) (*connect.Response[projectsv1.GetProjectStatsResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := h.baseService.Projects().GetStats(ctx, req.Msg.GetId(), userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	if stats.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project not found"))
	}

	return connect.NewResponse(&projectsv1.GetProjectStatsResponse{
		Stats: projectStatsToProto(stats),
	}), nil
}

func (h *ProjectsHandler) SubscribeProjects(
	ctx context.Context,
	req *connect.Request[projectsv1.SubscribeProjectsRequest],
	stream *connect.ServerStream[projectsv1.SubscribeProjectsResponse],
) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	ch := h.baseService.Projects().Broker().Subscribe()
	if ch == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription unavailable"))
	}
	defer h.baseService.Projects().Broker().Unsubscribe(ch)

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
			if err := stream.Send(&projectsv1.SubscribeProjectsResponse{}); err != nil {
				return err
			}
		}
	}
}
