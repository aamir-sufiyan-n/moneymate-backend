package hasher

import "github.com/moneymate-2026/moneymate-backend/shared/pkg/hash"

type Argon2Hasher struct{}

func New() *Argon2Hasher {
    return &Argon2Hasher{}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
    return hashpass.HashPassword(password)
}

func (h *Argon2Hasher) Verify(hash, password string) (bool, error) {
    return hashpass.VerifyPassword(hash, password)
}