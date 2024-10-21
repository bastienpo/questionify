package data

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidHash         = errors.New("invalid hash")
	ErrIncompatibleVersion = errors.New("incompatible version")
)

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateFromPassword(plaintext string, p argon2Params) ([]byte, error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return nil, err
	}

	hash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)

	// To store the hash with the salt, we encode them as base64 and format them as a string
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash,
	)

	return []byte(encodedHash), nil
}

func comparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters from the encoded hash
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (p *argon2Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"password"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := generateFromPassword(plaintext, argon2Params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	})
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintext string) (bool, error) {
	match, err := comparePasswordAndHash(plaintext, string(p.hash))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidHash):
			return false, nil
		case errors.Is(err, ErrIncompatibleVersion):
			return false, ErrIncompatibleVersion
		default:
			return false, err
		}
	}

	return match, nil
}
