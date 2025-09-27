#!/bin/bash

# Playwright UI Tests Setup Script
# This script sets up the testing environment for the Pod Metrics Dashboard UI tests

set -e

echo "🚀 Setting up Playwright UI Tests for Pod Metrics Dashboard"
echo "============================================================"

# Check if we're in the right directory
if [[ ! -f "requirements.txt" ]]; then
    echo "❌ Error: Please run this script from the qa/playwright-tests directory"
    exit 1
fi

# Create virtual environment
echo "📦 Creating virtual environment..."
python3 -m venv venv

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Install Python dependencies
echo "📚 Installing Python dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

# Install Playwright browsers
echo "🌐 Installing Playwright browsers..."
playwright install chromium firefox webkit

# Install system dependencies (Linux)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "🔧 Installing system dependencies..."
    playwright install-deps
fi

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p screenshots
mkdir -p playwright-report

# Check if frontend is running
echo "🔍 Checking if frontend is available..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend is running on http://localhost:3000"
else
    echo "⚠️  Frontend is not running on http://localhost:3000"
    echo "   Please start the frontend with:"
    echo "   cd ../../frontend && npm start"
fi

# Run a simple test to verify setup
echo "🧪 Running a quick test to verify setup..."
if python -m pytest tests/test_dark_mode.py::TestDarkMode::test_theme_toggle_button_exists -v --tb=short; then
    echo "✅ Setup completed successfully!"
    echo ""
    echo "🎉 Ready to run tests!"
    echo ""
    echo "Usage examples:"
    echo "  pytest                                    # Run all tests"
    echo "  pytest tests/test_sorting.py             # Run sorting tests only"
    echo "  pytest tests/test_dark_mode.py           # Run dark mode tests only"
    echo "  pytest --headed                          # Run tests with visible browser"
    echo "  pytest --html=report.html                # Generate HTML report"
    echo ""
    echo "For more options, see README.md"
else
    echo "❌ Setup test failed. Please check the frontend is running and try again."
    exit 1
fi
