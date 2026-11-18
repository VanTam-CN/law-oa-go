package services

import (
	"context"
	"fmt"
	"strings"
)

type HandoffService struct {
	clientService *ClientService
	inboxService  *InboxService
}

func NewHandoffService(clientService *ClientService, inboxService *InboxService) *HandoffService {
	return &HandoffService{
		clientService: clientService,
		inboxService:  inboxService,
	}
}

type CreateClientHandoffRequest struct {
	TargetUserID uint    `json:"target_user_id" binding:"required"`
	TargetRole   string  `json:"target_role"`
	Note         string  `json:"note" binding:"omitempty,max=5000"`
	Priority     string  `json:"priority" binding:"omitempty,oneof=critical high medium low"`
	DueDate      *string `json:"due_date"`
}

type ClientHandoffResponse struct {
	ClientID     uint               `json:"client_id"`
	ClientName   string             `json:"client_name"`
	TargetUserID uint               `json:"target_user_id"`
	TargetRole   string             `json:"target_role"`
	Status       string             `json:"status"`
	InboxItem    *InboxItemResponse `json:"inbox_item"`
}

func (s *HandoffService) CreateClientHandoff(ctx context.Context, clientID uint, actorID uint, actorName string, req *CreateClientHandoffRequest) (*ClientHandoffResponse, error) {
	if req.TargetUserID == 0 {
		return nil, fmt.Errorf("target_user_id is required")
	}

	client, err := s.clientService.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	title := truncateInboxTitle(fmt.Sprintf("客户资料已补齐并移交：%s", client.Name))
	content := fmt.Sprintf("助理 %s 已完成客户资料补充，请继续处理客户档案与进件工作。", actorName)
	if strings.TrimSpace(req.Note) != "" {
		content += "\n\n移交说明：" + req.Note
	}
	if actorID > 0 {
		content += fmt.Sprintf("\n\n移交人ID：%d", actorID)
	}

	item, err := s.inboxService.CreateInboxItem(ctx, req.TargetUserID, &CreateInboxItemRequest{
		UserID:      req.TargetUserID,
		SourceType:  "handoff",
		SourceID:    clientID,
		Title:       title,
		Content:     content,
		Priority:    priority,
		DueDate:     req.DueDate,
		DueDateType: "client_handoff",
	})
	if err != nil {
		return nil, err
	}

	return &ClientHandoffResponse{
		ClientID:     client.ID,
		ClientName:   client.Name,
		TargetUserID: req.TargetUserID,
		TargetRole:   req.TargetRole,
		Status:       "handoff_created",
		InboxItem:    item,
	}, nil
}

func truncateInboxTitle(title string) string {
	const maxTitleLength = 255
	runes := []rune(strings.TrimSpace(title))
	if len(runes) <= maxTitleLength {
		return string(runes)
	}
	if maxTitleLength <= 3 {
		return string(runes[:maxTitleLength])
	}
	return string(runes[:maxTitleLength-3]) + "..."
}
