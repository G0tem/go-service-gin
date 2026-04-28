package integration

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testrabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/G0tem/go-service-gin/internal/domain/order"
	pginfra "github.com/G0tem/go-service-gin/internal/infra/postgres"
	rmqinfra "github.com/G0tem/go-service-gin/internal/infra/rabbitmq"
	usecase "github.com/G0tem/go-service-gin/internal/usecases"
)

func TestCreateOrderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// 🗄️ Postgres
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithLogger(nil),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// 🐇 RabbitMQ
	rmqContainer, err := testrabbitmq.RunContainer(ctx,
		testcontainers.WithImage("rabbitmq:3-management-alpine"),
		testcontainers.WithLogger(nil),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rmqContainer.Terminate(ctx) })

	rmqURL, err := rmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	// 🛠️ Init Postgres
	pool, err := pginfra.NewPool(ctx, pgDSN)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			amount NUMERIC(10,2) NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	require.NoError(t, err)

	// 🛠️ Init RabbitMQ client (instrumented, for service)
	rmqClient, err := rmqinfra.NewClient(ctx, rmqURL)
	require.NoError(t, err)
	t.Cleanup(func() { rmqClient.Close() })

	// 🔧 RAW connection for test assertions (NOT instrumented)
	rawConn, err := amqp.Dial(rmqURL)
	require.NoError(t, err)
	defer rawConn.Close()

	rawCh, err := rawConn.Channel()
	require.NoError(t, err)
	defer rawCh.Close()

	// 📡 Consumer setup using RAW channel
	q, err := rawCh.QueueDeclare("test-order-events", false, true, true, false, nil)
	require.NoError(t, err)

	err = rawCh.QueueBind(q.Name, "order.created", "", false, nil)
	require.NoError(t, err)

	msgs, err := rawCh.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// ⚙️ UseCase (metrics = nil ok for test)
	uc := usecase.NewCreateOrderHandler(
		pginfra.NewOrderRepo(pool),
		rmqClient, // ← instrumented client for publishing
		nil,
	)

	cmd := order.CreateOrderCmd{UserID: "usr_int_test", Amount: 199.99}
	err = uc.Handle(ctx, cmd)
	require.NoError(t, err)

	// ✅ Assert DB
	var id, status string
	var amount float64
	err = pool.QueryRow(ctx, `SELECT id, status, amount FROM orders WHERE user_id = $1`, cmd.UserID).
		Scan(&id, &status, &amount)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, "pending", status)
	assert.Equal(t, cmd.Amount, amount)

	// ✅ Assert MQ (using RAW channel)
	select {
	case msg := <-msgs:
		assert.Equal(t, "order.created", msg.RoutingKey)
		assert.Contains(t, string(msg.Body), "usr_int_test")
		assert.Contains(t, string(msg.Body), "199.99")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for RabbitMQ message")
	}
}
