from playwright.async_api import Page, expect
from typing import List


class PodMetricsPage:
    def __init__(self, page: Page):
        self.page = page
        
        # Locators
        self.app_bar = page.locator('[data-testid="app-bar"], .MuiAppBar-root')
        self.theme_toggle_button = page.locator('button[aria-label*="mode"], button:has(svg[data-testid="Brightness4Icon"], svg[data-testid="Brightness7Icon"])')
        self.pod_table = page.locator('.MuiTable-root')
        self.table_headers = page.locator('.MuiTableHead-root .MuiTableCell-root')
        self.table_rows = page.locator('.MuiTableBody-root .MuiTableRow-root')
        self.refresh_button = page.locator('button:has(svg[data-testid="RefreshIcon"])')
        self.loading_spinner = page.locator('.MuiCircularProgress-root')
        
        # Background elements to check theme
        self.main_container = page.locator('body, .MuiContainer-root').first()
        self.paper_elements = page.locator('.MuiPaper-root')
        
    async def navigate(self, url: str = "http://localhost:3000"):
        """Navigate to the pod metrics page"""
        await self.page.goto(url)
        await self.wait_for_page_load()
        
    async def wait_for_page_load(self):
        """Wait for page to fully load"""
        # Wait for the table to be visible
        await self.pod_table.wait_for(state="visible", timeout=10000)
        # Wait for any loading spinners to disappear
        await self.page.wait_for_function("document.readyState === 'complete'")
        
    async def get_current_theme(self) -> str:
        """Determine current theme by checking background color"""
        # Check body background color or container background
        bg_color = await self.page.evaluate('''() => {
            const body = document.body;
            const computedStyle = window.getComputedStyle(body);
            return computedStyle.backgroundColor;
        }''')
        
        # Dark theme typically has dark backgrounds
        if 'rgb(18, 18, 18)' in bg_color or '18, 18, 18' in bg_color:
            return 'dark'
        elif 'rgb(245, 245, 245)' in bg_color or '245, 245, 245' in bg_color:
            return 'light'
        else:
            # Fallback: check if we can find dark/light mode indicators
            theme_button_icon = await self.theme_toggle_button.locator('svg').first().get_attribute('data-testid')
            if 'Brightness7Icon' in str(theme_button_icon):  # Sun icon means currently dark
                return 'dark'
            else:  # Moon icon means currently light
                return 'light'
    
    async def toggle_theme(self):
        """Click the theme toggle button"""
        await self.theme_toggle_button.click()
        # Wait for theme transition to complete
        await self.page.wait_for_timeout(500)
        
    async def get_sortable_columns(self) -> List[str]:
        """Get list of sortable column names"""
        clickable_headers = self.table_headers.filter(has=self.page.locator(':scope[style*="cursor: pointer"], :scope[style*="cursor:pointer"]'))
        column_names = []
        count = await clickable_headers.count()
        for i in range(count):
            text = await clickable_headers.nth(i).text_content()
            # Clean up the text (remove sort arrows)
            clean_text = text.replace(' ↑', '').replace(' ↓', '').strip()
            column_names.append(clean_text)
        return column_names
        
    async def click_column_header(self, column_name: str):
        """Click on a column header to sort"""
        # Find header containing the column name
        header = self.table_headers.filter(has_text=column_name).first()
        await header.click()
        # Wait for sort to complete
        await self.page.wait_for_timeout(300)
        
    async def get_column_sort_direction(self, column_name: str) -> str:
        """Get the current sort direction for a column (asc, desc, or none)"""
        header = self.table_headers.filter(has_text=column_name).first()
        header_text = await header.text_content()
        
        if ' ↑' in header_text:
            return 'asc'
        elif ' ↓' in header_text:
            return 'desc'
        else:
            return 'none'
            
    async def get_column_data(self, column_index: int) -> List[str]:
        """Get all data from a specific column"""
        column_cells = self.table_rows.locator(f'.MuiTableCell-root:nth-child({column_index + 1})')
        count = await column_cells.count()
        data = []
        for i in range(count):
            cell_text = await column_cells.nth(i).text_content()
            data.append(cell_text.strip())
        return data
        
    async def get_table_data(self) -> List[List[str]]:
        """Get all table data as a list of rows"""
        rows_data = []
        row_count = await self.table_rows.count()
        
        for i in range(row_count):
            row = self.table_rows.nth(i)
            cells = row.locator('.MuiTableCell-root')
            cell_count = await cells.count()
            
            row_data = []
            for j in range(cell_count):
                cell_text = await cells.nth(j).text_content()
                row_data.append(cell_text.strip())
            rows_data.append(row_data)
            
        return rows_data
        
    async def wait_for_data_load(self):
        """Wait for data to load in the table"""
        # Wait for at least one data row to appear
        await self.table_rows.first().wait_for(state="visible", timeout=15000)
        # Wait for loading spinners to disappear
        loading_count = await self.loading_spinner.count()
        if loading_count > 0:
            await self.loading_spinner.first().wait_for(state="hidden", timeout=10000)
            
    async def refresh_data(self):
        """Click refresh button and wait for data to reload"""
        await self.refresh_button.click()
        await self.wait_for_data_load()
        
    async def take_screenshot(self, name: str):
        """Take a screenshot for visual validation"""
        await self.page.screenshot(path=f"qa/playwright-tests/screenshots/{name}.png", full_page=True)
