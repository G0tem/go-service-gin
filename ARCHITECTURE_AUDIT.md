# Архитектурный аудит Go-сервиса

## Текущее состояние

### ✅ Сильные стороны
1. **Слоистая архитектура**: Четкое разделение на domain/usecases/http/infra
2. **Dependency Injection**: Зависимости передаются через конструкторы
3. **Порты и адаптеры**: Интерфейсы в domain/ports, реализации в infra
4. **OpenTelemetry**: Интеграция трейсинга и метрик
5. **Конфигурация**: Валидация конфигов при загрузке

### ❌ Критические проблемы

## 1. Нарушение принципов DDD

### Проблема 1.1: Анемичная доменная модель
**Файл**: `internal/domain/order/entity.go`

```go
type Order struct {
    ID        string
    UserID    string
    Amount    float64  // ❌ Должен быть Value Object
    Status    string   // ❌ Должен быть типизированный enum
    CreatedAt time.Time
}
```

**Почему это плохо**:
- Entity не содержит бизнес-логики
- Нет инвариантов (можно создать заказ с amount = -100)
- Status как string позволяет невалидные значения

**Решение**:
```go
package order

type Order struct {
    id        OrderID      // Value Object
    userID    UserID       // Value Object  
    amount    Money        // Value Object с валидацией
    status    OrderStatus  // Типизированный статус
    createdAt time.Time
}

// Factory method с валидацией
func NewOrder(userID UserID, amount Money) (*Order, error) {
    if amount.IsNegative() {
        return nil, ErrInvalidAmount
    }
    return &Order{
        id:        NewOrderID(),
        userID:    userID,
        amount:    amount,
        status:    StatusPending,
        createdAt: time.Now(),
    }, nil
}

// Методы агрегата
func (o *Order) Confirm() error {
    if o.status != StatusPending {
        return ErrInvalidStatusTransition
    }
    o.status = StatusConfirmed
    return nil
}
```

### Проблема 1.2: Отсутствие агрегатных границ
**Что нужно**:
- Определить Aggregate Root (Order)
- Выделить Value Objects (Money, OrderID, UserID, OrderStatus)
- Добавить Domain Events (OrderCreated, OrderConfirmed)

## 2. Проблемы с пакетной структурой

### Проблема 2.1: Конфиг в internal
**Файл**: `internal/config/config.go`

**Проблема**: Если конфиг понадобится в других сервисах или в тестах вне internal - будут сложности.

**Рекомендация**:
```
/workspace
├── pkg/
│   └── config/          # ← Вынести сюда
│       └── config.go
├── internal/
│   └── ...
```

### Проблема 2.2: Отсутствует pkg для shared кода
**Что должно быть в pkg**:
- Конфигурация (если переиспользуется)
- Общие middleware
- HTTP клиенты для внешних сервисов
- Utility функции

## 3. Use Case слой

### Проблема 3.1: Handler вместо Service
**Файл**: `internal/usecases/create_order.go`

```go
type CreateOrderHandler struct {  // ❌ Называется Handler, но это Service
    repo    ports.OrderRepository
    pub     ports.EventPublisher
    metrics *otel.BusinessMetrics
}
```

**Рекомендация**:
```go
package orderapp  // или application

type CreateOrderService struct {
    repo    ports.OrderRepository
    pub     ports.EventPublisher
    metrics MetricsPort  // ← Абстракция, а не конкретная реализация
}

func (s *CreateOrderService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (OrderID, error) {
    // Логика
}
```

### Проблема 3.2: Зависимость от конкретной реализации OTel
```go
metrics *otel.BusinessMetrics  // ❌ Конкретная реализация
```

**Должно быть**:
```go
metrics MetricsPort  // ← Интерфейс из domain/ports
```

## 4. Инфраструктурный слой

### Проблема 4.1: Репозиторий без unit of work
**Файл**: `internal/infra/postgres/repo.go`

```go
func (r *OrderRepo) Create(ctx context.Context, o order.Order) error {
    // ❌ Нет транзакций, нет unit of work
}
```

**Рекомендация**:
```go
type UnitOfWork interface {
    Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

### Проблема 4.2: Event Publisher без Outbox pattern
**Файл**: `internal/usecases/create_order.go`

```go
if err := h.pub.Publish(ctx, "order.created", payload); err != nil {
    // В production: Outbox / компенсация  ← Комментарий, но не реализовано!
}
```

**Рекомендация**: Реализовать Transactional Outbox pattern

## 5. HTTP слой

### Проблема 5.1: Handler знает про middleware
**Файл**: `internal/http/handlers/order.go`

```go
claims := middleware.GetClaims(c)  // ❌ Зависимость от реализации middleware
```

**Рекомендация**:
```go
type AuthContext interface {
    GetUserID(ctx context.Context) (string, error)
}

// Передавать UserID явно через context или как параметр
func (h *OrderHandler) Create(c *gin.Context) {
    userID := auth.GetUserIDFromContext(c.Request.Context())
    cmd.UserID = userID
}
```

## 6. Рекомендуемая структура проекта

```
/workspace
├── cmd/
│   └── server/
│       └── main.go              # Composition root
├── pkg/                          # ← НОВЫЙ ПАКЕТ (public)
│   ├── config/
│   │   └── config.go            # Конфигурация
│   ├── http/
│   │   ├── middleware/          # Общие middleware
│   │   └── response/            # HTTP response helpers
│   └── otel/
│       └── metrics.go           # Метрики (если переиспользуются)
├── internal/
│   ├── domain/
│   │   ├── order/
│   │   │   ├── entity.go        # Order aggregate root
│   │   │   ├── value_object.go  # Money, OrderID, etc.
│   │   │   ├── repository.go    # OrderRepository interface
│   │   │   ├── service.go       # Domain service interfaces
│   │   │   └── event.go         # Domain events
│   │   ├── user/
│   │   └── ports/               # ← Удалить, перенести в соответствующие пакеты
│   ├── application/             # ← Переименовать из usecases
│   │   ├── order/
│   │   │   ├── service.go       # Application services
│   │   │   ├── command.go       # Commands
│   │   │   └── query.go         # Queries (CQRS)
│   │   └── ports/               # Secondary ports (интерфейсы)
│   ├── infrastructure/          # ← Переименовать из infra
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── order_repo.go
│   │   │   │   └── transaction.go
│   │   │   └── redis/
│   │   ├── messaging/
│   │   │   └── rabbitmq/
│   │   └── auth/
│   │       └── jwt/
│   ├── delivery/                # ← Переименовать из http
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   └── grpc/                # На будущее
│   └── integration/             # Integration tests
└── docs/
```

## 7. Priority Refactoring Plan

### Phase 1 (Critical)
1. ✅ Создать Value Objects для Money, OrderID, OrderStatus
2. ✅ Добавить factory methods для Order с валидацией
3. ✅ Вынести config в pkg/config
4. ✅ Переименовать usecases → application

### Phase 2 (High)
5. ✅ Добавить Domain Events
6. ✅ Реализовать Unit of Work pattern
7. ✅ Абстрагировать метрики через интерфейс
8. ✅ Реализовать Transactional Outbox

### Phase 3 (Medium)
9. ✅ Добавить CQRS (отдельно команды/запросы)
10. ✅ Рефакторинг HTTP handlers
11. ✅ Добавить интеграционные тесты с testcontainers

## 8. Ответ на вопрос про pkg

**Да, вы правы!** Пакет `pkg` стоит вынести на один уровень с `internal`:

**Когда использовать pkg**:
- Код, который может переиспользоваться в других сервисах
- Конфигурация (если общая для нескольких сервисов)
- HTTP клиенты к внешним API
- Общие middleware и утилиты

**Когда оставлять в internal**:
- Бизнес-логика конкретного сервиса
- Доменные модели
- Инфраструктурные реализации специфичные для сервиса

**В вашем случае**:
- `config` → `pkg/config` ✅
- `otel` → можно оставить в internal, если не переиспользуется
- Создать `pkg/http/middleware` для общих middleware

## Итоговая оценка

| Критерий | Оценка | Комментарий |
|----------|--------|-------------|
| Clean Architecture | 6/10 | Слои есть, но нарушения Dependency Rule |
| DDD | 4/10 | Анемичная модель, нет агрегатов |
| Package Design | 5/10 | Нет pkg, смешанные ответственности |
| Error Handling | 7/10 | Есть валидация, но нет domain errors |
| Testing | 3/10 | Минимум тестов, нет unit tests для domain |
| Observability | 9/10 | Отличная интеграция OTel |

**Общая оценка**: 5.7/10 - Good start, needs refactoring for production
