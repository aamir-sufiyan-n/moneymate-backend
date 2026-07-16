package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"fmt"
	"math/big"
	"strings"

	"github.com/moneymate-2026/moneymate-backend/auth/config"
	"github.com/moneymate-2026/moneymate-backend/auth/internal/domain"
	apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

type EmailSender interface {
	SendOTP(ctx context.Context, toEmail, otp string) error
}

type OTPUsecase interface {
	SendRegistrationOTP(ctx context.Context, req SendRegistrationOTPRequest) (*SendRegistrationOTPResponse, error)
	VerifyRegistrationOTP(ctx context.Context, req VerifyRegistrationOTPRequest) (*VerifyRegistrationOTPResponse, error)
}

type otpUsecase struct {
	userRepo domain.UserRepository
	store    domain.Store
	mailer   EmailSender
	cfg      config.OTPConfig
}

func NewOTPUsecase(userRepo domain.UserRepository, store domain.Store, mailer EmailSender, conf config.OTPConfig) OTPUsecase {
	return &otpUsecase{
		userRepo: userRepo,
		store:    store,
		mailer:   mailer,
		cfg:      conf,
	}
}

func (u *otpUsecase) SendRegistrationOTP(ctx context.Context, req SendRegistrationOTPRequest) (*SendRegistrationOTPResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		return nil, apperrors.ErrInvalidInput
	}
	exists, err := u.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if exists {
		return nil, apperrors.ErrEmailAlreadyTaken
	}

	allowed, remainingTime, err := u.store.TrySetResendCooldown(ctx, email, u.cfg.ResendCooldown)
	if err != nil {
		return nil, fmt.Errorf("check resend cooldown: %w", err)
	}
	if !allowed {
		details := map[string]interface{}{
			"retry_after_seconds": int(remainingTime.Seconds()),
		}
		return nil, apperrors.NewAppErrorWithDetails(
			429,
			"OTP_COOLDOWN",
			"Please wait before requesting another code.",
			details,
			err,
		)
	}

	code, err := generateOTP(u.cfg.Length)
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}

	if err := u.store.SetRegistrationOTP(ctx, email, hashOTP(code), u.cfg.TTL); err != nil {
		return nil, fmt.Errorf("store otp: %w", err)
	}
	if err := u.store.ResetOTPAttempts(ctx, email); err != nil {
		return nil, fmt.Errorf("reset otp attempts: %w", err)
	}

	if err := u.mailer.SendOTP(ctx, email, code); err != nil {
		return nil, fmt.Errorf("send otp email: %w", err)
	}

	return &SendRegistrationOTPResponse{
		Email:             email,
		ExpiresIn:         int(u.cfg.TTL.Seconds()),
		ResendCooldownIn:  int(u.cfg.ResendCooldown.Seconds()),
		MaxVerifyAttempts: u.cfg.MaxVerifyAttempts,
	}, nil
}

func (u *otpUsecase) VerifyRegistrationOTP(ctx context.Context, req VerifyRegistrationOTPRequest) (*VerifyRegistrationOTPResponse, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)

	if email == "" || code == "" {
		return nil, apperrors.ErrInvalidInput
	}
	attempts, err := u.store.IncrementOTPAttempts(ctx, email, u.cfg.TTL)
	if err != nil {
		return nil, fmt.Errorf("increment otp attempts: %w", err)
	}
	if attempts > int64(u.cfg.MaxVerifyAttempts) {
		_ = u.store.DeleteRegistrationOTP(ctx, email)
		return nil, apperrors.ErrOTPTimout
	}

	storedHash, found, err := u.store.GetRegistrationOTP(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get registration otp: %w", err)
	}
	if !found {
		return nil, apperrors.ErrOTPExpired
	}

	suppliedHash := hashOTP(code)
	if !constantTimeEqual(storedHash, suppliedHash) {
		attemptsLeft := int64(u.cfg.MaxVerifyAttempts) - attempts
		if attemptsLeft < 0 {
			attemptsLeft = 0
		}

		details := map[string]interface{}{
			"attempts_left": int(attemptsLeft),
			"max_attempts":  u.cfg.MaxVerifyAttempts,
		}

		return nil, apperrors.NewAppErrorWithDetails(
			400,
			"OTP_INVALID",
			"The code you entered is incorrect.",
			details,
			nil,
		)
	}
	if err := u.store.DeleteRegistrationOTP(ctx, email); err != nil {
		return nil, fmt.Errorf("delete registration otp: %w", err)
	}
	if err := u.store.ResetOTPAttempts(ctx, email); err != nil {
		return nil, fmt.Errorf("reset otp attempts: %w", err)
	}
	if err := u.store.MarkEmailVerified(ctx, email, u.cfg.EmailVerifiedTTL); err != nil {
		return nil, fmt.Errorf("mark email verified: %w", err)
	}

	return &VerifyRegistrationOTPResponse{
		Email:    email,
		Verified: true,
	}, nil
}

// ── Helpers ────────────────────────────────────────────────────

func generateOTP(digits int) (string, error) {
	var sb strings.Builder
	for i := 0; i < digits; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteString(d.String())
	}
	return sb.String(), nil
}

func hashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
