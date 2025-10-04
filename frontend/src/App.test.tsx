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
  jest.setTimeout(15000); // Increase timeout for this test suite

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
    }, { timeout: 8000 });

    // Find the search input (don't wait for pods specifically)
    const searchInput = screen.getByPlaceholderText('Search pods, containers, namespaces...') as HTMLInputElement;
    expect(searchInput).toBeInTheDocument();

    // Type in the search box - this should not crash the app
    await userEvent.type(searchInput, 'test');

    // Verify search input has the typed value (simpler check)
    expect(searchInput.value).toBe('test');

    // Clear and try another search
    await userEvent.clear(searchInput);
    await userEvent.type(searchInput, 'app');

    // Verify search still works
    expect(searchInput.value).toBe('app');
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

});
