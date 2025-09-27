# Playwright UI Tests for Pod Metrics Dashboard

This directory contains Playwright Python tests for testing the UI functionality of the Pod Metrics Dashboard, specifically focusing on sorting functionality and dark mode toggle features.

## Overview

The test suite includes:
- **Sorting Tests**: Validate table column sorting functionality, sort direction indicators, and data ordering
- **Dark Mode Tests**: Verify theme toggling, persistence, visual changes, and localStorage behavior

## Test Architecture

### Page Object Model
- `pages/pod_metrics_page.py`: Contains the `PodMetricsPage` class that encapsulates UI interactions
- Provides reusable methods for interacting with table elements, theme toggle, and data validation

### Test Files
- `tests/test_sorting.py`: Comprehensive sorting functionality tests
- `tests/test_dark_mode.py`: Dark/light theme toggle and persistence tests
- `tests/conftest.py`: Pytest fixtures and configuration

## Setup Instructions

### 1. Install Dependencies

```bash
cd qa/playwright-tests

# Create a virtual environment (recommended)
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install Python dependencies
pip install -r requirements.txt

# Install Playwright browsers
playwright install chromium firefox webkit
```

### 2. Ensure Frontend is Running

The tests expect the Pod Metrics Dashboard to be running on `http://localhost:3000`.

```bash
# In the frontend directory
cd ../../frontend
npm install
npm start
```

The Playwright configuration includes a web server setup that can automatically start the frontend, but it's recommended to have it running separately for faster test execution.

## Running Tests

### Basic Test Execution

```bash
# Run all tests
pytest

# Run only sorting tests
pytest tests/test_sorting.py

# Run only dark mode tests
pytest tests/test_dark_mode.py

# Run tests in headed mode (visible browser)
pytest --headed

# Run tests with verbose output
pytest -v
```

### Advanced Options

```bash
# Run tests in parallel
pytest -n auto

# Run only fast tests (exclude visual tests)
pytest -m "not visual"

# Run with HTML report
pytest --html=report.html --self-contained-html

# Run specific test
pytest tests/test_sorting.py::TestSorting::test_sort_pod_name_ascending

# Run tests and keep browser open on failure (debugging)
pytest --headed --pdb
```

### Cross-Browser Testing

```bash
# Run on specific browser
pytest --browser chromium
pytest --browser firefox  
pytest --browser webkit

# Run on all browsers
pytest --browser chromium firefox webkit
```

## Test Categories

### Sorting Tests (`test_sorting.py`)

| Test | Description |
|------|-------------|
| `test_sortable_columns_exist` | Verifies sortable columns are identified correctly |
| `test_sort_pod_name_ascending` | Tests ascending sort on Pod Name column |
| `test_sort_pod_name_descending` | Tests descending sort on Pod Name column |
| `test_sort_cpu_usage_descending_default` | Validates CPU usage default sort direction |
| `test_sort_memory_usage` | Tests memory usage column sorting |
| `test_sort_toggle_direction` | Verifies sort direction toggles correctly |
| `test_sort_different_columns` | Tests switching between different column sorts |
| `test_sort_arrows_display_correctly` | Validates sort arrow indicators |
| `test_sorting_visual_validation` | Visual test with screenshots |

### Dark Mode Tests (`test_dark_mode.py`)

| Test | Description |
|------|-------------|
| `test_default_theme_is_dark` | Verifies default theme is dark mode |
| `test_toggle_to_light_mode` | Tests switching to light mode |
| `test_toggle_to_dark_mode` | Tests switching to dark mode |
| `test_theme_toggle_button_exists` | Validates theme toggle button presence |
| `test_theme_toggle_icon_changes` | Tests icon changes with theme |
| `test_background_color_changes` | Validates background color changes |
| `test_paper_component_styling_changes` | Tests Material-UI Paper component changes |
| `test_table_styling_changes` | Validates table styling changes |
| `test_theme_persistence_in_localstorage` | Tests localStorage theme persistence |
| `test_theme_persistence_after_refresh` | Validates theme persists after page refresh |
| `test_multiple_theme_toggles` | Tests multiple rapid theme changes |
| `test_app_bar_styling_changes` | Validates app bar styling changes |
| `test_dark_mode_visual_validation` | Visual comparison with screenshots |

## Configuration

### Environment Variables

- `BASE_URL`: Override the base URL for testing (default: `http://localhost:3000`)
- `HEADLESS`: Set to `false` to run tests in headed mode

```bash
# Run with custom base URL
BASE_URL=http://localhost:3001 pytest

# Run in headed mode
HEADLESS=false pytest
```

### Playwright Configuration

The `playwright.config.json` file contains:
- Test timeout and retry settings
- Cross-browser testing configuration
- Screenshot and video recording settings
- HTML report generation

### Pytest Configuration

The `conftest.py` file provides:
- Browser context setup
- Page fixtures with screenshot directory creation
- Custom test markers (`visual`, `slow`, `integration`)
- Environment variable configuration

## Screenshots and Reports

### Screenshots
Visual tests automatically capture screenshots saved to:
- `screenshots/` directory
- Named by test and state (e.g., `dark_mode.png`, `light_mode.png`)

### HTML Reports
Generate comprehensive HTML reports with:
```bash
pytest --html=report.html --self-contained-html
```

### Playwright Reports
Built-in Playwright HTML reporter:
```bash
playwright show-report
```

## Troubleshooting

### Common Issues

1. **Frontend not running**
   ```
   Error: connect ECONNREFUSED 127.0.0.1:3000
   ```
   **Solution**: Start the frontend application on port 3000

2. **Browser installation issues**
   ```bash
   playwright install --force
   ```

3. **Permission issues on Linux**
   ```bash
   playwright install-deps
   ```

4. **Timeout errors**
   - Increase timeout in `playwright.config.json`
   - Check if frontend is responding slowly
   - Ensure backend APIs are running

### Debug Mode

For debugging failing tests:
```bash
# Run single test in headed mode with debugging
pytest tests/test_dark_mode.py::TestDarkMode::test_toggle_to_light_mode --headed --pdb

# Add debug prints in tests
pytest -s  # Show print statements
```

### Updating Locators

If the UI changes and tests fail due to locator issues:
1. Inspect the updated UI elements
2. Update locators in `pages/pod_metrics_page.py`
3. Test locator changes with:
   ```bash
   playwright codegen http://localhost:3000
   ```

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Playwright Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          cd qa/playwright-tests
          pip install -r requirements.txt
          playwright install --with-deps chromium
      
      - name: Start frontend
        run: |
          cd frontend
          npm ci
          npm start &
          sleep 30
      
      - name: Run tests
        run: |
          cd qa/playwright-tests
          pytest --browser chromium
      
      - name: Upload reports
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: qa/playwright-tests/playwright-report/
```

## Contributing

When adding new tests:
1. Follow the existing Page Object Model pattern
2. Add appropriate test markers (`@pytest.mark.visual` for screenshot tests)
3. Include descriptive test names and docstrings
4. Update this README with new test descriptions

## Test Coverage

Current test coverage includes:
- ✅ Table sorting (all columns)
- ✅ Sort direction indicators
- ✅ Theme toggle functionality
- ✅ Theme persistence
- ✅ Visual validation with screenshots
- ✅ Cross-browser compatibility
- ✅ Material-UI component theme changes

Future enhancements could include:
- Namespace filter testing
- Search functionality testing
- Performance testing
- Mobile responsiveness testing
- API integration testing
