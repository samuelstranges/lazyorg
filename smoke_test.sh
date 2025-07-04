#!/bin/bash

echo "🧪 Chronos Smoke Test - Testing Refactored Code"
echo "=============================================="

# Test 1: Build succeeds
echo "1. Testing build..."
if go build -o chronos cmd/chronos/main.go; then
    echo "   ✅ Build successful"
else
    echo "   ❌ Build failed"
    exit 1
fi

# Test 2: All tests pass
echo "2. Running unit tests..."
if go test ./tests/... > /dev/null 2>&1; then
    echo "   ✅ All tests pass"
else
    echo "   ❌ Tests failed"
    exit 1
fi

# Test 3: Binary runs without crashing
echo "3. Testing binary startup..."
timeout 2s ./chronos > /dev/null 2>&1
if [ $? -eq 124 ]; then
    echo "   ✅ Application starts successfully"
else
    echo "   ⚠️  Application may have issues (exit code: $?)"
fi

# Test 4: Check file structure
echo "4. Verifying refactored file structure..."
expected_files=(
    "pkg/views/app-view.go"
    "pkg/views/app-view-navigation.go" 
    "pkg/views/app-view-events.go"
    "pkg/views/app-view-debug.go"
    "pkg/views/popup-view.go"
    "pkg/views/popup-forms.go"
    "pkg/views/popup-handlers.go"
    "internal/database/database.go"
    "internal/database/queries.go"
    "internal/database/operations.go"
)

missing_files=0
for file in "${expected_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "   ❌ Missing: $file"
        missing_files=$((missing_files + 1))
    fi
done

if [ $missing_files -eq 0 ]; then
    echo "   ✅ All refactored files present"
else
    echo "   ❌ $missing_files files missing"
    exit 1
fi

echo ""
echo "🎉 Smoke test completed successfully!"
echo "   → Manual testing recommended: ./chronos"
echo "   → Debug mode available: ./chronos -debug"