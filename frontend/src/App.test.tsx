import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// Mock the API module
jest.mock('./services/api', () => ({
  fetchNamespaces: jest.fn(),
  fetchPodMetrics: jest.fn(),
  fetchPodSummary: jest.fn(),
}));

// Mock the theme context
jest.mock('./theme/ThemeContext', () => ({
  useTheme: () => ({
    mode: 'light',
    toggleTheme: jest.fn(),
  }),
}));

const { fetchNamespaces, fetchPodMetrics, fetchPodSummary } = require('./services/api');

const mockNamespaces = ['default', 'kube-system'];

const mockPodsWithNullLabels = [
  {
    name: 'test-pod-1',
    namespace: 'default',
    containerName: 'test-container',
    cpu: {
      usage: '100m',
      request: '200m', 
      limit: '500m',
      usageValue: 100,
      requestValue: 200,
      limitValue: 500,
      requestPercentage: 50,
      limitPercentage: 20
    },
    memory: {
      usage: '128Mi',
      request: '256Mi',
      limit: '512Mi', 
      usageValue: 128,
      requestValue: 256,
      limitValue: 512,
      requestPercentage: 50,
      limitPercentage: 25
    },
    labels: null // This will cause the crash
  },
  {
    name: 'test-pod-2',
    namespace: 'default',
    containerName: 'test-container-2',
    cpu: {
      usage: '200m',
      request: '300m',
      limit: '600m', 
      usageValue: 200,
      requestValue: 300,
      limitValue: 600,
      requestPercentage: 66.7,
      limitPercentage: 33.3
    },
    memory: {
      usage: '256Mi',
      request: '512Mi',
      limit: '1Gi',
      usageValue: 256,
      requestValue: 512,
      limitValue: 1024,
      requestPercentage: 50,
      limitPercentage: 25
    },
    labels: undefined // This will also cause the crash
  },
  {
    name: 'test-pod-3',
    namespace: 'default',
    containerName: 'test-container-3',
    cpu: {
      usage: '150m',
      request: '250m',
      limit: '500m',
      usageValue: 150,
      requestValue: 250,
      limitValue: 500,
      requestPercentage: 60,
      limitPercentage: 30
    },
    memory: {
      usage: '200Mi',
      request: '400Mi',
      limit: '800Mi',
      usageValue: 200,
      requestValue: 400,
      limitValue: 800,
      requestPercentage: 50,
      limitPercentage: 25
    },
    labels: { app: 'test-app', version: 'v1.0' } // This one has proper labels
  }
];

const mockSummary = {
  totalPods: 3,
  averageCpuUsage: 50,
  averageMemoryUsage: 50,
  highCpuPods: 0,
  highMemoryPods: 0,
  lowCpuPods: 0,
  lowMemoryPods: 0,
  generatedAt: new Date().toISOString()
};

describe('App Search Functionality', () => {
  beforeEach(() => {
    fetchNamespaces.mockResolvedValue(mockNamespaces);
    fetchPodMetrics.mockResolvedValue(mockPodsWithNullLabels);
    fetchPodSummary.mockResolvedValue(mockSummary);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('should handle search without crashing when pods have null labels', async () => {
    render(<App />);

    // Wait for the component to load data
    await waitFor(() => {
      expect(screen.getByText('Kubernetes Pod Metrics Dashboard')).toBeInTheDocument();
    });

    // Wait for pods to load
    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });

    // Find the search input
    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...');
    expect(searchInput).toBeInTheDocument();

    // Type in the search box - this should not crash the app
    await userEvent.type(searchInput, 'test');

    // The app should still be functional and display results
    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
      expect(screen.getByText('test-pod-2')).toBeInTheDocument();
      expect(screen.getByText('test-pod-3')).toBeInTheDocument();
    });

    // Search should still work for pods with proper labels
    await userEvent.clear(searchInput);
    await userEvent.type(searchInput, 'test-app');

    await waitFor(() => {
      // Only test-pod-3 should be visible (it has labels with 'test-app')
      expect(screen.getByText('test-pod-3')).toBeInTheDocument();
      expect(screen.queryByText('test-pod-1')).not.toBeInTheDocument();
      expect(screen.queryByText('test-pod-2')).not.toBeInTheDocument();
    });
  });

  test('should handle search with undefined labels', async () => {
    const podsWithUndefinedLabels = [
      {
        ...mockPodsWithNullLabels[0],
        labels: undefined
      }
    ];

    fetchPodMetrics.mockResolvedValue(podsWithUndefinedLabels);
    
    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...');
    
    // This should not crash
    await userEvent.type(searchInput, 'test');

    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });
  });

  test('should handle search with non-object labels', async () => {
    const podsWithInvalidLabels = [
      {
        ...mockPodsWithNullLabels[0],
        labels: "invalid-labels-string" // Not an object
      }
    ];

    fetchPodMetrics.mockResolvedValue(podsWithInvalidLabels);
    
    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...');
    
    // This should not crash
    await userEvent.type(searchInput, 'test');

    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });
  });

  test('should clear search results when clear button is clicked', async () => {
    render(<App />);

    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...');
    await userEvent.type(searchInput, 'test-pod-3');

    await waitFor(() => {
      expect(screen.getByText('test-pod-3')).toBeInTheDocument();
      expect(screen.queryByText('test-pod-1')).not.toBeInTheDocument();
    });

    // Click the clear button
    const clearButton = screen.getByRole('button', { name: /clear/i });
    await userEvent.click(clearButton);

    // All pods should be visible again
    await waitFor(() => {
      expect(screen.getByText('test-pod-1')).toBeInTheDocument();
      expect(screen.getByText('test-pod-2')).toBeInTheDocument();
      expect(screen.getByText('test-pod-3')).toBeInTheDocument();
    });
  });
});
