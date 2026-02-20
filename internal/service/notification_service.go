package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/heth/STM/internal/middleware"
	"github.com/heth/STM/internal/model"
	"github.com/heth/STM/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TaskNotifier can broadcast task events (e.g. to gRPC subscribers).
type TaskNotifier interface {
	BroadcastTaskEvent(userID string, event *proto.TaskEvent)
}

// ModelTaskToProto converts a model.Task to proto.Task.
func ModelTaskToProto(t *model.Task) *proto.Task {
	if t == nil {
		return nil
	}
	status := "pending"
	if t.Completed {
		status = "completed"
	}
	return &proto.Task{
		Id:          fmt.Sprintf("%d", t.ID),
		Title:       t.Title,
		Description: t.Description,
		UserId:      fmt.Sprintf("%d", t.UserID),
		Status:      status,
	}
}

// TaskEventBroadcaster broadcasts task events to subscribed users.
type TaskEventBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *proto.TaskEvent // userID -> list of channels
}

// NewTaskEventBroadcaster creates a new broadcaster.
func NewTaskEventBroadcaster() *TaskEventBroadcaster {
	return &TaskEventBroadcaster{
		subscribers: make(map[string][]chan *proto.TaskEvent),
	}
}

// Subscribe adds a channel for the given user and returns it. The caller must
// listen on the channel until context is done, then call Unsubscribe.
func (b *TaskEventBroadcaster) Subscribe(userID string) (ch chan *proto.TaskEvent, unsubscribe func()) {
	ch = make(chan *proto.TaskEvent, 32)
	b.mu.Lock()
	b.subscribers[userID] = append(b.subscribers[userID], ch)
	b.mu.Unlock()
	unsubscribe = func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subscribers[userID]
		for i, c := range list {
			if c == ch {
				b.subscribers[userID] = append(list[:i], list[i+1:]...)
				if len(b.subscribers[userID]) == 0 {
					delete(b.subscribers, userID)
				}
				close(ch)
				break
			}
		}
	}
	return ch, unsubscribe
}

// BroadcastTaskEvent sends the event to all subscribers for the given userID.
func (b *TaskEventBroadcaster) BroadcastTaskEvent(userID string, event *proto.TaskEvent) {
	if event == nil {
		return
	}
	b.mu.RLock()
	list := make([]chan *proto.TaskEvent, len(b.subscribers[userID]))
	copy(list, b.subscribers[userID])
	b.mu.RUnlock()
	for _, ch := range list {
		select {
		case ch <- event:
		default:
			slog.Warn("notification channel full, dropping event", "user_id", userID)
		}
	}
}

// NotificationGrpcServer implements proto.NotificationServiceServer.
type NotificationGrpcServer struct {
	proto.UnimplementedNotificationServiceServer
	broadcaster *TaskEventBroadcaster
}

// NewNotificationGrpcServer creates a new gRPC notification server.
func NewNotificationGrpcServer(broadcaster *TaskEventBroadcaster) *NotificationGrpcServer {
	return &NotificationGrpcServer{broadcaster: broadcaster}
}

// taskEventStream is the stream interface with Send and Context.
type taskEventStream interface {
	Send(*proto.TaskEvent) error
	Context() context.Context
}

// SubscribeToTaskUpdates streams task events for the authenticated user.
func (s *NotificationGrpcServer) SubscribeToTaskUpdates(req *proto.SubscribeRequest, stream proto.NotificationService_SubscribeToTaskUpdatesServer) error {
	userID := req.GetUserId()
	if userID == "" {
		return status.Error(codes.InvalidArgument, "user_id required")
	}
	str, ok := stream.(taskEventStream)
	if !ok {
		slog.Error("stream does not implement taskEventStream")
		return status.Error(codes.Internal, "invalid stream")
	}
	ctxUserID, _ := str.Context().Value(middleware.GrpcUserIDKey).(string)
	if ctxUserID == "" || ctxUserID != userID {
		return status.Error(codes.PermissionDenied, "user_id must match authenticated user")
	}
	ch, unsubscribe := s.broadcaster.Subscribe(userID)
	defer unsubscribe()

	ctx := str.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			if err := str.Send(event); err != nil {
				slog.Error("failed to send task event", "error", err, "user_id", userID)
				return err
			}
		}
	}
}
