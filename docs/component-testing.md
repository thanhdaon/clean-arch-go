# Component Testing Guide

Component tests verify complete business cases through your public API (HTTP/gRPC) while stubbing external dependencies.

## Test Types Comparison

| Feature / Test Type           | Unit        | Integration | Component        | End-to-End |
|-------------------------------|-------------|-------------|------------------|------------|
| **Docker database**           | No          | Yes         | Yes              | Yes        |
| **Use of external systems**   | No          | No          | No               | Yes        |
| **Focused on business cases** | Varies      | No          | Yes              | Yes        |
| **Uses mocks or stubs**       | Most deps   | Usually none| External systems | None       |
| **Tested API**                | Go package  | Go package  | HTTP and gRPC    | HTTP       |
| **Execution speed**           | Fast        | Fast        | Medium           | Slow       |

- **Integration tests**: Verify adapters work with infrastructure you control (e.g., "Does my `TaskRepository` correctly store tasks in MySQL?")
- **Component tests**: Verify business flows end-to-end within your service (e.g., "When I create a task via HTTP, is the notification sent?")

## Architecture

### Production Infrastructure
```
┌─────────────────────────────────────────────────────────────┐
│                     Your Infrastructure                      │
│  ┌─────────────────┐                  ┌─────────────────┐  │
│  │    Service      │─────────────────▶│     MySQL       │  │
│  └────────┬────────┘                  └─────────────────┘  │
│           │                                                  │
└───────────┼──────────────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────┐
│                    External Systems                            │
│  ┌─────────────────┐    ┌─────────────────┐                   │
│  │   Notification  │    │     Email       │    ...            │
│  │     Service     │    │    Service      │                   │
│  └─────────────────┘    └─────────────────┘                   │
└───────────────────────────────────────────────────────────────┘
```

### Test Infrastructure
```
┌─────────────────────────────────────────────────────────────┐
│                         Test Suite                           │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service (real)                           │
│  ┌─────────────────┐                  ┌─────────────────┐  │
│  │    Service      │─────────────────▶│  MySQL (Docker) │  │
│  └────────┬────────┘                  └─────────────────┘  │
└───────────┼─────────────────────────────────────────────────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────┐
│                    Stub Implementations                        │
│  ┌─────────────────┐    ┌─────────────────┐                   │
│  │   Notification  │    │     Email       │    ...            │
│  │   Service Stub  │    │   Service Stub  │                   │
│  └─────────────────┘    └─────────────────┘                   │
└───────────────────────────────────────────────────────────────┘
```

## Implementation Guide

### 1. Define Test Fixtures

```go
type TestFixtures struct {
    DB *sqlx.DB
    
    // Stub adapters for external services
    NotificationService *NotificationServiceStub
    EmailService        *EmailServiceStub
    
    Cancel context.CancelFunc
}
```

### 2. Create Stub Implementations

Use stubs, not mocks — they're simpler and test behavior rather than method calls.

```go
type NotificationServiceStub struct {
    lock        sync.Mutex
    sentNotifications []SendNotificationRequest
}

func NewNotificationServiceStub() *NotificationServiceStub {
    return &NotificationServiceStub{
        sentNotifications: make([]SendNotificationRequest, 0),
    }
}

func (s *NotificationServiceStub) SendNotification(
    ctx context.Context,
    request SendNotificationRequest,
) error {
    s.lock.Lock()
    defer s.lock.Unlock()
    
    s.sentNotifications = append(s.sentNotifications, request)
    return nil
}

func (s *NotificationServiceStub) SentCount() int {
    s.lock.Lock()
    defer s.lock.Unlock()
    return len(s.sentNotifications)
}

func (s *NotificationServiceStub) FindByUserID(userID string) (SendNotificationRequest, bool) {
    s.lock.Lock()
    defer s.lock.Unlock()
    
    for _, n := range s.sentNotifications {
        if n.UserID == userID {
            return n, true
        }
    }
    return SendNotificationRequest{}, false
}
```

### 3. Setup Function

```go
func SetupComponentTest(t *testing.T) *TestFixtures {
    db, err := sqlx.Open("mysql", os.Getenv("MYSQL_URL"))
    require.NoError(t, err)
    
    ctx, cancel := context.WithCancel(context.Background())
    
    notificationService := NewNotificationServiceStub()
    emailService := NewEmailServiceStub()
    
    go func() {
        svc, err := service.New(db, notificationService, emailService)
        if err != nil {
            t.Errorf("failed to create service: %v", err)
            return
        }
        err = svc.Run(ctx)
        assert.NoError(t, err)
    }()
    
    waitForHttpServer(t)
    
    return &TestFixtures{
        DB:                  db,
        NotificationService: notificationService,
        EmailService:        emailService,
        Cancel:              cancel,
    }
}

func waitForHttpServer(t *testing.T) {
    t.Helper()
    
    condition := func(t *assert.CollectT) {
        resp, err := http.Get("http://localhost:8080/health")
        if !assert.NoError(t, err) {
            return
        }
        defer resp.Body.Close()
        
        assert.Less(t, resp.StatusCode, 300)
    }
    
    require.EventuallyWithT(t, condition, 10*time.Second, 50*time.Millisecond)
}
```

### 4. Entry Point Test

```go
func TestComponent(t *testing.T) {
    fixtures := SetupComponentTest(t)
    defer fixtures.Cancel()
    
    t.Run("create_task", func(t *testing.T) {
        testCreateTask(t, fixtures)
    })
    
    t.Run("complete_task", func(t *testing.T) {
        testCompleteTask(t, fixtures)
    })
    
    t.Run("delete_task", func(t *testing.T) {
        testDeleteTask(t, fixtures)
    })
}
```

### 5. Helper Functions

```go
// HTTP helpers
func createTask(t *testing.T, req CreateTaskRequest) (*http.Response, []byte) {
    t.Helper()
    
    payload, err := json.Marshal(req)
    require.NoError(t, err)
    
    httpReq, err := http.NewRequest(
        http.MethodPost,
        "http://localhost:8080/api/tasks",
        bytes.NewBuffer(payload),
    )
    require.NoError(t, err)
    
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(httpReq)
    require.NoError(t, err)
    
    body, err := io.ReadAll(resp.Body)
    require.NoError(t, err)
    resp.Body.Close()
    
    return resp, body
}

// Assertion helpers
func assertTaskStoredInDB(t *testing.T, db *sqlx.DB, taskID string) {
    t.Helper()
    
    var count int
    err := db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID)
    require.NoError(t, err)
    assert.Equal(t, 1, count, "task should be stored in database")
}

func assertNotificationSent(t *testing.T, stub *NotificationServiceStub, userID string) {
    t.Helper()
    
    notification, ok := stub.FindByUserID(userID)
    assert.True(t, ok, "notification should be sent to user %s", userID)
    assert.Equal(t, userID, notification.UserID)
}
```

### 6. Test Cases

```go
func testCreateTask(t *testing.T, fixtures *TestFixtures) {
    userID := "user-123"
    
    resp, body := createTask(t, CreateTaskRequest{
        Title:       "Test Task",
        Description: "Test Description",
        UserID:      userID,
    })
    
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var result CreateTaskResponse
    err := json.Unmarshal(body, &result)
    require.NoError(t, err)
    assert.NotEmpty(t, result.TaskID)
    
    assertTaskStoredInDB(t, fixtures.DB, result.TaskID)
    assertNotificationSent(t, fixtures.NotificationService, userID)
}

func testCreateTaskIdempotency(t *testing.T, fixtures *TestFixtures) {
    idempotencyKey := uuid.NewString()
    userID := "user-456"
    
    for i := 0; i < 3; i++ {
        resp, _ := createTaskWithIdempotencyKey(t, CreateTaskRequest{
            Title:  "Idempotent Task",
            UserID: userID,
        }, idempotencyKey)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
    }
    
    count := fixtures.NotificationService.SentCount()
    assert.Equal(t, 1, count, "notification should be sent exactly once")
}
```

## Best Practices

### What to Test
- **Happy paths** in component tests
- **Edge cases** in unit tests
- **Critical scenarios** in both component and E2E tests

### Stubs vs Mocks
- **Avoid generated mocks** — fragile and test implementation details
- **Use hand-written stubs** — flexible, reusable, test behavior

### Cleanup-Independent Tests

**Use unique identifiers** instead of shared state:

```go
// DON'T - Depends on cleanup working
func TestUserCreation(t *testing.T) {
    CreateUser("test-user")
    // If test crashes, "test-user" lingers and conflicts with next run
}

// DO - Use unique IDs
func TestUserCreation(t *testing.T) {
    userID := fmt.Sprintf("test-user-%s", uuid.New())
    CreateUser(userID)
    // If cleanup fails, next run gets a new unique ID
}
```

**Use transactions** that auto-rollback:

```go
func TestWithTransaction(t *testing.T) {
    tx := db.Begin()
    defer tx.Rollback() // Even if we crash, DB rolls back

    // ... test code using tx instead of db
}
```

**Use ephemeral resources** that self-destruct:

```go
func TestWithContainer(t *testing.T) {
    ctx := context.Background()
    container, _ := testcontainers.GenericContainer(ctx, ...)
    defer container.Terminate(ctx) // Nice-to-have, not required

    // Container dies when test process exits anyway
}
```

**Accept messiness in external services** — test data accumulating in dev/staging is fine. Query for fresh data rather than depend on a clean slate.

### Thread Safety

```go
type Stub struct {
    lock   sync.Mutex
    inputs []Input
}

func (s *Stub) DoStuff(ctx context.Context, input Input) error {
    s.lock.Lock()
    defer s.lock.Unlock()
    
    s.inputs = append(s.inputs, input)
    return nil
}
```

### Database Isolation
- Use unique identifiers per test
- Clean up data between tests or use transactions
- Use targeted assertions ("this specific row exists" vs "exactly one row exists")

## Running Component Tests

```bash
docker-compose up -d
go test ./tests/... -v
go test ./tests/... -race
```

## Directory Structure

```
tests/
├── component_test.go    # Entry point, organizes all tests
├── setup_test.go        # TestFixtures and setup logic
├── helpers_test.go      # HTTP helpers and assertion helpers
├── task_test.go         # Task-related test scenarios
└── user_test.go         # User-related test scenarios
```
