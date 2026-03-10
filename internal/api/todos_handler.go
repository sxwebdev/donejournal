package api

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	todosv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/todos/v1"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/todos"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TodosHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewTodosHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *TodosHandler {
	return &TodosHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *TodosHandler) ListTodos(ctx context.Context, req *connect.Request[todosv1.ListTodosRequest]) (*connect.Response[todosv1.ListTodosResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := repo_todos.FindParams{
		UserID: userID,
	}

	if req.Msg.GetPageSize() > 0 {
		ps := uint32(req.Msg.GetPageSize())
		params.PageSize = &ps
	}

	if req.Msg.GetPageToken() != "" {
		// page_token is the page number as string
		var p uint32
		if _, err := fmt.Sscanf(req.Msg.GetPageToken(), "%d", &p); err == nil {
			params.Page = &p
		}
	} else {
		p := uint32(1)
		params.Page = &p
	}

	if req.Msg.PlannedDateFrom != nil {
		t := req.Msg.GetPlannedDateFrom().AsTime()
		params.DateFrom = &t
	}
	if req.Msg.PlannedDateTo != nil {
		t := req.Msg.GetPlannedDateTo().AsTime()
		params.DateTo = &t
	}

	// Map status filter
	for _, s := range req.Msg.GetStatuses() {
		if s != todosv1.TodoStatus_TODO_STATUS_UNSPECIFIED {
			params.Statuses = append(params.Statuses, todoStatusFromProto(s))
		}
	}

	result, err := h.baseService.Todos().Find(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbTodos := make([]*todosv1.Todo, len(result.Items))
	for i, t := range result.Items {
		pbTodos[i] = todoToProto(t)
	}

	var nextPageToken string
	if params.Page != nil && params.PageSize != nil && uint32(len(result.Items)) == *params.PageSize {
		nextPageToken = fmt.Sprintf("%d", *params.Page+1)
	}

	return connect.NewResponse(&todosv1.ListTodosResponse{
		Todos:         pbTodos,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *TodosHandler) GetTodo(ctx context.Context, req *connect.Request[todosv1.GetTodoRequest]) (*connect.Response[todosv1.GetTodoResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	todo, err := h.baseService.Todos().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}

	if todo.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}

	return connect.NewResponse(&todosv1.GetTodoResponse{
		Todo: todoToProto(todo),
	}), nil
}

func (h *TodosHandler) CreateTodo(ctx context.Context, req *connect.Request[todosv1.CreateTodoRequest]) (*connect.Response[todosv1.CreateTodoResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetTitle() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title is required"))
	}

	plannedDate := req.Msg.GetPlannedDate().AsTime()
	if plannedDate.IsZero() {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("planned_date is required"))
	}

	todo, err := h.baseService.Todos().CreateFromAPI(ctx, userID, req.Msg.GetTitle(), req.Msg.GetDescription(), plannedDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todosv1.CreateTodoResponse{
		Todo: todoToProto(todo),
	}), nil
}

func (h *TodosHandler) UpdateTodo(ctx context.Context, req *connect.Request[todosv1.UpdateTodoRequest]) (*connect.Response[todosv1.UpdateTodoResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Todos().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}

	params := todos.UpdateParams{}
	if req.Msg.Title != nil {
		params.Title = req.Msg.Title
	}
	if req.Msg.Description != nil {
		params.Description = req.Msg.Description
	}
	if req.Msg.Status != nil {
		status := todoStatusFromProto(*req.Msg.Status)
		params.Status = &status
	}
	if req.Msg.PlannedDate != nil {
		t := req.Msg.GetPlannedDate().AsTime()
		params.PlannedDate = &t
	}

	todo, err := h.baseService.Todos().Update(ctx, req.Msg.GetId(), params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todosv1.UpdateTodoResponse{
		Todo: todoToProto(todo),
	}), nil
}

func (h *TodosHandler) DeleteTodo(ctx context.Context, req *connect.Request[todosv1.DeleteTodoRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Todos().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}

	if err := h.baseService.Todos().Delete(ctx, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *TodosHandler) CompleteTodo(ctx context.Context, req *connect.Request[todosv1.CompleteTodoRequest]) (*connect.Response[todosv1.CompleteTodoResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Todos().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("todo not found"))
	}

	todo, err := h.baseService.Todos().Complete(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&todosv1.CompleteTodoResponse{
		Todo: todoToProto(todo),
	}), nil
}

func (h *TodosHandler) GetCalendarEntries(ctx context.Context, req *connect.Request[todosv1.GetCalendarEntriesRequest]) (*connect.Response[todosv1.GetCalendarEntriesResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	from := req.Msg.GetFrom().AsTime()
	to := req.Msg.GetTo().AsTime()

	if from.IsZero() || to.IsZero() {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("from and to are required"))
	}

	// Fetch all todos in the date range (no pagination limit for calendar)
	pageSize := uint32(100)
	page := uint32(1)
	result, err := h.baseService.Todos().Find(ctx, repo_todos.FindParams{
		UserID:   userID,
		DateFrom: &from,
		DateTo:   &to,
		Page:     &page,
		PageSize: &pageSize,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Group by date
	dayMap := make(map[string]*todosv1.CalendarDay)
	for _, t := range result.Items {
		dateKey := t.PlannedDate.Format(time.DateOnly)
		day, ok := dayMap[dateKey]
		if !ok {
			day = &todosv1.CalendarDay{
				Date: timestampFromDate(t.PlannedDate),
			}
			dayMap[dateKey] = day
		}
		day.Todos = append(day.Todos, todoToProto(t))
		day.TotalCount++
		if t.Status == "completed" {
			day.CompletedCount++
		}
	}

	days := make([]*todosv1.CalendarDay, 0, len(dayMap))
	for _, day := range dayMap {
		days = append(days, day)
	}

	return connect.NewResponse(&todosv1.GetCalendarEntriesResponse{
		Days: days,
	}), nil
}

func timestampFromDate(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
}
