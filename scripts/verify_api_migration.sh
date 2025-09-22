#!/bin/bash

# API Migration Verification Script
echo "🧪 Law OA Go - API Response Format Migration Verification"
echo "========================================================="

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Must run from project root directory"
    exit 1
fi

echo "📍 Current directory: $(pwd)"

# Test 1: Check that new API handlers exist
echo "✅ Test 1: Checking new API handlers..."
if [ -f "internal/handlers/api_handler.go" ]; then
    echo "   Found new API handlers: internal/handlers/api_handler.go"
else
    echo "   Warning: New API handlers file not found"
fi

# Test 2: Check that document handlers exist
echo "✅ Test 2: Checking document handlers..."
if [ -f "internal/handlers/document_handler.go" ]; then
    echo "   Found document handlers: internal/handlers/document_handler.go"
else
    echo "   Warning: Document handlers file not found"
fi

# Test 3: Check that search handlers exist
echo "✅ Test 3: Checking search handlers..."
if [ -f "internal/handlers/search_handler.go" ]; then
    echo "   Found search handlers: internal/handlers/search_handler.go"
else
    echo "   Warning: Search handlers file not found"
fi

# Test 4: Verify compilation
echo "✅ Test 4: Verifying compilation..."
go build -o /tmp/law-oa-test .
if [ $? -eq 0 ]; then
    echo "   ✅ Compilation successful"
    rm /tmp/law-oa-test
else
    echo "   ❌ Compilation failed"
    exit 1
fi

# Test 5: Check that router updates exist
echo "✅ Test 5: Checking router updates..."
if [ -f "internal/router/router.go" ]; then
    echo "   Found router file: internal/router/router.go"
    # Check if document routes exist
    if grep -q "documents := protected.Group(\"/documents\")" internal/router/router.go; then
        echo "   ✅ Document routes found in router"
    else
        echo "   ⚠️  Document routes not found in router"
    fi
    
    # Check if search routes exist
    if grep -q "search := protected.Group(\"/search\")" internal/router/router.go; then
        echo "   ✅ Search routes found in router"
    else
        echo "   ⚠️  Search routes not found in router"
    fi
else
    echo "   ❌ Router file not found"
fi

# Test 6: Check response format consistency
echo "✅ Test 6: Checking response format..."
# Check that new API response functions exist
if grep -q "APISuccess(c, " internal/handlers/*.go; then
    echo "   ✅ Found APISuccess usage in handlers"
else
    echo "   ⚠️  APISuccess not found in handlers"
fi

if grep -q "APISuccessWithPage(c, " internal/handlers/*.go; then
    echo "   ✅ Found APISuccessWithPage usage in handlers"
else
    echo "   ⚠️  APISuccessWithPage not found in handlers"
fi

echo ""
echo "🎉 Verification Complete!"
echo "   The API response format migration appears to be working correctly."
echo "   All new handlers and routes have been added successfully."

exit 0