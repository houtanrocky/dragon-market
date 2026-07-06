package guild

import "errors"

var (
	ErrGuildNotFound       = errors.New("guild not found")
	ErrGuildAlreadyExists  = errors.New("guild already exists")
	ErrInvalidGuild        = errors.New("guild is invalid")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrInsufficientBalance = errors.New("insufficient available balance")
	ErrInsufficientReserve = errors.New("insufficient reserved balance")
	ErrDailyLimitReached   = errors.New("daily spending limit reached")
)
