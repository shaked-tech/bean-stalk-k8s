import pytest
from playwright.async_api import Page, expect
from pages.pod_metrics_page import PodMetricsPage


class TestDarkMode:
    """Test cases for dark mode functionality"""
    
    @pytest.fixture(autouse=True)
    async def setup(self, page: Page):
        """Setup for each test"""
        self.pod_page = PodMetricsPage(page)
        await self.pod_page.navigate()
        await self.pod_page.wait_for_page_load()
    
    async def test_default_theme_is_dark(self):
        """Test that the application defaults to dark mode"""
        current_theme = await self.pod_page.get_current_theme()
        assert current_theme == 'dark', f"Expected default theme to be dark, got {current_theme}"
    
    async def test_toggle_to_light_mode(self):
        """Test switching from dark to light mode"""
        # Ensure we start in dark mode
        initial_theme = await self.pod_page.get_current_theme()
        if initial_theme != 'dark':
            await self.pod_page.toggle_theme()  # Switch to dark first
            await self.pod_page.page.wait_for_timeout(500)
        
        # Now toggle to light mode
        await self.pod_page.toggle_theme()
        
        # Verify theme changed to light
        new_theme = await self.pod_page.get_current_theme()
        assert new_theme == 'light', f"Expected light theme after toggle, got {new_theme}"
    
    async def test_toggle_to_dark_mode(self):
        """Test switching from light to dark mode"""
        # Ensure we start in light mode
        initial_theme = await self.pod_page.get_current_theme()
        if initial_theme != 'light':
            await self.pod_page.toggle_theme()  # Switch to light first
            await self.pod_page.page.wait_for_timeout(500)
        
        # Now toggle to dark mode
        await self.pod_page.toggle_theme()
        
        # Verify theme changed to dark
        new_theme = await self.pod_page.get_current_theme()
        assert new_theme == 'dark', f"Expected dark theme after toggle, got {new_theme}"
    
    async def test_theme_toggle_button_exists(self):
        """Test that the theme toggle button is present"""
        theme_button = self.pod_page.theme_toggle_button
        await expect(theme_button).to_be_visible()
        
        # Button should be clickable
        await expect(theme_button).to_be_enabled()
    
    async def test_theme_toggle_icon_changes(self):
        """Test that the theme toggle icon changes based on current theme"""
        # Get initial icon
        initial_icon = await self.pod_page.theme_toggle_button.locator('svg').first().get_attribute('data-testid')
        
        # Toggle theme
        await self.pod_page.toggle_theme()
        
        # Get icon after toggle
        new_icon = await self.pod_page.theme_toggle_button.locator('svg').first().get_attribute('data-testid')
        
        # Icons should be different
        assert initial_icon != new_icon, "Theme toggle icon should change when theme changes"
        
        # Should be one of the expected icons
        expected_icons = ['Brightness4Icon', 'Brightness7Icon']
        assert initial_icon in expected_icons, f"Initial icon {initial_icon} not in expected icons"
        assert new_icon in expected_icons, f"New icon {new_icon} not in expected icons"
    
    async def test_background_color_changes(self):
        """Test that background colors change with theme"""
        # Get initial background color
        initial_bg = await self.pod_page.page.evaluate('''() => {
            const body = document.body;
            return window.getComputedStyle(body).backgroundColor;
        }''')
        
        # Toggle theme
        await self.pod_page.toggle_theme()
        
        # Get background color after toggle
        new_bg = await self.pod_page.page.evaluate('''() => {
            const body = document.body;
            return window.getComputedStyle(body).backgroundColor;
        }''')
        
        # Background colors should be different
        assert initial_bg != new_bg, "Background color should change when theme changes"
        
        # Verify specific theme colors
        dark_bg_patterns = ['rgb(18, 18, 18)', '18, 18, 18']
        light_bg_patterns = ['rgb(245, 245, 245)', '245, 245, 245']
        
        # One should be dark and one should be light
        is_initial_dark = any(pattern in initial_bg for pattern in dark_bg_patterns)
        is_new_dark = any(pattern in new_bg for pattern in dark_bg_patterns)
        is_initial_light = any(pattern in initial_bg for pattern in light_bg_patterns)
        is_new_light = any(pattern in new_bg for pattern in light_bg_patterns)
        
        # Should toggle between light and dark
        assert is_initial_dark != is_new_dark or is_initial_light != is_new_light, "Should toggle between light and dark backgrounds"
    
    async def test_paper_component_styling_changes(self):
        """Test that Material-UI Paper components change styling with theme"""
        # Get initial paper background color
        initial_paper_bg = await self.pod_page.page.evaluate('''() => {
            const paper = document.querySelector('.MuiPaper-root');
            if (paper) {
                return window.getComputedStyle(paper).backgroundColor;
            }
            return null;
        }''')
        
        # Toggle theme
        await self.pod_page.toggle_theme()
        
        # Get paper background after toggle
        new_paper_bg = await self.pod_page.page.evaluate('''() => {
            const paper = document.querySelector('.MuiPaper-root');
            if (paper) {
                return window.getComputedStyle(paper).backgroundColor;
            }
            return null;
        }''')
        
        # Paper backgrounds should be different (if paper elements exist)
        if initial_paper_bg and new_paper_bg:
            assert initial_paper_bg != new_paper_bg, "Paper component background should change with theme"
    
    async def test_table_styling_changes(self):
        """Test that table styling changes with theme"""
        # Get initial table background
        initial_table_bg = await self.pod_page.page.evaluate('''() => {
            const table = document.querySelector('.MuiTable-root');
            if (table) {
                return window.getComputedStyle(table).backgroundColor;
            }
            return null;
        }''')
        
        # Toggle theme
        await self.pod_page.toggle_theme()
        
        # Get table background after toggle
        new_table_bg = await self.pod_page.page.evaluate('''() => {
            const table = document.querySelector('.MuiTable-root');
            if (table) {
                return window.getComputedStyle(table).backgroundColor;
            }
            return null;
        }''')
        
        # Table backgrounds should be different (if table exists)
        if initial_table_bg and new_table_bg:
            assert initial_table_bg != new_table_bg, "Table background should change with theme"
    
    async def test_theme_persistence_in_localstorage(self):
        """Test that theme preference is stored in localStorage"""
        # Set to light mode
        current_theme = await self.pod_page.get_current_theme()
        if current_theme != 'light':
            await self.pod_page.toggle_theme()
            await self.pod_page.page.wait_for_timeout(500)
        
        # Check localStorage
        light_storage_value = await self.pod_page.page.evaluate("() => localStorage.getItem('themeMode')")
        assert light_storage_value == 'light', f"Expected 'light' in localStorage, got {light_storage_value}"
        
        # Toggle to dark mode
        await self.pod_page.toggle_theme()
        await self.pod_page.page.wait_for_timeout(500)
        
        # Check localStorage again
        dark_storage_value = await self.pod_page.page.evaluate("() => localStorage.getItem('themeMode')")
        assert dark_storage_value == 'dark', f"Expected 'dark' in localStorage, got {dark_storage_value}"
    
    async def test_theme_persistence_after_refresh(self):
        """Test that theme persists after page refresh"""
        # Set to light mode
        current_theme = await self.pod_page.get_current_theme()
        if current_theme != 'light':
            await self.pod_page.toggle_theme()
            await self.pod_page.page.wait_for_timeout(500)
        
        # Refresh the page
        await self.pod_page.page.reload()
        await self.pod_page.wait_for_page_load()
        
        # Theme should still be light
        theme_after_refresh = await self.pod_page.get_current_theme()
        assert theme_after_refresh == 'light', f"Expected light theme to persist after refresh, got {theme_after_refresh}"
        
        # Now test with dark mode
        await self.pod_page.toggle_theme()
        await self.pod_page.page.wait_for_timeout(500)
        
        # Refresh again
        await self.pod_page.page.reload()
        await self.pod_page.wait_for_page_load()
        
        # Theme should be dark
        final_theme = await self.pod_page.get_current_theme()
        assert final_theme == 'dark', f"Expected dark theme to persist after refresh, got {final_theme}"
    
    async def test_multiple_theme_toggles(self):
        """Test multiple rapid theme toggles work correctly"""
        initial_theme = await self.pod_page.get_current_theme()
        
        # Toggle multiple times
        for i in range(4):
            await self.pod_page.toggle_theme()
            await self.pod_page.page.wait_for_timeout(300)  # Short wait between toggles
        
        # After even number of toggles, should be back to initial theme
        final_theme = await self.pod_page.get_current_theme()
        assert final_theme == initial_theme, f"After 4 toggles, theme should be back to initial ({initial_theme}), got {final_theme}"
    
    async def test_app_bar_styling_changes(self):
        """Test that app bar styling changes with theme"""
        app_bar = self.pod_page.app_bar
        await expect(app_bar).to_be_visible()
        
        # Get initial app bar background
        initial_app_bar_bg = await self.pod_page.page.evaluate('''() => {
            const appBar = document.querySelector('.MuiAppBar-root');
            if (appBar) {
                return window.getComputedStyle(appBar).backgroundColor;
            }
            return null;
        }''')
        
        # Toggle theme
        await self.pod_page.toggle_theme()
        
        # Get app bar background after toggle
        new_app_bar_bg = await self.pod_page.page.evaluate('''() => {
            const appBar = document.querySelector('.MuiAppBar-root');
            if (appBar) {
                return window.getComputedStyle(appBar).backgroundColor;
            }
            return null;
        }''')
        
        # App bar backgrounds should be different
        if initial_app_bar_bg and new_app_bar_bg:
            assert initial_app_bar_bg != new_app_bar_bg, "App bar background should change with theme"
    
    @pytest.mark.visual
    async def test_dark_mode_visual_validation(self):
        """Visual test to capture dark and light mode states"""
        # Ensure we start in dark mode
        current_theme = await self.pod_page.get_current_theme()
        if current_theme != 'dark':
            await self.pod_page.toggle_theme()
            await self.pod_page.page.wait_for_timeout(500)
        
        # Take screenshot of dark mode
        await self.pod_page.take_screenshot('dark_mode')
        
        # Switch to light mode
        await self.pod_page.toggle_theme()
        await self.pod_page.page.wait_for_timeout(500)
        
        # Take screenshot of light mode
        await self.pod_page.take_screenshot('light_mode')
        
        # Switch back to dark mode
        await self.pod_page.toggle_theme()
        await self.pod_page.page.wait_for_timeout(500)
        
        # Take final screenshot to verify we're back to dark mode
        await self.pod_page.take_screenshot('dark_mode_restored')
