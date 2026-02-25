package tests_test

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"clean-arch-go/adapters"
	"clean-arch-go/bootstrap"
	"clean-arch-go/core/auth"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type VideoServiceStub struct {
	mu    sync.Mutex
	calls int
}

func (s *VideoServiceStub) GetAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return nil
}

func (s *VideoServiceStub) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// CallCountSnapshot returns the current call count for use as a baseline.
// Compare against a later CallCount() to assert exactly how many calls
// a specific operation made, independent of prior test runs.
func (s *VideoServiceStub) CallCountSnapshot() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

type TestFixtures struct {
	DB           *sqlx.DB
	VideoService *VideoServiceStub
	Cancel       context.CancelFunc
	AuthToken    string
}

func SetupComponentTest(t *testing.T) *TestFixtures {
	t.Helper()

	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	videoStub := &VideoServiceStub{}

	logger := logrus.NewEntry(logrus.New())

	svc, err := bootstrap.New(db, nil, logger, videoStub)
	require.NoError(t, err)

	go func() {
		err := svc.Run(ctx)
		assert.NoError(t, err)
	}()

	waitForHTTPServer(t)

	createAdminUser(t, db)
	token := generateTestToken(t)

	return &TestFixtures{
		DB:           db,
		VideoService: videoStub,
		Cancel:       cancel,
		AuthToken:    token,
	}
}

func waitForHTTPServer(t *testing.T) {
	t.Helper()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	condition := func(t *assert.CollectT) {
		resp, err := http.Get("http://localhost:" + port + "/api/users")
		if !assert.NoError(t, err) {
			return
		}
		defer resp.Body.Close()
		assert.Less(t, resp.StatusCode, 500)
	}

	require.EventuallyWithT(t, condition, 10*time.Second, 50*time.Millisecond)
}

func generateTestToken(t *testing.T) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret-key-for-development"
	}

	a := auth.NewAuth(secret)
	token, err := a.CreateIDToken(map[string]any{
		"user_uuid": "test-admin-uuid",
		"user_role": "admin",
	})
	require.NoError(t, err)
	return token
}

func createAdminUser(t *testing.T, db *sqlx.DB) string {
	t.Helper()

	email := "test-admin@example.com"
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE email = ?", email)
	require.NoError(t, err)

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO users (id, role, name, email, password_hash)
			VALUES ('test-admin-uuid', 'admin', 'Test Admin', 'test-admin@example.com', 'hashed-password')
		`)
		require.NoError(t, err)
	}

	return "test-admin-uuid"
}
