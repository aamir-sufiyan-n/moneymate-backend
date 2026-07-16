package idgen


import (
    "github.com/google/uuid"

    sharedUUID "github.com/moneymate-2026/moneymate-backend/shared/pkg/uuid"
)
type Generator struct{}

func New() *Generator {
    return &Generator{}
}
func (g *Generator) NewV7() (uuid.UUID, error) {
    s, err := sharedUUID.New()
    if err != nil {
        return uuid.UUID{}, err
    }
    return uuid.Parse(s)
}