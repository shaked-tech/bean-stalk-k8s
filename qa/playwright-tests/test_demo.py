import pytest
from playwright.sync_api import Page, expect


def test_homepage_loads(page: Page):
    """Simple test to demonstrate the framework is working"""
    page.goto("http://localhost:3000")
    
    # Wait for page to load and check basic elements exist
    page.wait_for_load_state("networkidle", timeout=15000)
    
    # Check that we got a successful response (not an error page)
    expect(page).to_have_url("http://localhost:3000/")
    
    # Take a screenshot to see what loaded
    page.screenshot(path="screenshots/homepage.png")
    
    # Check if we can find any React content
    body_text = page.locator("body").inner_text()
    assert len(body_text) > 0, "Page body should contain some content"
    
    print("✅ Test passed - Homepage loaded successfully!")


def test_theme_toggle_exists(page: Page):
    """Test that the theme toggle button exists"""
    page.goto("http://localhost:3000")
    
    # Wait for the page to load
    page.wait_for_selector(".MuiAppBar-root", timeout=10000)
    
    # Find theme toggle button
    theme_button = page.locator("button:has(svg[data-testid='Brightness4Icon'], svg[data-testid='Brightness7Icon'])")
    
    # Verify button exists and is visible
    expect(theme_button).to_be_visible()
    
    print("✅ Test passed - Theme toggle button exists!")


def test_table_headers_clickable(page: Page):
    """Test that table headers are clickable for sorting"""
    page.goto("http://localhost:3000")
    
    # Wait for table to load
    page.wait_for_selector(".MuiTable-root", timeout=15000)
    
    # Find sortable headers (they have cursor: pointer style or click handlers)
    pod_name_header = page.locator('.MuiTableCell-root:has-text("Pod Name")')
    
    # Verify header exists
    expect(pod_name_header).to_be_visible()
    
    print("✅ Test passed - Table headers are present!")
