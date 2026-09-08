package http

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/abijith/moneymate-backend/services/support/internal/usecase"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

type ChatHandler struct {
	useCase usecase.ChatUseCase
}

func NewChatHandler(useCase usecase.ChatUseCase) *ChatHandler {
	return &ChatHandler{useCase: useCase}
}

type SendMessageRequest struct {
	ReceiverID   string `json:"receiver_id"`
	ReceiverType string `json:"receiver_type"`
	Message      string `json:"message"`
}

func (h *ChatHandler) SendMessage(c fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "Invalid request payload")
	}

	senderIDStr := c.Get("X-User-ID")
	senderType := c.Get("X-User-Role")
	if senderType == "" {
		senderType = "user"
	}

	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid sender ID")
	}

	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid receiver ID")
	}

	msg, err := h.useCase.SendMessage(c.Context(), senderID, senderType, receiverID, req.ReceiverType, req.Message)
	if err != nil {
		log.Printf("[Support] SendMessage failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Message sent successfully", msg)
}

func (h *ChatHandler) GetChatHistoryForAdmin(c fiber.Ctx) error {
	adminIDStr := c.Get("X-User-ID")
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid admin ID")
	}

	userIDStr := c.Params("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid user ID")
	}

	history, err := h.useCase.GetChatHistory(c.Context(), adminID, userID)
	if err != nil {
		log.Printf("[Support] GetChatHistoryForAdmin failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Admin chat history with user retrieved", history)
}

func (h *ChatHandler) GetAdminChatHistory(c fiber.Ctx) error {
	adminIDStr := c.Get("X-User-ID")
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid admin ID")
	}

	history, err := h.useCase.GetAdminChatHistory(c.Context(), adminID)
	if err != nil {
		log.Printf("[Support] GetAdminChatHistory failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Admin chat history retrieved", history)
}

func (h *ChatHandler) MarkMessagesAsRead(c fiber.Ctx) error {
	senderIDStr := c.Params("sender_id")
	senderID, err := uuid.Parse(senderIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid sender ID")
	}

	receiverIDStr := c.Get("X-User-ID")
	receiverID, err := uuid.Parse(receiverIDStr)
	if err != nil {
		return response.BadRequest(c, nil, "Invalid receiver ID")
	}

	if err := h.useCase.MarkMessagesAsRead(c.Context(), senderID, receiverID); err != nil {
		log.Printf("[Support] MarkMessagesAsRead failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Messages marked as read", nil)
}
