# RFC: Jadiles MVP - Educational Activity Booking Platform

## Document Information
- **Status**: Draft
- **Created**: 2025-10-22
- **Author**: System Architect
- **Version**: 1.0

---

## 1. Executive Summary

**Jadiles** is a marketplace platform connecting parents with educational activity providers (swimming schools, tutoring centers, art studios, coaches) in Tangerang, Indonesia.

### MVP Goal
Build a functional booking platform where:
- Parents can search and book courses for their children
- Vendors can register and manage their courses
- Secure payment processing via Midtrans
- Basic review system for trust building

### Out of Scope (MVP)
- Notification system (will use n8n integration)
- Advanced analytics/dashboards
- Mobile apps (web-responsive only)
- Multi-language support
- Real-time chat

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────┐
│   Web Client    │ (React/Next.js)
│   (Responsive)  │
└────────┬────────┘
         │ HTTPS/REST
         │
┌────────▼────────────────────────────────────┐
│          API Gateway / Load Balancer        │
└────────┬────────────────────────────────────┘
         │
┌────────▼────────────────────────────────────┐
│        Go Backend (Modular Monolith)        │
│  ┌──────────────────────────────────────┐   │
│  │  HTTP Layer (Chi/Fiber/Echo)         │   │
│  └──────────────┬───────────────────────┘   │
│                 │                            │
│  ┌──────────────▼───────────────────────┐   │
│  │  Domain Services Layer               │   │
│  │  - User Service                      │   │
│  │  - Vendor Service                    │   │
│  │  - Booking Service                   │   │
│  │  - Payment Service                   │   │
│  │  - Review Service                    │   │
│  └──────────────┬───────────────────────┘   │
│                 │                            │
│  ┌──────────────▼───────────────────────┐   │
│  │  Repository Layer (PostgreSQL)       │   │
│  └──────────────────────────────────────┘   │
└────────┬────────────────────┬────────────────┘
         │                    │
         │                    │
┌────────▼─────────┐   ┌──────▼──────────┐
│   PostgreSQL     │   │   Midtrans      │
│   (Primary DB)   │   │   Payment GW    │
└──────────────────┘   └─────────────────┘

External Integration Points:
┌──────────────────┐
│      n8n         │ ← Webhook events for notifications
│  (Notifications) │
└──────────────────┘
```

### 2.2 Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Backend** | Go 1.23+ | Performance, type safety, excellent stdlib |
| **Web Framework** | Chi/Fiber | Lightweight, middleware support |
| **Database** | PostgreSQL 15+ | ACID compliance, JSONB, spatial support |
| **Migration** | goose | Already in use, simple, bidirectional |
| **API Spec** | OpenAPI 3.0 | Contract-first development |
| **Auth** | JWT + Refresh Tokens | Stateless, scalable |
| **Payment** | Midtrans | Local payment methods (QRIS, e-wallet) |
| **File Storage** | S3-compatible (AWS/MinIO) | Scalable, CDN-ready |
| **Deployment** | Docker + Docker Compose | Consistent environments |

**Scale-Ready Stack (Post-MVP):**
- **Cache**: Redis (session, rate limiting, hot data)
- **Search**: Elasticsearch/Typesense (full-text, geo search)
- **Queue**: RabbitMQ/NATS (async processing)
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack or Loki
- **CDN**: CloudFlare or AWS CloudFront

---

## 3. Domain Model & Architecture

### 3.1 Modular Monolith Structure

Following your CLAUDE.md guidelines with clean, focused files:

```
internal/
├── core/
│   └── datamodel/
│       ├── user.go              # User database schema model
│       ├── parent_profile.go    # ParentProfile database model
│       ├── child.go             # Child database model
│       ├── vendor.go            # Vendor database model
│       ├── coach.go             # Coach database model
│       ├── service.go           # Service database model
│       ├── schedule.go          # Schedule database model
│       ├── booking.go           # Booking database model
│       ├── booking_session.go   # BookingSession database model
│       ├── payment.go           # Payment database model
│       ├── review.go            # Review database model
│       └── service_category.go  # ServiceCategory database model
│
├── auth/
│   ├── handler.go               # HTTP handlers (register, login, refresh)
│   ├── auth.go                  # Domain logic (JWT, password hashing, validation)
│   ├── service.go               # Business logic (orchestration, DTO conversion)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction (connection, ORM setup)
│       └── auth.go              # Repository implementation
│
├── user/
│   ├── handler.go               # HTTP handlers (get/update profile, manage children)
│   ├── user.go                  # Domain logic (struct, validation rules)
│   ├── service.go               # Business logic (profile updates, child management)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       ├── user.go              # User repository
│       └── child.go             # Child repository
│
├── vendor/
│   ├── handler.go               # HTTP handlers (vendor CRUD, service CRUD, schedules)
│   ├── vendor.go                # Domain logic (vendor validation, business rules)
│   ├── service.go               # Business logic (vendor management, service management)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       ├── vendor.go            # Vendor repository
│       ├── service.go           # Service repository
│       ├── coach.go             # Coach repository
│       └── schedule.go          # Schedule repository
│
├── booking/
│   ├── handler.go               # HTTP handlers (create/list/cancel bookings)
│   ├── booking.go               # Domain logic (booking validation, slot checking)
│   ├── service.go               # Business logic (booking orchestration, session generation)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       ├── booking.go           # Booking repository
│       └── session.go           # Session repository
│
├── payment/
│   ├── handler.go               # HTTP handlers (create payment, handle webhooks)
│   ├── payment.go               # Domain logic (payment validation, Midtrans integration)
│   ├── service.go               # Business logic (payment orchestration)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       └── payment.go           # Payment repository
│
├── review/
│   ├── handler.go               # HTTP handlers (submit review, list reviews)
│   ├── review.go                # Domain logic (review validation, rating calculation)
│   ├── service.go               # Business logic (review submission, vendor rating update)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       └── review.go            # Review repository
│
├── search/
│   ├── handler.go               # HTTP handlers (search services/vendors)
│   ├── search.go                # Domain logic (filter validation, query building)
│   ├── service.go               # Business logic (search orchestration)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       └── search.go            # Search repository (complex queries)
│
├── admin/
│   ├── handler.go               # HTTP handlers (vendor approval, stats)
│   ├── admin.go                 # Domain logic (approval validation, audit logging)
│   ├── service.go               # Business logic (vendor verification workflow)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       ├── postgresql.go        # DB abstraction
│       └── admin.go             # Admin repository
│
├── webhook/
│   ├── handler.go               # HTTP handlers (receive n8n callbacks)
│   ├── webhook.go               # Domain logic (event structure, validation)
│   ├── service.go               # Business logic (event dispatcher to n8n)
│   ├── endpoint/
│   │   └── endpoint.go          # Route registration & dependency injection
│   └── postgresql/
│       └── postgresql.go        # DB abstraction (if needed for event logging)
│
└── transport/
    ├── http/
    │   ├── server.go            # HTTP server setup
    │   ├── router.go            # Main router & middleware setup
    │   └── middleware/
    │       ├── auth.go          # JWT validation middleware
    │       ├── cors.go          # CORS middleware
    │       ├── logger.go        # Request logging middleware
    │       └── recovery.go      # Panic recovery middleware
    └── response/
        └── response.go          # Standardized API response helpers
```

### 3.2 File Responsibilities

#### **handler.go**
- Parse HTTP request
- Call service layer
- Return HTTP response
- Handle HTTP-specific errors (400, 401, 404, etc.)

**Example**:
```go
// internal/user/handler.go
type Handler struct {
    service *Service
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())

    user, err := h.service.GetProfile(r.Context(), userID)
    if err != nil {
        response.Error(w, err)
        return
    }

    response.JSON(w, http.StatusOK, user)
}
```

#### **{domain}.go** (e.g., user.go, vendor.go)
- Domain structs (request/response DTOs)
- Domain validation rules
- Business constants/enums
- Domain-specific errors

**Example**:
```go
// internal/user/user.go
type UpdateProfileRequest struct {
    FullName string `json:"full_name" validate:"required,min=3"`
    Phone    string `json:"phone" validate:"required,e164"`
    Address  string `json:"address"`
}

type UserResponse struct {
    ID        int64  `json:"id"`
    Email     string `json:"email"`
    FullName  string `json:"full_name"`
    Role      string `json:"role"`
    CreatedAt string `json:"created_at"`
}

func (r *UpdateProfileRequest) Validate() error {
    return validator.Validate(r)
}
```

#### **service.go**
- Business logic orchestration
- DTO ↔ Datamodel conversion
- Transaction management
- Calls to multiple repositories
- Integration with external services (n8n webhooks)

**Example**:
```go
// internal/user/service.go
type Service struct {
    userRepo  *postgresql.UserRepository
    childRepo *postgresql.ChildRepository
    webhook   *webhook.Service
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*UserResponse, error) {
    // 1. Get existing user
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 2. Update fields
    user.FullName = req.FullName
    user.Phone = req.Phone

    // 3. Save to DB
    if err := s.userRepo.Update(ctx, user); err != nil {
        return nil, err
    }

    // 4. Convert to response DTO
    return toUserResponse(user), nil
}
```

#### **endpoint/endpoint.go**
- Route registration
- Dependency injection
- Middleware attachment

**Example**:
```go
// internal/user/endpoint/endpoint.go
func RegisterRoutes(r chi.Router, db *sql.DB) {
    // Initialize repositories
    userRepo := postgresql.NewUserRepository(db)
    childRepo := postgresql.NewChildRepository(db)

    // Initialize service
    service := user.NewService(userRepo, childRepo)

    // Initialize handler
    handler := user.NewHandler(service)

    // Register routes
    r.Route("/users", func(r chi.Router) {
        r.Use(middleware.Auth) // Apply auth middleware

        r.Get("/me", handler.GetProfile)
        r.Put("/me", handler.UpdateProfile)
        r.Post("/me/children", handler.AddChild)
        r.Get("/me/children", handler.ListChildren)
    })
}
```

#### **postgresql/postgresql.go**
- Database connection/pool management
- ORM initialization (if using GORM/sqlx)
- Common DB utilities
- Transaction helpers

**Example**:
```go
// internal/user/postgresql/postgresql.go
type DB struct {
    *sql.DB
}

func NewDB(connStr string) (*DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    return &DB{db}, nil
}

func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}
```

#### **postgresql/{entity}.go** (e.g., user.go)
- Repository implementation
- SQL queries
- Datamodel ↔ DB row mapping
- CRUD operations

**Example**:
```go
// internal/user/postgresql/user.go
type UserRepository struct {
    db *DB
}

func NewUserRepository(db *DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*datamodel.User, error) {
    var user datamodel.User

    err := r.db.QueryRowContext(ctx, `
        SELECT id, email, full_name, phone, role, created_at, updated_at
        FROM users WHERE id = $1
    `, id).Scan(&user.ID, &user.Email, &user.FullName, &user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt)

    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }

    return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *datamodel.User) error {
    _, err := r.db.ExecContext(ctx, `
        UPDATE users
        SET full_name = $1, phone = $2, updated_at = NOW()
        WHERE id = $3
    `, user.FullName, user.Phone, user.ID)

    return err
}
```

#### **core/datamodel/{entity}.go**
- Database schema representation
- Pure data structures (no business logic)
- Shared across all domains

**Example**:
```go
// internal/core/datamodel/user.go
package datamodel

import "time"

type User struct {
    ID            int64     `db:"id"`
    Email         string    `db:"email"`
    PasswordHash  string    `db:"password_hash"`
    FullName      string    `db:"full_name"`
    Phone         string    `db:"phone"`
    Role          string    `db:"role"`
    Status        string    `db:"status"`
    EmailVerified bool      `db:"email_verified"`
    PhoneVerified bool      `db:"phone_verified"`
    CreatedAt     time.Time `db:"created_at"`
    UpdatedAt     time.Time `db:"updated_at"`
}
```

### 3.2 Key Architectural Decisions

#### **Decision 1: Modular Monolith vs Microservices**
- **Choice**: Modular Monolith
- **Rationale**:
  - Faster development for MVP
  - Lower operational complexity
  - Single deployment unit
  - Easy to extract services later
- **Future**: Can split into microservices at domain boundaries

#### **Decision 2: Repository Pattern**
- **Choice**: Repository per aggregate root
- **Rationale**:
  - Clean separation of data access
  - Testable (mock repositories)
  - Database-agnostic interface

#### **Decision 3: Service Layer**
- **Choice**: Service orchestrates multiple repositories and domain logic
- **Rationale**:
  - DTO ↔ Domain conversion happens here
  - Transaction boundaries
  - Business rule enforcement

#### **Decision 4: n8n Integration**
- **Choice**: Webhook-based event dispatch
- **Implementation**:
  ```go
  // When booking is created:
  webhookService.Dispatch("booking.created", BookingEventDTO{
      BookingID: booking.ID,
      ParentEmail: parent.Email,
      VendorID: vendor.ID,
      // ... other fields
  })
  ```
- **Scale Path**: Replace with message queue (RabbitMQ/NATS)

---

## 4. MVP Feature Specification

### 4.1 User Management

#### **UC-001: User Registration**
**Actor**: Parent, Vendor, Coach

**Flow**:
1. User submits registration form (email, password, full_name, phone, role)
2. System validates email uniqueness
3. System hashes password (bcrypt cost=12)
4. System creates user record
5. System returns JWT access token + refresh token
6. **Webhook**: Dispatch `user.registered` event to n8n (for welcome email)

**API**:
```
POST /api/v1/auth/register
{
  "email": "parent@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe",
  "phone": "+628123456789",
  "role": "parent"
}

Response 201:
{
  "user": {
    "id": 1,
    "email": "parent@example.com",
    "full_name": "John Doe",
    "role": "parent"
  },
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 3600
}
```

#### **UC-002: User Login**
**API**:
```
POST /api/v1/auth/login
{
  "email": "parent@example.com",
  "password": "SecurePass123!"
}
```

#### **UC-003: Parent Profile Management**
**APIs**:
- `GET /api/v1/users/me` - Get current user
- `PUT /api/v1/users/me` - Update profile
- `POST /api/v1/users/me/children` - Add child
- `GET /api/v1/users/me/children` - List children
- `PUT /api/v1/users/me/children/:id` - Update child

---

### 4.2 Vendor Management

#### **UC-004: Vendor Registration**
**Actor**: Vendor (user with role='vendor')

**Flow**:
1. User registers with role='vendor'
2. User completes vendor profile
3. System creates vendor record (status='pending')
4. **Webhook**: Dispatch `vendor.registered` to n8n (notify admin)
5. Admin approves/rejects

**API**:
```
POST /api/v1/vendors
{
  "business_name": "ABC Swimming School",
  "business_type": "swimming_school",
  "description": "Professional swimming lessons",
  "phone": "+628123456789",
  "address": "Jl. Example No. 123",
  "city": "Tangerang",
  "latitude": -6.1783,
  "longitude": 106.6319
}

Response 201:
{
  "id": 1,
  "user_id": 5,
  "business_name": "ABC Swimming School",
  "status": "pending",
  "created_at": "2025-10-22T10:00:00Z"
}
```

#### **UC-005: Service/Course Registration**
**Actor**: Vendor (status='active')

**API**:
```
POST /api/v1/vendors/me/services
{
  "category_id": 1,
  "name": "Kids Swimming Beginner",
  "description": "Learn basic swimming skills",
  "class_type": "small_group",
  "age_min": 5,
  "age_max": 12,
  "skill_level": "beginner",
  "max_participants": 8,
  "duration_minutes": 60,
  "price_per_session": 150000,
  "trial_price": 100000,
  "package_4_price": 550000,
  "package_8_price": 1000000,
  "package_12_price": 1400000
}
```

#### **UC-006: Schedule Management**
**API**:
```
POST /api/v1/vendors/me/services/:service_id/schedules
{
  "day_of_week": 1,  // Monday (0=Sunday)
  "start_time": "09:00",
  "end_time": "10:00",
  "available_slots": 8
}

POST /api/v1/vendors/me/schedule-exceptions
{
  "exception_date": "2025-12-25",
  "reason": "Christmas Holiday",
  "is_closed": true
}
```

---

### 4.3 Search & Discovery

#### **UC-007: Search Vendors/Services**
**Actor**: Parent (public or authenticated)

**API**:
```
GET /api/v1/search/services?
  category=swimming&
  city=Tangerang&
  age=7&
  class_type=small_group&
  min_rating=4.0&
  sort=rating_desc&
  page=1&
  limit=20

Response 200:
{
  "data": [
    {
      "id": 1,
      "vendor": {
        "id": 1,
        "business_name": "ABC Swimming",
        "rating_avg": 4.5,
        "total_reviews": 120,
        "verified": true
      },
      "name": "Kids Swimming Beginner",
      "class_type": "small_group",
      "price_per_session": 150000,
      "duration_minutes": 60,
      "next_available": "2025-10-25T09:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

**Filters**:
- `category` - service category slug
- `city`, `district` - location
- `lat`, `lng`, `radius` - geo search (future: use PostGIS)
- `age` - child age (filters by age_min/age_max)
- `class_type` - private/small_group/large_group
- `min_rating` - minimum vendor rating
- `business_type` - swimming_school, tutoring_center, etc.
- `sort` - rating_desc, price_asc, distance_asc

#### **UC-008: View Service Detail**
**API**:
```
GET /api/v1/services/:id

Response 200:
{
  "id": 1,
  "vendor": { /* full vendor details */ },
  "category": { /* category details */ },
  "coaches": [ /* assigned coaches */ ],
  "schedules": [ /* weekly schedules */ ],
  "reviews": [ /* recent reviews */ ],
  "similar_services": [ /* recommendations */ ]
}
```

---

### 4.4 Booking System

#### **UC-009: Create Booking**
**Actor**: Parent (authenticated)

**Flow**:
1. Parent selects service + child + booking type
2. System checks schedule availability
3. System calculates total amount
4. System creates booking (status='pending', payment_status='unpaid')
5. System generates booking sessions
6. System reserves slots (decrement `available_slots`)
7. **Webhook**: Dispatch `booking.created` to n8n
8. Return booking + payment instructions

**API**:
```
POST /api/v1/bookings
{
  "child_id": 3,
  "service_id": 1,
  "booking_type": "package_8",
  "preferred_schedule_id": 5,
  "start_date": "2025-10-28",
  "parent_notes": "Child has swimming experience"
}

Response 201:
{
  "id": 1,
  "booking_number": "BK20251022001",
  "service": { /* service details */ },
  "child": { /* child details */ },
  "booking_type": "package_8",
  "total_sessions": 8,
  "total_amount": 1000000,
  "status": "pending",
  "payment_status": "unpaid",
  "sessions": [
    {
      "session_number": 1,
      "session_date": "2025-10-28",
      "start_time": "09:00",
      "end_time": "10:00",
      "status": "scheduled"
    }
    // ... 7 more sessions
  ],
  "payment": {
    "payment_url": "https://app.midtrans.com/snap/v2/...",
    "expires_at": "2025-10-22T12:00:00Z"
  }
}
```

**Business Rules**:
- Can only book if service is active
- Can only book if vendor is active/verified
- Must have available slots for all sessions
- Booking expires if unpaid after 24 hours

#### **UC-010: View Bookings**
**APIs**:
- `GET /api/v1/bookings` - List my bookings (parent/vendor)
- `GET /api/v1/bookings/:id` - Booking detail

**Filters** (for parents):
- `status` - pending, confirmed, ongoing, completed, cancelled
- `payment_status` - unpaid, paid, refunded

---

### 4.5 Payment System

#### **UC-011: Create Payment**
**Actor**: Parent

**Flow**:
1. System creates payment record (status='pending')
2. System calls Midtrans Snap API
3. Midtrans returns payment URL
4. Parent redirects to Midtrans payment page
5. Midtrans sends webhook on payment success/failure
6. System updates payment + booking status
7. **Webhook**: Dispatch `payment.success` to n8n (send receipt)

**API**:
```
POST /api/v1/payments
{
  "booking_id": 1,
  "payment_method": "qris"  // or credit_card, bank_transfer, e_wallet
}

Response 201:
{
  "id": 1,
  "payment_number": "PAY20251022001",
  "amount": 1000000,
  "payment_method": "qris",
  "status": "pending",
  "payment_url": "https://app.midtrans.com/snap/v2/...",
  "expired_at": "2025-10-22T12:00:00Z"
}
```

#### **UC-012: Midtrans Webhook**
**Actor**: Midtrans

**API**:
```
POST /api/v1/webhooks/midtrans
{
  "transaction_status": "settlement",
  "order_id": "PAY20251022001",
  "gross_amount": "1000000.00",
  "transaction_id": "abc123xyz",
  "signature_key": "..."
}
```

**Webhook Handler**:
1. Validate signature
2. Update payment status
3. Update booking status (pending → confirmed)
4. Update booking payment_status (unpaid → paid)
5. Dispatch `payment.success` event to n8n

---

### 4.6 Review System

#### **UC-013: Submit Review**
**Actor**: Parent (only for completed bookings)

**Business Rules**:
- Can only review once per booking
- Booking must be completed
- At least 1 session must be attended

**API**:
```
POST /api/v1/reviews
{
  "booking_id": 1,
  "rating": 5,
  "review_text": "Excellent instructor, my child loves it!",
  "child_enjoyed": true,
  "would_recommend": true
}

Response 201:
{
  "id": 1,
  "booking_id": 1,
  "vendor": { /* vendor details */ },
  "service": { /* service details */ },
  "rating": 5,
  "review_text": "...",
  "created_at": "2025-10-22T10:00:00Z"
}
```

**Post-Processing**:
- Recalculate vendor `rating_avg` and `total_reviews`
- Dispatch `review.created` event to n8n (notify vendor)

#### **UC-014: View Reviews**
**APIs**:
- `GET /api/v1/vendors/:id/reviews` - Vendor reviews
- `GET /api/v1/services/:id/reviews` - Service reviews

---

### 4.7 Admin Functions

#### **UC-015: Vendor Approval**
**Actor**: Admin

**API**:
```
PUT /api/v1/admin/vendors/:id/approve
{
  "status": "active",  // or "rejected"
  "rejection_reason": "Missing business license"  // if rejected
}
```

**Actions**:
- Update vendor status
- Log admin action to `admin_actions`
- Dispatch `vendor.approved` or `vendor.rejected` to n8n

#### **UC-016: Admin Dashboard**
**APIs**:
- `GET /api/v1/admin/stats` - System statistics
- `GET /api/v1/admin/vendors?status=pending` - Pending vendors
- `GET /api/v1/admin/bookings?status=disputed` - Disputed bookings

---

## 5. n8n Integration Strategy

### 5.1 Webhook Event Schema

**Standard Event Structure**:
```json
{
  "event_type": "booking.created",
  "event_id": "evt_abc123",
  "timestamp": "2025-10-22T10:00:00Z",
  "data": {
    // Event-specific payload
  }
}
```

### 5.2 Event Types (MVP)

| Event | Trigger | n8n Action |
|-------|---------|------------|
| `user.registered` | User signs up | Send welcome email |
| `vendor.registered` | Vendor completes profile | Notify admin for approval |
| `vendor.approved` | Admin approves vendor | Send approval email to vendor |
| `vendor.rejected` | Admin rejects vendor | Send rejection email with reason |
| `booking.created` | Parent creates booking | Send booking confirmation to parent + vendor |
| `payment.pending` | Payment created | Send payment reminder (if not paid in 12h) |
| `payment.success` | Payment completed | Send receipt to parent, notify vendor |
| `payment.failed` | Payment failed | Send retry instructions |
| `booking.cancelled` | Booking cancelled | Send cancellation notice to both parties |
| `session.reminder` | 24h before session | Send reminder to parent |
| `booking.completed` | All sessions done | Request review from parent |
| `review.created` | Parent submits review | Notify vendor of new review |

### 5.3 Implementation

**Go Webhook Dispatcher**:
```go
// internal/webhook/service/dispatcher.go
type Dispatcher struct {
    webhookURL string
    httpClient *http.Client
}

func (d *Dispatcher) Dispatch(eventType string, data interface{}) error {
    event := WebhookEvent{
        EventType: eventType,
        EventID:   generateEventID(),
        Timestamp: time.Now(),
        Data:      data,
    }

    payload, _ := json.Marshal(event)

    // Async dispatch (use goroutine + error logging)
    go func() {
        resp, err := d.httpClient.Post(d.webhookURL, "application/json", bytes.NewBuffer(payload))
        if err != nil {
            log.Error("webhook dispatch failed", "event", eventType, "error", err)
            // TODO: Retry logic or dead letter queue
        }
        defer resp.Body.Close()
    }()

    return nil
}
```

**Scale Path**:
- **Phase 1 (MVP)**: Direct HTTP webhooks to n8n
- **Phase 2**: Add retry logic + idempotency keys
- **Phase 3**: Replace with message queue (RabbitMQ/NATS)
  ```
  Go App → Publish to Queue → n8n Consumer
  ```
- **Phase 4**: Event sourcing pattern

---

## 6. API Design Standards

### 6.1 RESTful Conventions

```
# Resources
GET    /api/v1/services              # List services
POST   /api/v1/services              # Create service (vendor only)
GET    /api/v1/services/:id          # Get service detail
PUT    /api/v1/services/:id          # Update service
DELETE /api/v1/services/:id          # Delete service

# Nested resources
GET    /api/v1/vendors/:id/services  # Vendor's services
POST   /api/v1/bookings/:id/cancel   # Action endpoint
```

### 6.2 Response Format

**Success Response**:
```json
{
  "data": { /* resource or array */ },
  "meta": {
    "pagination": { /* if applicable */ }
  }
}
```

**Error Response**:
```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Email already exists"
      }
    ]
  }
}
```

### 6.3 HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Success (GET, PUT, DELETE) |
| 201 | Created (POST) |
| 204 | No Content (DELETE) |
| 400 | Bad Request (validation errors) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate resource) |
| 422 | Unprocessable Entity (business logic error) |
| 500 | Internal Server Error |

---

## 7. Authentication & Authorization

### 7.1 JWT Strategy

**Access Token**:
- Lifetime: 1 hour
- Payload:
  ```json
  {
    "sub": "1",           // user ID
    "email": "user@example.com",
    "role": "parent",
    "exp": 1729684800
  }
  ```

**Refresh Token**:
- Lifetime: 30 days
- Stored in database for revocation
- Refresh endpoint: `POST /api/v1/auth/refresh`

### 7.2 Authorization Rules

| Endpoint | Parent | Vendor | Coach | Admin |
|----------|--------|--------|-------|-------|
| `POST /bookings` | ✅ | ❌ | ❌ | ✅ |
| `GET /bookings` | ✅ (own) | ✅ (own) | ✅ (own) | ✅ (all) |
| `POST /vendors/me/services` | ❌ | ✅ | ❌ | ❌ |
| `PUT /vendors/:id/approve` | ❌ | ❌ | ❌ | ✅ |
| `GET /search/services` | ✅ | ✅ | ✅ | ✅ |

**Implementation**: Middleware-based RBAC
```go
router.Post("/vendors/me/services",
    middleware.Auth,
    middleware.RequireRole("vendor"),
    handler.CreateService)
```

---

## 8. Data Validation

### 8.1 Validation Library
Use `go-playground/validator` v10

**Example**:
```go
type CreateBookingRequest struct {
    ChildID     int64  `json:"child_id" validate:"required"`
    ServiceID   int64  `json:"service_id" validate:"required"`
    BookingType string `json:"booking_type" validate:"required,oneof=trial single package_4 package_8 package_12"`
    StartDate   string `json:"start_date" validate:"required,datetime=2006-01-02"`
}
```

### 8.2 Business Validation

**In Service Layer**:
```go
func (s *BookingService) CreateBooking(req dto.CreateBookingRequest) error {
    // 1. Validate service is active
    service, err := s.serviceRepo.GetByID(req.ServiceID)
    if service.Status != "active" {
        return errors.New("service not active")
    }

    // 2. Validate child belongs to parent
    child, err := s.childRepo.GetByID(req.ChildID)
    if child.ParentID != currentUserID {
        return errors.New("unauthorized")
    }

    // 3. Validate age requirement
    if !service.IsAgeEligible(child.Age()) {
        return errors.New("child age not eligible")
    }

    // 4. Check slot availability
    // ... etc
}
```

---

## 9. Database Strategy

### 9.1 Connection Pooling

```go
// config
MaxOpenConns:    25,
MaxIdleConns:    5,
ConnMaxLifetime: 5 * time.Minute,
ConnMaxIdleTime: 10 * time.Minute,
```

### 9.2 Migration Strategy

**Development**:
```bash
goose -dir db/migrations postgres "connection_string" up
```

**Production**:
- Migrations run in CI/CD before deployment
- Use transaction-wrapped migrations where possible
- Keep migrations small and reversible

### 9.3 Indexes (Already in Schema)

Critical indexes for MVP queries:
- `idx_vendors_city` - Location search
- `idx_services_status` - Active service filtering
- `idx_bookings_parent_id` - Parent's bookings
- `idx_bookings_vendor_id` - Vendor's bookings
- `idx_payments_booking_id` - Payment lookup

**Future Optimization**:
- Composite indexes: `(city, status, rating_avg)` for filtered search
- Partial indexes: `WHERE status = 'active'`
- GIN index on JSONB columns if querying

---

## 10. File Upload Strategy

### 10.1 Supported File Types

| Entity | Field | Types | Max Size |
|--------|-------|-------|----------|
| User | profile_image | JPG, PNG | 2MB |
| Vendor | logo, cover_image | JPG, PNG | 5MB |
| Vendor | photos | JPG, PNG | 5MB each |
| Child | photo | JPG, PNG | 2MB |
| Review | photos | JPG, PNG | 3MB each |

### 10.2 Upload Flow

```
Client → POST /api/v1/upload → Backend → S3 → Return URL
                                   ↓
                            Generate presigned URL
                            or CloudFront URL
```

**Implementation**:
```go
POST /api/v1/upload
Content-Type: multipart/form-data

{
  "file": <binary>,
  "type": "vendor_logo"  // for validation rules
}

Response 200:
{
  "url": "https://cdn.jadiles.com/vendors/logos/abc123.jpg"
}
```

**Scale Path**:
- **Phase 1**: Direct upload to S3 via backend
- **Phase 2**: Presigned URLs (client uploads directly to S3)
- **Phase 3**: Image optimization (resize, WebP conversion)
- **Phase 4**: CDN integration (CloudFront/CloudFlare)

---

## 11. Performance Requirements

### 11.1 Response Time SLAs (95th percentile)

| Endpoint Type | Target | Max |
|---------------|--------|-----|
| Authentication | < 200ms | 500ms |
| Search/List | < 300ms | 1s |
| Detail Views | < 200ms | 500ms |
| Booking Creation | < 500ms | 2s |
| Payment Creation | < 1s | 3s |

### 11.2 Database Query Optimization

**N+1 Query Prevention**:
```go
// Bad: N+1 queries
services, _ := repo.GetServices()
for _, service := range services {
    vendor, _ := repo.GetVendor(service.VendorID)  // N queries
}

// Good: Eager loading
services, _ := repo.GetServicesWithVendors()  // Single join query
```

**Pagination**:
- Default limit: 20
- Max limit: 100
- Use cursor-based pagination for large datasets (future)

---

## 12. Error Handling

### 12.1 Error Types

```go
package errors

type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Details    map[string]string
}

var (
    ErrNotFound          = &AppError{"NOT_FOUND", "Resource not found", 404, nil}
    ErrUnauthorized      = &AppError{"UNAUTHORIZED", "Authentication required", 401, nil}
    ErrForbidden         = &AppError{"FORBIDDEN", "Insufficient permissions", 403, nil}
    ErrValidation        = &AppError{"VALIDATION_ERROR", "Invalid input", 400, nil}
    ErrBookingNotAllowed = &AppError{"BOOKING_NOT_ALLOWED", "Cannot book this service", 422, nil}
)
```

### 12.2 Error Response Examples

```json
// Validation error
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": {
      "email": "Email already exists",
      "phone": "Phone number format invalid"
    }
  }
}

// Business logic error
{
  "error": {
    "code": "BOOKING_NOT_ALLOWED",
    "message": "Child age does not meet service requirements",
    "details": {
      "child_age": "7",
      "required_age": "8-12"
    }
  }
}
```

---

## 13. Testing Strategy

### 13.1 Test Pyramid

```
         /\
        /  \    E2E Tests (10%)
       /────\   - Critical user flows
      /      \  - Payment flow
     /────────\ Integration Tests (30%)
    /          \ - API endpoint tests
   /────────────\ - Database integration
  /──────────────\ Unit Tests (60%)
 /                \ - Service layer logic
/──────────────────\ - Domain logic
```

### 13.2 Testing Tools

| Type | Tool | Usage |
|------|------|-------|
| Unit | `testing` (stdlib) | Service/domain logic |
| Mocking | `gomock` or `testify/mock` | Repository mocks |
| Integration | `testcontainers-go` | Real PostgreSQL in Docker |
| API | `httptest` | HTTP handler testing |
| E2E | Postman/Newman or Playwright | Critical flows |

### 13.3 Critical Test Cases (MVP)

**Must Test**:
1. ✅ User registration (duplicate email handling)
2. ✅ JWT token generation & validation
3. ✅ Booking creation (slot availability check)
4. ✅ Payment webhook (signature validation)
5. ✅ Review submission (authorization check)
6. ✅ Vendor search (filtering & pagination)
7. ✅ Booking cancellation (refund logic)

---

## 14. Security Requirements

### 14.1 Security Checklist

**Authentication**:
- ✅ Passwords hashed with bcrypt (cost >= 12)
- ✅ JWT tokens with short expiration (1h)
- ✅ Refresh token rotation
- ✅ Token blacklist on logout (store in Redis, future)

**Authorization**:
- ✅ Role-based access control (RBAC)
- ✅ Resource ownership validation (e.g., parent can only view own bookings)
- ✅ Admin actions audit logging

**Input Validation**:
- ✅ Request validation (go-playground/validator)
- ✅ SQL injection prevention (parameterized queries)
- ✅ XSS prevention (escape HTML in user content)
- ✅ File upload validation (type, size, magic bytes)

**API Security**:
- ✅ CORS configuration (whitelist frontend domain)
- ✅ Rate limiting (future: Redis-based)
- ✅ HTTPS only (TLS 1.3)
- ✅ Secure headers (Helmet-equivalent middleware)

**Payment Security**:
- ✅ Midtrans webhook signature validation
- ✅ No sensitive payment data stored (PCI compliance)
- ✅ Idempotency keys for payment creation

**Data Privacy**:
- ✅ Passwords never logged
- ✅ Sensitive fields masked in logs (email, phone)
- ✅ GDPR compliance (user data export/deletion, future)

---

## 15. Deployment Strategy

### 15.1 MVP Deployment (Docker Compose)

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - MIDTRANS_SERVER_KEY=${MIDTRANS_SERVER_KEY}
      - JWT_SECRET=${JWT_SECRET}
      - N8N_WEBHOOK_URL=${N8N_WEBHOOK_URL}
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=jadiles
      - POSTGRES_USER=jadiles
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  pgdata:
```

### 15.2 Environment Variables

```bash
# .env.example
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=jadiles
DB_USER=jadiles
DB_PASSWORD=changeme
DB_SSL_MODE=disable

# Server
SERVER_PORT=8080
SERVER_ENV=production  # development, staging, production

# JWT
JWT_SECRET=your-256-bit-secret
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=720h

# Midtrans
MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_CLIENT_KEY=your-client-key
MIDTRANS_ENVIRONMENT=sandbox  # sandbox or production

# Storage
S3_ENDPOINT=https://s3.amazonaws.com
S3_BUCKET=jadiles-uploads
S3_REGION=ap-southeast-1
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key

# Webhooks
N8N_WEBHOOK_URL=https://n8n.example.com/webhook/jadiles

# CORS
CORS_ALLOWED_ORIGINS=https://jadiles.com,https://www.jadiles.com
```

### 15.3 CI/CD Pipeline (GitHub Actions Example)

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run migrations
        run: |
          goose -dir db/migrations postgres "${{ secrets.DB_CONNECTION_STRING }}" up
      - name: Build Docker image
        run: docker build -t jadiles:${{ github.sha }} .
      - name: Deploy to server
        # SSH into server and restart containers
```

---

## 16. Monitoring & Observability (Future)

### 16.1 Metrics to Track

**Business Metrics**:
- New user registrations (daily/weekly)
- Bookings created (by status, payment_status)
- Revenue (total, per vendor, per category)
- Conversion rate (searches → bookings)
- Vendor approval time (avg)

**Technical Metrics**:
- API response times (p50, p95, p99)
- Error rates (4xx, 5xx)
- Database query times
- Active connections
- Request throughput (req/sec)

### 16.2 Logging Strategy

**Structured Logging** (use `zap` or `zerolog`):
```go
log.Info("booking created",
    "booking_id", booking.ID,
    "user_id", userID,
    "amount", booking.TotalAmount,
    "vendor_id", booking.VendorID)
```

**Log Levels**:
- `DEBUG` - Detailed debugging (development only)
- `INFO` - Important events (booking created, payment success)
- `WARN` - Recoverable errors (rate limit hit, retry successful)
- `ERROR` - Errors requiring attention (payment failed, webhook timeout)
- `FATAL` - Unrecoverable errors (DB connection lost)

---

## 17. Scalability Roadmap

### Phase 1: MVP (0-1K users)
- Single server deployment
- PostgreSQL on same VPS
- Direct n8n webhooks
- **Cost**: ~$20-50/month (DigitalOcean/Hetzner)

### Phase 2: Growth (1K-10K users)
- **Horizontal scaling**: Multiple app instances behind load balancer
- **Database**: Managed PostgreSQL (AWS RDS, DigitalOcean Managed DB)
- **Cache layer**: Redis for sessions, rate limiting, hot data
- **CDN**: CloudFlare for static assets
- **Cost**: ~$200-500/month

### Phase 3: Scale (10K-100K users)
- **Database**: Read replicas for search queries
- **Search**: Elasticsearch/Typesense for full-text + geo search
- **Queue**: RabbitMQ/NATS for async processing (replace direct webhooks)
- **Microservices**: Extract payment service, notification service
- **Cost**: ~$1K-3K/month

### Phase 4: Enterprise (100K+ users)
- **Kubernetes**: Container orchestration
- **Database sharding**: Partition by region/vendor
- **Event sourcing**: Full audit trail + event replay
- **Multi-region**: Deploy in multiple AWS regions
- **Cost**: $5K+/month

---

## 18. MVP Success Criteria

### 18.1 Functional Requirements
- ✅ 4 core features working (search, book, vendor registration, payment)
- ✅ All critical user flows tested
- ✅ Midtrans payment integration live
- ✅ n8n webhook integration working
- ✅ Mobile-responsive web UI

### 18.2 Non-Functional Requirements
- ✅ API response time < 500ms (95th percentile)
- ✅ 99% uptime
- ✅ SSL/HTTPS enabled
- ✅ Basic error monitoring (Sentry or similar)
- ✅ Automated database backups

### 18.3 Business Metrics (First 3 Months)
- 🎯 50+ vendors registered
- 🎯 500+ parent accounts
- 🎯 100+ bookings completed
- 🎯 10+ positive reviews
- 🎯 Payment success rate > 95%

---

## 19. Open Questions & Decisions Needed

### 19.1 Business Questions
1. **Refund policy**: Full refund if cancelled X hours before session?
2. **Commission model**: % of booking or monthly vendor subscription?
3. **Vendor onboarding**: Manual approval or auto-approve with verification?
4. **Free trial**: Offer free trial bookings for first-time parents?

### 19.2 Technical Questions
1. **File storage**: AWS S3 or DigitalOcean Spaces or MinIO (self-hosted)?
2. **Email provider** (for n8n): SendGrid, AWS SES, or Mailgun?
3. **Monitoring**: Sentry for errors? Grafana Cloud for metrics?
4. **Deployment**: VPS (DigitalOcean, Hetzner) or cloud (AWS, GCP)?

---

## 20. Timeline Estimate

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Setup & Infrastructure** | Week 1 | - Project scaffolding<br>- Database setup<br>- CI/CD pipeline<br>- Docker configuration |
| **Auth & User Management** | Week 2 | - Registration/Login<br>- JWT implementation<br>- User profile CRUD<br>- Child management |
| **Vendor Management** | Week 3 | - Vendor registration<br>- Service/course CRUD<br>- Schedule management<br>- Admin approval |
| **Search & Discovery** | Week 4 | - Service search API<br>- Filtering & pagination<br>- Service detail page<br>- Vendor listing |
| **Booking System** | Week 5-6 | - Booking creation<br>- Session generation<br>- Availability checking<br>- Booking management |
| **Payment Integration** | Week 7 | - Midtrans integration<br>- Webhook handler<br>- Payment status tracking |
| **Review System** | Week 8 | - Review submission<br>- Rating calculation<br>- Review moderation |
| **n8n Integration** | Week 9 | - Webhook dispatcher<br>- Event definitions<br>- n8n workflow setup |
| **Testing & Bug Fixes** | Week 10 | - Integration tests<br>- E2E testing<br>- Bug fixes<br>- Performance optimization |
| **Deployment & Launch** | Week 11-12 | - Production deployment<br>- Monitoring setup<br>- Documentation<br>- Soft launch |

**Total: 12 weeks (3 months)**

---

## 21. Appendix

### 21.1 Database Schema ER Diagram
```
users (1) ─────── (1) parent_profiles
  │                     │
  │                     └─── (N) children
  │
  ├─── (1) vendors ───────┬─── (N) services ───────┬─── (N) schedules
  │                       │                        │
  │                       └─── (N) coaches         └─── (N) service_coaches
  │
  └─── (N) bookings ──────┬─── (N) booking_sessions
                          │
                          ├─── (N) payments
                          │
                          └─── (1) reviews
```

### 21.2 Key Dependencies (go.mod)

```go
module github.com/jadiles/backend

go 1.23

require (
    github.com/go-chi/chi/v5 v5.0.10        // HTTP router
    github.com/lib/pq v1.10.9               // PostgreSQL driver
    github.com/golang-jwt/jwt/v5 v5.2.0     // JWT
    github.com/go-playground/validator/v10  // Validation
    golang.org/x/crypto v0.17.0             // bcrypt
    github.com/midtrans/midtrans-go v1.3.7  // Midtrans SDK
    github.com/aws/aws-sdk-go v1.49.0       // S3 upload
    go.uber.org/zap v1.26.0                 // Logging
)
```

### 21.3 Glossary

- **Booking Type**: Trial, single session, or package (4/8/12 sessions)
- **Class Type**: Private (1-on-1), small group (2-6), large group (7+)
- **Service**: A course/class offered by vendor (e.g., "Kids Swimming Beginner")
- **Session**: Individual class occurrence within a booking
- **Vendor**: Business providing educational services
- **Coach**: Instructor employed by vendor

---

## 22. Approval & Sign-off

This RFC requires approval from:
- [ ] Product Owner
- [ ] Tech Lead
- [ ] Backend Engineer
- [ ] DevOps Engineer

**Feedback Deadline**: [Date]
**Target Start Date**: [Date]

---

**Document Status**: Draft
**Last Updated**: 2025-10-22
**Next Review**: Upon implementation completion
