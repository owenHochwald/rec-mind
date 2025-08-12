#!/bin/bash

# Distributed ML Services Startup Script
# Sets up multiple ML service instances for distributed chunk processing

set -e

echo "🚀 Starting Distributed ML Services Setup"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_success "Docker is running"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed. Please install it first."
    exit 1
fi

print_success "docker-compose is available"

# Start infrastructure services
print_status "Starting infrastructure services (PostgreSQL, Redis, RabbitMQ)..."
cd infra
docker-compose up -d postgres redis rabbitmq

# Wait for services to be ready
print_status "Waiting for services to initialize..."
sleep 15

# Check if services are healthy
print_status "Checking service health..."

# Check PostgreSQL
if docker-compose exec -T postgres pg_isready -U ${DB_USER:-myuser} >/dev/null 2>&1; then
    print_success "PostgreSQL is ready"
else
    print_warning "PostgreSQL may not be fully ready yet"
fi

# Check Redis
if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
    print_success "Redis is ready"
else
    print_warning "Redis may not be fully ready yet"
fi

# Check RabbitMQ
print_status "RabbitMQ Management UI available at: http://localhost:15672 (myuser/secret)"

# Start ML service instances
print_status "Starting ML service instances..."
docker-compose up -d ml-service-1 ml-service-2 ml-service-3

print_status "Waiting for ML services to initialize..."
sleep 20

# Check ML service health
print_status "Checking ML service health..."
for i in {1..3}; do
    port=$((8000 + i))
    if curl -s http://localhost:${port}/health >/dev/null 2>&1; then
        print_success "ML Service ${i} is healthy (port ${port})"
    else
        print_warning "ML Service ${i} may not be ready yet (port ${port})"
    fi
done

print_success "All services started!"

echo ""
echo "📊 Service Status:"
echo "=================="
echo "🗄️  PostgreSQL:   localhost:5431"
echo "🔴  Redis:         localhost:6379"
echo "🐰  RabbitMQ:      localhost:5672 (UI: localhost:15672)"
echo "🤖  ML Service 1:  localhost:8001"
echo "🤖  ML Service 2:  localhost:8002"
echo "🤖  ML Service 3:  localhost:8003"

echo ""
echo "🧪 Testing Commands:"
echo "==================="
echo "# Test ML service health:"
echo "curl http://localhost:8001/health/detail"
echo ""
echo "# View RabbitMQ Management UI:"
echo "open http://localhost:15672"
echo ""
echo "# Test chunk search (requires existing chunks):"
echo "cd llm && python test_chunk_search.py"
echo ""
echo "# View logs:"
echo "docker-compose logs -f ml-service-1"

echo ""
echo "🛑 To stop all services:"
echo "======================="
echo "cd infra && docker-compose down"

cd ..
print_success "Distributed ML services setup complete! 🎉"