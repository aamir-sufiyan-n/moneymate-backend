package http

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/abijith/moneymate-backend/services/support/internal/usecase"
	"github.com/moneymate-2026/moneymate-backend/shared/pkg/responses"
)

type SupportHandler struct {
	uc usecase.SupportUseCase
}

func NewSupportHandler(uc usecase.SupportUseCase) *SupportHandler {
	return &SupportHandler{uc: uc}
}

type createFeedbackReq struct {
	Rating      int    `json:"rating"`
	Description string `json:"description"`
}

func (h *SupportHandler) CreateFeedback(c fiber.Ctx) error {
	var req createFeedbackReq
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request")
	}

	uidStr := c.Get("X-User-ID")
	userType := c.Get("X-User-Role")
	if userType == "" {
		userType = "user"
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return response.BadRequest(c, nil, "invalid user_id from token")
	}

	fb, err := h.uc.CreateFeedback(c.Context(), uid, userType, req.Rating, req.Description)
	if err != nil {
		log.Printf("[Support] CreateFeedback failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.Created(c, "Feedback created successfully", fb)
}

func (h *SupportHandler) ListFeedbacks(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	fbs, err := h.uc.ListFeedbacks(c.Context(), int32(limit), int32(offset))
	if err != nil {
		log.Printf("[Support] ListFeedbacks failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Feedbacks fetched successfully", fbs)
}

type createComplaintReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *SupportHandler) CreateComplaint(c fiber.Ctx) error {
	var req createComplaintReq
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request")
	}

	uidStr := c.Get("X-User-ID")
	userType := c.Get("X-User-Role")
	if userType == "" {
		userType = "user"
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return response.BadRequest(c, nil, "invalid user_id from token")
	}

	comp, err := h.uc.CreateComplaint(c.Context(), uid, userType, req.Title, req.Description)
	if err != nil {
		log.Printf("[Support] CreateComplaint failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.Created(c, "Complaint created successfully", comp)
}

func (h *SupportHandler) ListComplaints(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	comps, err := h.uc.ListComplaints(c.Context(), int32(limit), int32(offset))
	if err != nil {
		log.Printf("[Support] ListComplaints failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Complaints fetched successfully", comps)
}

func (h *SupportHandler) ListComplaintsByUser(c fiber.Ctx) error {
	userID := c.Get("X-User-ID")
	userType := c.Get("X-User-Role")
	if userType == "" {
		userType = "user"
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid user_id from token")
	}

	comps, err := h.uc.ListComplaintsByUser(c.Context(), uid, userType)
	if err != nil {
		log.Printf("[Support] ListComplaintsByUser failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Complaints fetched successfully", comps)
}

type createReportReq struct {
	ReportedVPA string `json:"reported_vpa"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *SupportHandler) CreateReport(c fiber.Ctx) error {
	var req createReportReq
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request")
	}

	uidStr := c.Get("X-User-ID")
	userType := c.Get("X-User-Role")
	if userType == "" {
		userType = "user"
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return response.BadRequest(c, nil, "invalid reporter_id from token")
	}

	rep, err := h.uc.CreateReport(c.Context(), uid, userType, req.ReportedVPA, req.Title, req.Description)
	if err != nil {
		log.Printf("[Support] CreateReport failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.Created(c, "Report created successfully", rep)
}

func (h *SupportHandler) ListReports(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	reps, err := h.uc.ListReports(c.Context(), int32(limit), int32(offset))
	if err != nil {
		log.Printf("[Support] ListReports failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Reports fetched successfully", reps)
}

func (h *SupportHandler) ListReportsByUser(c fiber.Ctx) error {
	reporterID := c.Get("X-User-ID")
	reporterType := c.Get("X-User-Role")
	if reporterType == "" {
		reporterType = "user"
	}

	uid, err := uuid.Parse(reporterID)
	if err != nil {
		return response.BadRequest(c, nil, "invalid reporter_id from token")
	}

	reps, err := h.uc.ListReportsByUser(c.Context(), uid, reporterType)
	if err != nil {
		log.Printf("[Support] ListReportsByUser failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Reports fetched successfully", reps)
}

type createAuditLogReq struct {
	AdminName string `json:"admin_name"`
	Module    string `json:"module"`
	Action    string `json:"action"`
}

func (h *SupportHandler) CreateAuditLog(c fiber.Ctx) error {
	var req createAuditLogReq
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request")
	}

	uidStr := c.Get("X-User-ID")
	userRole := c.Get("X-User-Role")
	if userRole == "" {
		userRole = "admin"
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return response.BadRequest(c, nil, "invalid admin_id from token")
	}

	logEntry, err := h.uc.CreateAuditLog(c.Context(), uid, req.AdminName, userRole, req.Module, req.Action)
	if err != nil {
		log.Printf("[Support] CreateAuditLog failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.Created(c, "Audit log created successfully", logEntry)
}

func (h *SupportHandler) ListAuditLogs(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logs, err := h.uc.ListAuditLogs(c.Context(), int32(limit), int32(offset))
	if err != nil {
		log.Printf("[Support] ListAuditLogs failed: %v", err)
		return response.InternalServerError(c)
	}

	return response.OK(c, "Audit logs fetched successfully", logs)
}
