import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// Mock the API module with a simplified setup
jest.mock('./services/api', () => ({
  fetchNamespaces: () => Promise.resolve(['default']),
  fetchPodMetrics: () => Promise.resolve([
    {
      name: 'test-pod',
      namespace: 'default',
      containerName: 'test-container',
      cpu: {
        usage: '100m', request: '200m', limit: '500m',
        usageValue: 100, requestValue: 200, limitValue: 500,
        requestPercentage: 50, limitPercentage: 20
      },
      memory: {
        usage: '128Mi', request: '256Mi', limit: '512Mi',
        usageValue: 128, requestValue: 256, limitValue: 512,
        requestPercentage: 50, limitPercentage: 25
      },
      labels: null // This would previously crash the app
    }
  ]),
  fetchPodSummary: () => Promise.resolve({
    totalPods: 1, averageCpuUsage: 50, averageMemoryUsage: 50,
    highCpuPods: 0, highMemoryPods: 0, lowCpuPods: 0, lowMemoryPods: 0,
    generatedAt: new Date().toISOString()
  }),
}));

// Mock the theme context
jest.mock('./theme/ThemeContext', () => ({
  useTheme: () => ({
    mode: 'light',
    toggleTheme: jest.fn(),
  }),
}));

describe('Search Fix Verification', () => {
  test('search should not crash when pod has null labels', async () => {
    // Render the app
    render(<App />);

    // Wait for pod to load
    await waitFor(() => {
      expect(screen.getByText('test-pod')).toBeInTheDocument();
    }, { timeout: 10000 });

    // Find search input
    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...');
    
    // Perform search - this should NOT crash the app
    await userEvent.type(searchInput, 'test');
    
    // Verify the app is still functional and shows the pod
    await waitFor(() => {
      expect(screen.getByText('test-pod')).toBeInTheDocument();
    });

    // Test passes if we reach here without errors
    expect(true).toBe(true);
  });
});
