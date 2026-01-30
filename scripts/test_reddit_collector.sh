#!/bin/bash

# Reddit Collector Test Script
# Usage: ./scripts/test_reddit_collector.sh [unit|integration|all]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo -e "${GREEN}=== Reddit Collector Test Suite ===${NC}\n"

# Parse command line argument
TEST_TYPE="${1:-all}"

run_unit_tests() {
    echo -e "${YELLOW}Running Unit Tests...${NC}"
    go test -v ./internal/collector -run TestRedditCollector_ 2>&1 | tee /tmp/reddit_unit_tests.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo -e "\n${GREEN}✓ Unit tests passed${NC}\n"
        return 0
    else
        echo -e "\n${RED}✗ Unit tests failed${NC}\n"
        return 1
    fi
}

run_integration_tests() {
    echo -e "${YELLOW}Running Integration Tests (requires network)...${NC}"
    echo -e "${YELLOW}This will make real requests to Reddit RSS feeds${NC}\n"
    
    go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI 2>&1 | tee /tmp/reddit_integration_tests.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo -e "\n${GREEN}✓ Integration tests passed${NC}\n"
        return 0
    else
        echo -e "\n${RED}✗ Integration tests failed${NC}\n"
        echo -e "${YELLOW}Note: Integration tests may fail due to network issues or Reddit rate limiting${NC}\n"
        return 1
    fi
}

run_coverage() {
    echo -e "${YELLOW}Generating Coverage Report...${NC}"
    go test -coverprofile=coverage.out ./internal/collector -run TestRedditCollector_
    
    if [ -f coverage.out ]; then
        go tool cover -func=coverage.out | grep "total:" | awk '{print $3}'
        echo -e "${GREEN}Coverage report generated: coverage.out${NC}"
        echo -e "${YELLOW}To view HTML report: go tool cover -html=coverage.out${NC}\n"
    fi
}

# Main execution
case "$TEST_TYPE" in
    unit)
        run_unit_tests
        EXIT_CODE=$?
        ;;
    integration)
        run_integration_tests
        EXIT_CODE=$?
        ;;
    coverage)
        run_coverage
        EXIT_CODE=$?
        ;;
    all)
        UNIT_RESULT=0
        INTEGRATION_RESULT=0
        
        run_unit_tests || UNIT_RESULT=$?
        echo ""
        run_integration_tests || INTEGRATION_RESULT=$?
        echo ""
        run_coverage
        
        if [ $UNIT_RESULT -eq 0 ] && [ $INTEGRATION_RESULT -eq 0 ]; then
            echo -e "${GREEN}=== All tests passed! ===${NC}"
            EXIT_CODE=0
        else
            echo -e "${RED}=== Some tests failed ===${NC}"
            [ $UNIT_RESULT -ne 0 ] && echo -e "${RED}  - Unit tests failed${NC}"
            [ $INTEGRATION_RESULT -ne 0 ] && echo -e "${RED}  - Integration tests failed${NC}"
            EXIT_CODE=1
        fi
        ;;
    *)
        echo -e "${RED}Invalid test type: $TEST_TYPE${NC}"
        echo "Usage: $0 [unit|integration|coverage|all]"
        exit 1
        ;;
esac

echo -e "\n${YELLOW}Test logs saved to:${NC}"
[ -f /tmp/reddit_unit_tests.log ] && echo "  - Unit tests: /tmp/reddit_unit_tests.log"
[ -f /tmp/reddit_integration_tests.log ] && echo "  - Integration tests: /tmp/reddit_integration_tests.log"

exit $EXIT_CODE
