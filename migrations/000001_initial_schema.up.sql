CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─── Parents ──────────────────────────────────────────────────────────────
CREATE TABLE parents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(30),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Students (children of parents) ──────────────────────────────────────
CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    age INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Trial Classes ───────────────────────────────────────────────────────
CREATE TABLE trial_classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    instructor VARCHAR(255),
    scheduled_at TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 60,
    max_capacity INT NOT NULL DEFAULT 4,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Bookings ────────────────────────────────────────────────────────────
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    trial_class_id UUID NOT NULL REFERENCES trial_classes(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Payment Attempts ────────────────────────────────────────────────────
CREATE TABLE payment_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    amount_cents INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Constraints ─────────────────────────────────────────────────────────

-- Prevent duplicate active bookings (pending or confirmed) for same student+class
CREATE UNIQUE INDEX idx_unique_active_booking
    ON bookings(student_id, trial_class_id)
    WHERE status IN ('pending', 'confirmed');

-- ─── Performance Indexes ─────────────────────────────────────────────────
CREATE INDEX idx_students_parent ON students(parent_id);
CREATE INDEX idx_bookings_student ON bookings(student_id);
CREATE INDEX idx_bookings_class ON bookings(trial_class_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_payment_booking ON payment_attempts(booking_id);

-- ─── Seed Data ───────────────────────────────────────────────────────────

INSERT INTO parents (id, name, email, phone) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Sarah Johnson', 'sarah@example.com', '08123456789'),
    ('a0000000-0000-0000-0000-000000000002', 'Michael Chen', 'michael@example.com', '08234567890'),
    ('a0000000-0000-0000-0000-000000000003', 'Amanda Miller', 'amanda@example.com', '08345678901'),
    ('a0000000-0000-0000-0000-000000000004', 'David Tan', 'david@example.com', '08456789012');

INSERT INTO students (id, parent_id, name, age) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'Emma Johnson', 7),
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 'Liam Johnson', 5),
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', 'Olivia Chen', 8),
    ('b0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000002', 'Noah Chen', 6),
    ('b0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000003', 'Lucas Miller', 7),
    ('b0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000003', 'Sophia Miller', 9),
    ('b0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000004', 'Ethan Tan', 6),
    ('b0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000004', 'Chloe Tan', 8);

INSERT INTO trial_classes (id, title, description, instructor, scheduled_at, duration_minutes, max_capacity) VALUES
    ('c0000000-0000-0000-0000-000000000001', 'Introduction to Robotics', 'Learn the basics of robotics with hands-on projects', 'Mr. Smith', NOW() + INTERVAL '3 days', 60, 4),
    ('c0000000-0000-0000-0000-000000000002', 'Creative Coding for Kids', 'Fun introduction to coding through visual programming', 'Ms. Lee', NOW() + INTERVAL '5 days', 90, 4),
    ('c0000000-0000-0000-0000-000000000003', 'Art & Design Workshop', 'Express creativity through digital and traditional art', 'Mrs. Davis', NOW() + INTERVAL '7 days', 45, 4);

-- Pre-seed 3 confirmed bookings in Class 1 ("Introduction to Robotics") so that only 1 slot is left (max_capacity is 4)
-- This allows demonstrating last-seat race conditions immediately out of the box
INSERT INTO bookings (id, student_id, trial_class_id, status, created_at, updated_at) VALUES
    ('d0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000001', 'confirmed', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours'),
    ('d0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000006', 'c0000000-0000-0000-0000-000000000001', 'confirmed', NOW() - INTERVAL '90 minutes', NOW() - INTERVAL '90 minutes'),
    ('d0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000007', 'c0000000-0000-0000-0000-000000000001', 'confirmed', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour');

INSERT INTO payment_attempts (id, booking_id, amount_cents, status, created_at) VALUES
    ('e0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000001', 5000, 'success', NOW() - INTERVAL '2 hours'),
    ('e0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000002', 5000, 'success', NOW() - INTERVAL '90 minutes'),
    ('e0000000-0000-0000-0000-000000000003', 'd0000000-0000-0000-0000-000000000003', 5000, 'success', NOW() - INTERVAL '1 hour');

