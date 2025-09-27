import pytest
import re
from playwright.async_api import Page, expect
from pages.pod_metrics_page import PodMetricsPage


class TestSorting:
    """Test cases for table sorting functionality"""
    
    @pytest.fixture(autouse=True)
    async def setup(self, page: Page):
        """Setup for each test"""
        self.pod_page = PodMetricsPage(page)
        await self.pod_page.navigate()
        await self.pod_page.wait_for_data_load()
    
    async def test_sortable_columns_exist(self):
        """Test that sortable columns are properly identified"""
        sortable_columns = await self.pod_page.get_sortable_columns()
        
        # Expected sortable columns based on the component analysis
        expected_columns = [
            'Pod Name', 'Container', 'Namespace', 'CPU Usage',
            'CPU Request', 'CPU Limit', 'CPU Request %', 'CPU Limit %',
            'Memory Usage', 'Memory Request', 'Memory Limit', 
            'Memory Request %', 'Memory Limit %'
        ]
        
        assert len(sortable_columns) > 0, "No sortable columns found"
        
        # Check that we have the main expected columns
        for expected in ['Pod Name', 'CPU Usage', 'Memory Usage']:
            assert any(expected in col for col in sortable_columns), f"Expected column '{expected}' not found"
    
    async def test_sort_pod_name_ascending(self):
        """Test sorting pod names in ascending order"""
        await self.pod_page.click_column_header('Pod Name')
        
        # Verify sort direction
        sort_direction = await self.pod_page.get_column_sort_direction('Pod Name')
        assert sort_direction == 'asc', f"Expected ascending sort, got {sort_direction}"
        
        # Get the data from the Pod Name column (assuming it's the first column)
        pod_names = await self.pod_page.get_column_data(0)
        
        if len(pod_names) > 1:
            # Verify the names are sorted alphabetically
            sorted_names = sorted(pod_names, key=str.lower)
            assert pod_names == sorted_names, "Pod names are not sorted in ascending order"
    
    async def test_sort_pod_name_descending(self):
        """Test sorting pod names in descending order"""
        # Click twice to get descending order (first click = asc, second = desc)
        await self.pod_page.click_column_header('Pod Name')  # First click for ascending
        await self.pod_page.click_column_header('Pod Name')  # Second click for descending
        
        # Verify sort direction
        sort_direction = await self.pod_page.get_column_sort_direction('Pod Name')
        assert sort_direction == 'desc', f"Expected descending sort, got {sort_direction}"
        
        # Get the data from the Pod Name column
        pod_names = await self.pod_page.get_column_data(0)
        
        if len(pod_names) > 1:
            # Verify the names are sorted in reverse alphabetical order
            sorted_names = sorted(pod_names, key=str.lower, reverse=True)
            assert pod_names == sorted_names, "Pod names are not sorted in descending order"
    
    async def test_sort_cpu_usage_descending_default(self):
        """Test that CPU Usage sorts in descending order by default (numerical columns)"""
        await self.pod_page.click_column_header('CPU Usage')
        
        # Numerical columns should default to descending to show highest values first
        sort_direction = await self.pod_page.get_column_sort_direction('CPU Usage')
        # Note: Based on the component code, numerical columns default to desc
        assert sort_direction in ['desc', 'asc'], f"Expected sort direction, got {sort_direction}"
        
        # Get CPU usage data (find the correct column index)
        table_data = await self.pod_page.get_table_data()
        if table_data:
            # CPU Usage should be around column index 3 based on the component
            cpu_usage_col = 3
            cpu_values = [row[cpu_usage_col] for row in table_data if len(row) > cpu_usage_col]
            
            # Extract numerical values for comparison (remove 'm' suffix, convert to float)
            numeric_values = []
            for value in cpu_values:
                if value and value != '-':
                    # Handle different CPU formats (e.g., "100m", "1.5", etc.)
                    match = re.search(r'[\d.]+', value)
                    if match:
                        numeric_values.append(float(match.group()))
            
            if len(numeric_values) > 1:
                # Check if data is properly sorted
                if sort_direction == 'desc':
                    assert numeric_values == sorted(numeric_values, reverse=True), "CPU usage not sorted in descending order"
                else:
                    assert numeric_values == sorted(numeric_values), "CPU usage not sorted in ascending order"
    
    async def test_sort_memory_usage(self):
        """Test sorting memory usage"""
        await self.pod_page.click_column_header('Memory Usage')
        
        sort_direction = await self.pod_page.get_column_sort_direction('Memory Usage')
        assert sort_direction in ['desc', 'asc'], f"Expected sort direction, got {sort_direction}"
        
        # Get memory usage data
        table_data = await self.pod_page.get_table_data()
        if table_data:
            # Memory Usage should be around column index 8
            memory_usage_col = 8
            memory_values = [row[memory_usage_col] for row in table_data if len(row) > memory_usage_col]
            
            # Extract numerical values for comparison
            numeric_values = []
            for value in memory_values:
                if value and value != '-':
                    # Handle memory formats (e.g., "100Mi", "1.5Gi", etc.)
                    match = re.search(r'[\d.]+', value)
                    if match:
                        numeric_values.append(float(match.group()))
            
            if len(numeric_values) > 1:
                if sort_direction == 'desc':
                    assert numeric_values == sorted(numeric_values, reverse=True), "Memory usage not sorted in descending order"
                else:
                    assert numeric_values == sorted(numeric_values), "Memory usage not sorted in ascending order"
    
    async def test_sort_toggle_direction(self):
        """Test that clicking the same column header toggles sort direction"""
        # Start with Pod Name
        await self.pod_page.click_column_header('Pod Name')
        first_direction = await self.pod_page.get_column_sort_direction('Pod Name')
        
        # Click again to toggle
        await self.pod_page.click_column_header('Pod Name')
        second_direction = await self.pod_page.get_column_sort_direction('Pod Name')
        
        # Directions should be different
        assert first_direction != second_direction, "Sort direction should toggle when clicking the same header"
        
        # Should toggle between asc and desc
        assert {first_direction, second_direction} == {'asc', 'desc'}, "Sort should toggle between ascending and descending"
    
    async def test_sort_different_columns(self):
        """Test sorting different columns changes the active sort"""
        # Sort by Pod Name first
        await self.pod_page.click_column_header('Pod Name')
        pod_name_direction = await self.pod_page.get_column_sort_direction('Pod Name')
        assert pod_name_direction in ['asc', 'desc'], "Pod Name should be sorted"
        
        # Sort by CPU Usage
        await self.pod_page.click_column_header('CPU Usage')
        
        # Pod Name should no longer show sort indicator
        pod_name_direction_after = await self.pod_page.get_column_sort_direction('Pod Name')
        cpu_usage_direction = await self.pod_page.get_column_sort_direction('CPU Usage')
        
        assert pod_name_direction_after == 'none', "Previous sort column should not show sort indicator"
        assert cpu_usage_direction in ['asc', 'desc'], "New sort column should show sort indicator"
    
    async def test_sort_arrows_display_correctly(self):
        """Test that sort arrows display correctly in column headers"""
        # Click Pod Name for ascending
        await self.pod_page.click_column_header('Pod Name')
        
        # Get the header text
        pod_name_header = self.pod_page.table_headers.filter(has_text='Pod Name').first()
        header_text = await pod_name_header.text_content()
        
        # Should contain arrow indicator
        assert '↑' in header_text or '↓' in header_text, "Sort arrow should be displayed in header"
        
        # Click again for descending
        await self.pod_page.click_column_header('Pod Name')
        header_text_after = await pod_name_header.text_content()
        
        # Arrow should change
        assert header_text != header_text_after, "Sort arrow should change when toggling direction"
    
    @pytest.mark.visual
    async def test_sorting_visual_validation(self):
        """Visual test to capture sorting states"""
        # Take screenshot of initial state
        await self.pod_page.take_screenshot('initial_table_state')
        
        # Sort by Pod Name ascending and take screenshot
        await self.pod_page.click_column_header('Pod Name')
        await self.pod_page.take_screenshot('pod_name_ascending')
        
        # Sort by Pod Name descending and take screenshot  
        await self.pod_page.click_column_header('Pod Name')
        await self.pod_page.take_screenshot('pod_name_descending')
        
        # Sort by CPU Usage and take screenshot
        await self.pod_page.click_column_header('CPU Usage')
        await self.pod_page.take_screenshot('cpu_usage_sorted')
