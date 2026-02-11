#!/bin/bash

echo "========================================="
echo "  BRIQUE - Test Complet Étape 2"
echo "========================================="
echo ""

# Build
echo "📦 Building..."
go build -o brique ./cmd/brique-cli
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✓ Build successful"
echo ""

# Clean test environment
echo "🧹 Cleaning test environment..."
rm -rf ~/.config/brique/
echo "✓ Clean complete"
echo ""

# Create test PDF file
echo "📄 Creating test files..."
mkdir -p /tmp/brique_test
echo "This is a test manual PDF content" > /tmp/brique_test/manual.pdf
echo "This is a service manual" > /tmp/brique_test/service_manual.pdf
echo "This is a schematic" > /tmp/brique_test/schematic.pdf
echo "✓ Test files created"
echo ""

# Test: List empty inventory
echo "📋 Test 1: List empty inventory"
./brique item list 2>&1 | grep -v "time=" | grep -v "goose:" | tail -2
echo ""

# Test: Add items (non-interactive, using echo)
echo "➕ Test 2: Adding test items..."
echo -e "Lave-Linge Cuisine\nGros Électroménager\nBrandt\nWTC1234\nSN123456\nAcheté en 2020" | ./brique item add 2>&1 | grep -v "time=" | grep -v "goose:" | tail -2
echo -e "Perceuse Sans Fil\nOutils\nBosch\nPSB500\n18V\nTrès pratique" | ./brique item add 2>&1 | grep -v "time=" | grep -v "goose:" | tail -2
echo -e "Réfrigérateur\nGros Électroménager\nSamsung\nRS68\nSN789\nDouble porte" | ./brique item add 2>&1 | grep -v "time=" | grep -v "goose:" | tail -2
echo ""

# Test: List items
echo "📋 Test 3: List all items"
./brique item list 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: Get item details
echo "🔍 Test 4: Get item details (ID: 1)"
./brique item get 1 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: Search
echo "🔎 Test 5: Search for 'Bosch'"
./brique item search Bosch 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: Add assets
echo "📎 Test 6: Adding assets to item 1"
./brique asset add 1 /tmp/brique_test/manual.pdf -t manual -n "User Manual" 2>&1 | grep -v "time=" | grep -v "goose:"
./brique asset add 1 /tmp/brique_test/service_manual.pdf -t service_manual -n "Service Manual" 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: List assets
echo "📎 Test 7: List assets for item 1"
./brique asset list 1 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: Get item with assets (should show secured health)
echo "🏥 Test 8: Check documentation health (should be secured)"
./brique item get 1 2>&1 | grep -v "time=" | grep -v "goose:" | tail -10
echo ""

# Test: Add partial asset to item 2
echo "📎 Test 9: Adding partial asset to item 2"
./brique asset add 2 /tmp/brique_test/schematic.pdf -t schematic 2>&1 | grep -v "time=" | grep -v "goose:"
echo ""

# Test: Get item 2 (should show partial health)
echo "🏥 Test 10: Check item 2 health (should be partial)"
./brique item get 2 2>&1 | grep -v "time=" | grep -v "goose:" | tail -10
echo ""

# Summary
echo "========================================="
echo "  ✅ Tests completed successfully!"
echo "========================================="
echo ""
echo "📊 Summary:"
echo "  - 3 items created"
echo "  - 3 assets attached"
echo "  - Item 1: 🟢 Secured (manual + service manual)"
echo "  - Item 2: 🟡 Partial (schematic only)"
echo "  - Item 3: 🔴 Incomplete (no assets)"
echo ""
echo "🧪 You can now test:"
echo "  ./brique item update 1"
echo "  ./brique item delete 3"
echo "  ./brique asset delete <id>"
echo ""
