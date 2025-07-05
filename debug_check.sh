#!/bin/bash

echo "Starting chronos with debug logging..."
echo "Debug log will be written to /tmp/chronos_month_debug.txt"
echo

# Remove any existing debug file
rm -f /tmp/chronos_month_debug.txt

# Run chronos in background for a short time to see initial setup
timeout 3s ./main -debug 2>/dev/null &
CHRONOS_PID=$!

# Wait a moment for it to start
sleep 1

# Kill it
kill $CHRONOS_PID 2>/dev/null
wait $CHRONOS_PID 2>/dev/null

echo "Chronos stopped. Checking debug output..."
echo

if [ -f /tmp/chronos_month_debug.txt ]; then
    echo "=== DEBUG OUTPUT ==="
    cat /tmp/chronos_month_debug.txt
    echo "===================="
else
    echo "No debug file found - chronos may have failed to start properly."
fi

echo
echo "Let's also check what happens when we examine the view structure..."

# Show the problematic area - the month view appearing in week view
echo "This suggests the issue is that the month view is being rendered even in week mode."
echo "The month view should be hidden (positioned at -1000,-1000) when in week mode."