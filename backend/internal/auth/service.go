package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
)

type Service struct {
	users      core.UserRepository
	clock      core.Clock
	ids        core.IDGenerator
	sessions   core.RefreshSessionRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	hashIters  int
}

type Claims struct {
	UserID core.ID `json:"sub"`
	Email  string  `json:"email"`
	Role   string  `json:"role"`
	Exp    int64   `json:"exp"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewService(users core.UserRepository, sessions core.RefreshSessionRepository, clock core.Clock, ids core.IDGenerator, secret string, accessTTL time.Duration, refreshTTL time.Duration, hashIters int) *Service {
	if hashIters <= 0 {
		hashIters = 120000
	}
	return &Service{
		users: users, sessions: sessions, clock: clock, ids: ids, jwtSecret: []byte(secret),
		accessTTL: accessTTL, refreshTTL: refreshTTL, hashIters: hashIters,
	}
}

func (s *Service) Register(ctx context.Context, email string, name string, password string) (core.User, TokenPair, error) {
	if email == "" || password == "" {
		return core.User{}, TokenPair{}, fmt.Errorf("email and password are required")
	}
	hash, err := HashPassword(password, s.hashIters)
	if err != nil {
		return core.User{}, TokenPair{}, err
	}
	now := s.clock.Now()
	user, err := s.users.CreateUser(ctx, core.User{
		ID: s.ids.NewID(), Email: strings.ToLower(email), Name: name, Role: "user",
		PasswordHash: hash, Active: true, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return core.User{}, TokenPair{}, err
	}
	tokens, err := s.issueTokens(ctx, user)
	return user, tokens, err
}

func (s *Service) Login(ctx context.Context, email string, password string) (core.User, TokenPair, error) {
	user, err := s.users.FindUserByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return core.User{}, TokenPair{}, fmt.Errorf("invalid credentials")
	}
	if !user.Active || !VerifyPassword(password, user.PasswordHash) {
		return core.User{}, TokenPair{}, fmt.Errorf("invalid credentials")
	}
	tokens, err := s.issueTokens(ctx, user)
	return user, tokens, err
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	session, err := s.sessions.FindRefreshSession(ctx, refreshToken)
	if err != nil || session.RevokedAt != nil || session.ExpiresAt.Before(s.clock.Now()) {
		return TokenPair{}, fmt.Errorf("refresh token expired")
	}
	user, err := s.users.FindUserByID(ctx, session.UserID)
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.sessions.RevokeRefreshSession(ctx, refreshToken, s.clock.Now()); err != nil {
		return TokenPair{}, err
	}
	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.sessions.RevokeRefreshSession(ctx, refreshToken, s.clock.Now())
}

func (s *Service) ValidateAccessToken(token string) (Claims, error) {
	claims, err := ParseJWT(token, s.jwtSecret)
	if err != nil {
		return Claims{}, err
	}
	if claims.Exp < s.clock.Now().Unix() {
		return Claims{}, fmt.Errorf("access token expired")
	}
	return claims, nil
}

func (s *Service) issueTokens(ctx context.Context, user core.User) (TokenPair, error) {
	expiresAt := s.clock.Now().Add(s.accessTTL)
	access, err := SignJWT(Claims{UserID: user.ID, Email: user.Email, Role: user.Role, Exp: expiresAt.Unix()}, s.jwtSecret)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := randomToken()
	if err != nil {
		return TokenPair{}, err
	}
	now := s.clock.Now()
	if err := s.sessions.SaveRefreshSession(ctx, core.RefreshSession{Token: refresh, UserID: user.ID, Role: user.Role, ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now}); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}

func SignJWT(claims Claims, secret []byte) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func ParseJWT(token string, secret []byte) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, fmt.Errorf("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(expected, actual) {
		return Claims{}, fmt.Errorf("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	return claims, json.Unmarshal(payload, &claims)
}

func HashPassword(password string, iterations int) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	digest := passwordDigest([]byte(password), salt, iterations)
	return "sha256:" + strconv.Itoa(iterations) + ":" + base64.RawStdEncoding.EncodeToString(salt) + ":" + base64.RawStdEncoding.EncodeToString(digest), nil
}

func VerifyPassword(password string, encoded string) bool {
	parts := strings.Split(encoded, ":")
	if len(parts) != 4 || parts[0] != "sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	return hmac.Equal(passwordDigest([]byte(password), salt, iterations), expected)
}

func passwordDigest(password []byte, salt []byte, iterations int) []byte {
	digest := append([]byte{}, salt...)
	for i := 0; i < iterations; i++ {
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(digest)
		digest = mac.Sum(nil)
	}
	return digest
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
