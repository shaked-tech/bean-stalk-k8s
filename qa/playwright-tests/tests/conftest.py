import pytest
import asyncio
import os
from playwright.async_api import async_playwright


@pytest.fixture(scope="session")
def event_loop():
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture(scope="session")
async def browser_context():
    """Create a browser context for the test session."""
    async with async_playwright() as p:
        browser = await p.chromium.launch(
            headless=True,  # Set to False for debugging
            args=['--no-sandbox', '--disable-dev-shm-usage']
        )
        context = await browser.new_context(
            viewport={'width': 1280, 'height': 720},
            # Clear any existing localStorage/cookies for consistent test state
            storage_state=None
        )
        yield context
        await browser.close()


@pytest.fixture
async def page(browser_context):
    """Create a new page for each test."""
    page = await browser_context.new_page()
    
    # Create screenshots directory if it doesn't exist
    os.makedirs('qa/playwright-tests/screenshots', exist_ok=True)
    
    yield page
    await page.close()


# Configure pytest markers
def pytest_configure(config):
    """Configure custom pytest markers."""
    config.addinivalue_line(
        "markers", "visual: marks tests as visual tests that take screenshots"
    )
    config.addinivalue_line(
        "markers", "slow: marks tests as slow running"
    )


# Custom test collection modifier
def pytest_collection_modifyitems(config, items):
    """Modify test collection to add markers based on test names."""
    for item in items:
        # Add slow marker to visual tests
        if "visual" in item.name:
            item.add_marker(pytest.mark.slow)
            
        # Add integration marker to tests that navigate
        if "test_" in item.name and hasattr(item, 'function'):
            if any(keyword in item.name.lower() for keyword in ['navigation', 'refresh', 'persistence']):
                item.add_marker(pytest.mark.integration)


# Fixture for setting test environment variables
@pytest.fixture(autouse=True)
def set_test_env():
    """Set environment variables for testing."""
    # Set base URL for testing (can be overridden via environment variable)
    if 'BASE_URL' not in os.environ:
        os.environ['BASE_URL'] = 'http://localhost:3000'
    yield
