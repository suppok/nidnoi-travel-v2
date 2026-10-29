package password

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"backend/pkg/core/validation"
)

// Rule mirrored in frontend/backoffice/src/lib/password.ts.
const (
	MinLength = 8
	MaxBytes  = 72 // bcrypt input limit
)

// Service hashes and checks staff passwords. Used by the auth and user modules.
type Service interface {
	Hash(plain string) (string, error)
	// Matches reports whether plain matches hash. With an empty hash it still spends the bcrypt
	// time, so unknown emails can't be told apart by response time.
	Matches(hash, plain string) bool
	// Rule validates the password policy inside a use case's Validate.
	Rule() validation.Rule
	// GenerateTemporary returns a random password that satisfies the policy.
	GenerateTemporary() (string, error)
}

type service struct {
	cost      int
	dummyHash []byte
}

func NewService(cost int) (Service, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("timing-equalizer-not-a-password"), cost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}
	return &service{cost: cost, dummyHash: dummy}, nil
}

func (s *service) Hash(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), s.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(h), nil
}

func (s *service) Matches(hash, plain string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(plain))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func (s *service) Rule() validation.Rule {
	return validation.By(func(value any) error {
		v, _ := value.(string)
		return Problem(v)
	})
}

// Problem returns the policy violation of v, or nil.
func Problem(v string) error {
	if len([]rune(v)) < MinLength {
		return fmt.Errorf("must be at least %d characters", MinLength)
	}
	if len(v) > MaxBytes {
		return fmt.Errorf("must be at most %d bytes", MaxBytes)
	}
	var letter, digit bool
	for _, r := range v {
		switch {
		case r < unicode.MaxASCII && unicode.IsLetter(r):
			letter = true
		case r < unicode.MaxASCII && unicode.IsDigit(r):
			digit = true
		}
	}
	if !letter || !digit {
		return errors.New("must contain both an English letter and a digit")
	}
	return nil
}

// No look-alike characters (0/O, 1/l/I): staff read these out or type them from a chat message.
const temporaryAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func (s *service) GenerateTemporary() (string, error) {
	for {
		buf := make([]byte, 12)
		for i := range buf {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(temporaryAlphabet))))
			if err != nil {
				return "", fmt.Errorf("random: %w", err)
			}
			buf[i] = temporaryAlphabet[n.Int64()]
		}
		if Problem(string(buf)) == nil {
			return string(buf), nil
		}
	}
}
