package service

import (
	"context"
	"time"

	"notification-service/internal/model"
	"notification-service/internal/store"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo store.NotificationRepository
}

func New(repo store.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Enqueue(ctx context.Context, typ, recipient string, payload []byte) error {
	now := time.Now().UTC()
	n := model.Notification{
		ID:        uuid.NewString(),
		Type:      typ,
		Recipient: recipient,
		Payload:   payload,
		Status:    model.StatusPending,
		Attempts:  0,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	return s.repo.Insert(ctx, &n)
}
