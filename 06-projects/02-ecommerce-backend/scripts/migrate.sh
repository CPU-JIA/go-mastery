#!/bin/bash

# E-commerce Backend Database Migration Script
# Handles database schema migrations and data seeding

set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-ecommerce_system}"
DB_USER="${DB_USER:-ecommerce_user}"
DB_PASSWORD="${DB_PASSWORD:-ecommerce_password}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if PostgreSQL client is available
check_psql() {
    if ! command -v psql &> /dev/null; then
        print_error "PostgreSQL client (psql) is not installed"
        exit 1
    fi
}

# Test database connection
test_connection() {
    print_status "Testing database connection..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" &> /dev/null

    if [ $? -eq 0 ]; then
        print_status "Database connection successful"
    else
        print_error "Cannot connect to database"
        exit 1
    fi
}

# Create database if it doesn't exist
create_database() {
    print_status "Creating database if not exists..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || true
    print_status "Database $DB_NAME is ready"
}

# Run database migrations
migrate_up() {
    print_status "Running database migrations..."

    # Users table
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    role VARCHAR(20) DEFAULT 'customer',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES categories(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    cost_price DECIMAL(10,2),
    category_id INTEGER REFERENCES categories(id),
    sku VARCHAR(100) UNIQUE,
    stock_quantity INTEGER DEFAULT 0,
    min_stock_level INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    weight DECIMAL(8,3),
    dimensions JSONB,
    images JSONB,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    shipping_address JSONB,
    billing_address JSONB,
    payment_status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Shopping cart table
CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

-- Product reviews table
CREATE TABLE IF NOT EXISTS product_reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    comment TEXT,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, user_id)
);

-- Coupons table
CREATE TABLE IF NOT EXISTS coupons (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'percentage' or 'fixed'
    value DECIMAL(10,2) NOT NULL,
    minimum_amount DECIMAL(10,2),
    usage_limit INTEGER,
    used_count INTEGER DEFAULT 0,
    starts_at TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_user ON cart_items(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_product ON product_reviews(product_id);
EOF

    print_status "Database migrations completed successfully"
}

# Seed initial data
seed_data() {
    print_status "Seeding initial data..."

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- Insert default admin user
INSERT INTO users (username, email, password_hash, first_name, last_name, role) VALUES
('admin', 'admin@example.com', '\$2a\$10\$rUvDXJpJqJKaQWnNHLlV3.KQTtAc4qKFOF8nUVr8QF.Pz8QXGy8V6', 'Admin', 'User', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Insert sample customer
INSERT INTO users (username, email, password_hash, first_name, last_name, role) VALUES
('customer', 'customer@example.com', '\$2a\$10\$rUvDXJpJqJKaQWnNHLlV3.KQTtAc4qKFOF8nUVr8QF.Pz8QXGy8V6', 'Test', 'Customer', 'customer')
ON CONFLICT (username) DO NOTHING;

-- Insert sample categories
INSERT INTO categories (name, description) VALUES
('Electronics', 'Electronic devices and accessories'),
('Clothing', 'Apparel and fashion items'),
('Home & Garden', 'Home improvement and garden supplies'),
('Books', 'Books and educational materials'),
('Sports', 'Sports equipment and outdoor gear')
ON CONFLICT DO NOTHING;

-- Insert sample products
INSERT INTO products (name, slug, description, price, category_id, sku, stock_quantity) VALUES
('Wireless Headphones', 'wireless-headphones', 'High-quality wireless headphones with noise cancellation', 99.99, 1, 'WH001', 50),
('Cotton T-Shirt', 'cotton-t-shirt', 'Comfortable 100% cotton t-shirt', 19.99, 2, 'TS001', 100),
('Garden Hose', 'garden-hose', '50ft expandable garden hose', 29.99, 3, 'GH001', 25),
('Programming Book', 'go-programming-book', 'Learn Go Programming Language', 39.99, 4, 'BK001', 30),
('Tennis Racket', 'tennis-racket', 'Professional tennis racket', 89.99, 5, 'TR001', 15)
ON CONFLICT (slug) DO NOTHING;

-- Insert sample coupon
INSERT INTO coupons (code, type, value, minimum_amount, usage_limit, expires_at) VALUES
('WELCOME10', 'percentage', 10.00, 50.00, 100, CURRENT_TIMESTAMP + INTERVAL '30 days')
ON CONFLICT (code) DO NOTHING;
EOF

    print_status "Data seeding completed successfully"
}

# Drop all tables (use with caution!)
migrate_down() {
    print_warning "This will DROP ALL TABLES. Are you sure? (yes/no)"
    read -r confirmation
    if [ "$confirmation" = "yes" ]; then
        print_status "Dropping all tables..."

        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS product_reviews CASCADE;
DROP TABLE IF EXISTS coupons CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;
EOF

        print_status "All tables dropped"
    else
        print_status "Operation cancelled"
    fi
}

# Reset database (drop and recreate)
reset_database() {
    migrate_down
    migrate_up
    seed_data
}

# Show database status
status() {
    print_status "Database Status:"
    echo "Host: $DB_HOST:$DB_PORT"
    echo "Database: $DB_NAME"
    echo "User: $DB_USER"

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT
        schemaname,
        tablename,
        tableowner
    FROM pg_tables
    WHERE schemaname = 'public'
    ORDER BY tablename;
    "
}

# Show usage
usage() {
    echo "Usage: $0 {up|down|reset|seed|status|test}"
    echo ""
    echo "Commands:"
    echo "  up      - Run database migrations"
    echo "  down    - Drop all tables (destructive!)"
    echo "  reset   - Drop and recreate all tables"
    echo "  seed    - Seed initial data"
    echo "  status  - Show database status"
    echo "  test    - Test database connection"
}

# Main script logic
case "${1:-}" in
    up)
        check_psql
        test_connection
        create_database
        migrate_up
        ;;
    down)
        check_psql
        test_connection
        migrate_down
        ;;
    reset)
        check_psql
        test_connection
        create_database
        reset_database
        ;;
    seed)
        check_psql
        test_connection
        seed_data
        ;;
    status)
        check_psql
        test_connection
        status
        ;;
    test)
        check_psql
        test_connection
        ;;
    *)
        usage
        exit 1
        ;;
esac

print_status "Migration script completed successfully"