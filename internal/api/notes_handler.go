package api

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	notesv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/notes/v1"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/services/notes"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_notes"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotesHandler struct {
	logger      logger.Logger
	baseService *baseservices.BaseServices
	store       *store.Store
}

func NewNotesHandler(l logger.Logger, baseService *baseservices.BaseServices, st *store.Store) *NotesHandler {
	return &NotesHandler{
		logger:      l,
		baseService: baseService,
		store:       st,
	}
}

func (h *NotesHandler) ListNotes(ctx context.Context, req *connect.Request[notesv1.ListNotesRequest]) (*connect.Response[notesv1.ListNotesResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params := repo_notes.FindParams{
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
		s := req.Msg.GetSearch()
		params.Search = &s
	}

	if req.Msg.WorkspaceId != nil {
		params.WorkspaceID = req.Msg.WorkspaceId
	}

	if len(req.Msg.GetTagIds()) > 0 {
		params.TagIDs = req.Msg.GetTagIds()
	}

	result, err := h.baseService.Notes().Find(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Batch load tag IDs for all notes
	noteIDs := make([]string, len(result.Items))
	for i, n := range result.Items {
		noteIDs[i] = n.ID
	}
	tagIDsMap, err := h.baseService.Tags().FindTagIDsByNoteIDs(ctx, noteIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbNotes := make([]*notesv1.Note, len(result.Items))
	for i, n := range result.Items {
		pb := noteToProto(n)
		pb.TagIds = tagIDsMap[n.ID]
		pbNotes[i] = pb
	}

	var nextPageToken string
	if params.Page != nil && params.PageSize != nil && uint32(len(result.Items)) == *params.PageSize {
		nextPageToken = fmt.Sprintf("%d", *params.Page+1)
	}

	return connect.NewResponse(&notesv1.ListNotesResponse{
		Notes:         pbNotes,
		NextPageToken: nextPageToken,
		TotalCount:    int32(result.Count),
	}), nil
}

func (h *NotesHandler) GetNote(ctx context.Context, req *connect.Request[notesv1.GetNoteRequest]) (*connect.Response[notesv1.GetNoteResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	note, err := h.baseService.Notes().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}

	if note.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}

	pb := noteToProto(note)
	noteTags, err := h.baseService.Tags().FindByNoteID(ctx, note.ID)
	if err == nil {
		for _, t := range noteTags {
			pb.TagIds = append(pb.TagIds, t.ID)
		}
	}

	return connect.NewResponse(&notesv1.GetNoteResponse{
		Note: pb,
	}), nil
}

func (h *NotesHandler) CreateNote(ctx context.Context, req *connect.Request[notesv1.CreateNoteRequest]) (*connect.Response[notesv1.CreateNoteResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.GetTitle() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title is required"))
	}

	note, err := h.baseService.Notes().Create(ctx, userID, req.Msg.GetTitle(), req.Msg.GetBody(), req.Msg.WorkspaceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pb := noteToProto(note)
	if len(req.Msg.GetTagIds()) > 0 {
		if err := h.baseService.Tags().SetNoteTags(ctx, userID, note.ID, req.Msg.GetTagIds()); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pb.TagIds = req.Msg.GetTagIds()
	}

	return connect.NewResponse(&notesv1.CreateNoteResponse{
		Note: pb,
	}), nil
}

func (h *NotesHandler) UpdateNote(ctx context.Context, req *connect.Request[notesv1.UpdateNoteRequest]) (*connect.Response[notesv1.UpdateNoteResponse], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Notes().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}

	params := notes.UpdateParams{}
	if req.Msg.Title != nil {
		params.Title = req.Msg.Title
	}
	if req.Msg.Body != nil {
		params.Body = req.Msg.Body
	}
	if req.Msg.WorkspaceId != nil {
		params.WorkspaceID = req.Msg.WorkspaceId
	}

	note, err := h.baseService.Notes().Update(ctx, userID, req.Msg.GetId(), params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pb := noteToProto(note)
	if len(req.Msg.GetTagIds()) > 0 {
		if err := h.baseService.Tags().SetNoteTags(ctx, userID, note.ID, req.Msg.GetTagIds()); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		pb.TagIds = req.Msg.GetTagIds()
	}

	return connect.NewResponse(&notesv1.UpdateNoteResponse{
		Note: pb,
	}), nil
}

func (h *NotesHandler) DeleteNote(ctx context.Context, req *connect.Request[notesv1.DeleteNoteRequest]) (*connect.Response[emptypb.Empty], error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check ownership
	existing, err := h.baseService.Notes().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}
	if existing.UserID != userID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("note not found"))
	}

	if err := h.baseService.Notes().Delete(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *NotesHandler) SubscribeNotes(
	ctx context.Context,
	req *connect.Request[notesv1.SubscribeNotesRequest],
	stream *connect.ServerStream[notesv1.SubscribeNotesResponse],
) error {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	ch := h.baseService.Notes().Broker().Subscribe()
	if ch == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("subscription unavailable"))
	}
	defer h.baseService.Notes().Broker().Unsubscribe(ch)

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
			if err := stream.Send(&notesv1.SubscribeNotesResponse{}); err != nil {
				return err
			}
		}
	}
}
