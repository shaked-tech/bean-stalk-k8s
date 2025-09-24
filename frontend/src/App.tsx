import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  CircularProgress,
  Alert,
  AppBar,
  Toolbar,
  IconButton,
  Tooltip,
  Card,
  CardContent
} from '@mui/material';
import { Refresh as RefreshIcon } from '@mui/icons-material';
import PodMetricsTable from './components/PodMetricsTable';
import NamespaceFilter from './components/NamespaceFilter';
import { fetchNamespaces, fetchPodMetrics, fetchPodSummary, PodMetrics, PodSummaryResponse } from './services/api';

function App() {
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [selectedNamespace, setSelectedNamespace] = useState<string>('');
  const [pods, setPods] = useState<PodMetrics[]>([]);
  const [summary, setSummary] = useState<PodSummaryResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<string>('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [refreshing, setRefreshing] = useState<boolean>(false);
  const [activeFilter, setActiveFilter] = useState<'high-cpu' | 'low-cpu' | 'high-memory' | 'low-memory' | null>(null);

  // Fetch namespaces on component mount
  useEffect(() => {
    const loadNamespaces = async () => {
      try {
        const namespacesData = await fetchNamespaces();
        setNamespaces(namespacesData);
        setError(null);
      } catch (err) {
        setError('Failed to fetch namespaces');
        console.error('Error fetching namespaces:', err);
      }
    };
    
    loadNamespaces();
  }, []);

  // Fetch pod metrics and summary when namespace changes or on initial load
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        const [podsData, summaryData] = await Promise.all([
          fetchPodMetrics(selectedNamespace || undefined),
          fetchPodSummary(selectedNamespace || undefined)
        ]);
        
        // Ensure we always set an array, even if API returns null/undefined
        setPods(podsData || []);
        setSummary(summaryData);
        setError(null);
      } catch (err) {
        setError('Failed to fetch pod data');
        console.error('Error fetching pod data:', err);
        // Set empty array on error to maintain UI stability
        setPods([]);
        setSummary(null);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [selectedNamespace]);

  const handleNamespaceChange = (namespace: string) => {
    setSelectedNamespace(namespace);
  };

  const handleSortChange = (property: string) => {
    if (sortBy === property) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(property);
      // Default to descending for numerical columns to show highest values first
      const numericalColumns = [
        'cpuUsage', 'cpuRequest', 'cpuLimit', 'cpuRequestPercentage', 'cpuLimitPercentage',
        'memoryUsage', 'memoryRequest', 'memoryLimit', 'memoryRequestPercentage', 'memoryLimitPercentage'
      ];
      setSortDirection(numericalColumns.includes(property) ? 'desc' : 'asc');
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      // Refresh namespaces, pod metrics, and summary
      const [namespacesData, podsData, summaryData] = await Promise.all([
        fetchNamespaces(),
        fetchPodMetrics(selectedNamespace || undefined),
        fetchPodSummary(selectedNamespace || undefined)
      ]);
      
      setNamespaces(namespacesData);
      setPods(podsData);
      setSummary(summaryData);
      setError(null);
    } catch (err) {
      setError('Failed to refresh data');
      console.error('Error refreshing data:', err);
    } finally {
      setRefreshing(false);
    }
  };

  const handleFilterClick = (filterType: 'high-cpu' | 'low-cpu' | 'high-memory' | 'low-memory') => {
    // Toggle filter - if same filter is clicked, clear it
    if (activeFilter === filterType) {
      setActiveFilter(null);
    } else {
      setActiveFilter(filterType);
    }
  };

  // Filter pods based on active filter
  const getFilteredPods = () => {
    if (!activeFilter) {
      return pods;
    }

    switch (activeFilter) {
      case 'high-cpu':
        return pods.filter(pod => pod.cpu.requestPercentage > 80);
      case 'low-cpu':
        return pods.filter(pod => pod.cpu.requestPercentage < 40 && pod.cpu.requestPercentage > 0);
      case 'high-memory':
        return pods.filter(pod => pod.memory.requestPercentage > 80);
      case 'low-memory':
        return pods.filter(pod => pod.memory.requestPercentage < 40 && pod.memory.requestPercentage > 0);
      default:
        return pods;
    }
  };

  const filteredPods = getFilteredPods();

  // Use summary data from API (with fallback to manual calculations)
  const totalPods = summary?.totalPods ?? pods.length;
  const averageCpuUsage = summary?.averageCpuUsage ?? (
    pods.length > 0 
      ? pods.reduce((sum, pod) => sum + pod.cpu.requestPercentage, 0) / pods.length 
      : 0
  );
  const averageMemoryUsage = summary?.averageMemoryUsage ?? (
    pods.length > 0 
      ? pods.reduce((sum, pod) => sum + pod.memory.requestPercentage, 0) / pods.length 
      : 0
  );
  const highCpuPods = summary?.highCpuPods ?? pods.filter(pod => pod.cpu.requestPercentage > 80).length;
  const highMemoryPods = summary?.highMemoryPods ?? pods.filter(pod => pod.memory.requestPercentage > 80).length;
  const lowCpuPods = summary?.lowCpuPods ?? pods.filter(pod => pod.cpu.requestPercentage < 40 && pod.cpu.requestPercentage > 0).length;
  const lowMemoryPods = summary?.lowMemoryPods ?? pods.filter(pod => pod.memory.requestPercentage < 40 && pod.memory.requestPercentage > 0).length;

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static" sx={{ mb: 3 }}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Kubernetes Pod Metrics Dashboard
          </Typography>
          <Tooltip title="Refresh Data">
            <IconButton 
              color="inherit" 
              onClick={handleRefresh}
              disabled={refreshing}
            >
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      <Container maxWidth={false} sx={{ width: '95%', maxWidth: 'none' }}>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {/* Summary Cards */}
        <Box sx={{ 
          display: 'flex', 
          gap: 2, 
          mb: 3,
          flexWrap: 'wrap'
        }}>
          <Card sx={{ minWidth: 200, flex: '1 1 200px' }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Pods
              </Typography>
              <Typography variant="h4">
                {totalPods}
              </Typography>
            </CardContent>
          </Card>
          <Card sx={{ minWidth: 200, flex: '1 1 200px' }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Avg CPU Usage
              </Typography>
              <Typography variant="h4">
                {Math.round(averageCpuUsage)}%
              </Typography>
            </CardContent>
          </Card>
          <Card sx={{ minWidth: 200, flex: '1 1 200px' }}>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Avg Memory Usage
              </Typography>
              <Typography variant="h4">
                {Math.round(averageMemoryUsage)}%
              </Typography>
            </CardContent>
          </Card>
          <Card 
            sx={{ 
              minWidth: 200, 
              flex: '1 1 200px',
              cursor: 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 3
              },
              ...(activeFilter === 'high-cpu' && {
                border: '2px solid',
                borderColor: 'error.main',
                backgroundColor: 'error.light',
                '& .MuiTypography-root': {
                  color: 'error.contrastText'
                }
              })
            }}
            onClick={() => handleFilterClick('high-cpu')}
          >
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                High CPU Pods
              </Typography>
              <Typography variant="h4" color={highCpuPods > 0 ? "error" : "inherit"}>
                {highCpuPods}
              </Typography>
            </CardContent>
          </Card>
          <Card 
            sx={{ 
              minWidth: 200, 
              flex: '1 1 200px',
              cursor: 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 3
              },
              ...(activeFilter === 'high-memory' && {
                border: '2px solid',
                borderColor: 'error.main',
                backgroundColor: 'error.light',
                '& .MuiTypography-root': {
                  color: 'error.contrastText'
                }
              })
            }}
            onClick={() => handleFilterClick('high-memory')}
          >
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                High Memory Pods
              </Typography>
              <Typography variant="h4" color={highMemoryPods > 0 ? "error" : "inherit"}>
                {highMemoryPods}
              </Typography>
            </CardContent>
          </Card>
          <Card 
            sx={{ 
              minWidth: 200, 
              flex: '1 1 200px',
              cursor: 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 3
              },
              ...(activeFilter === 'low-cpu' && {
                border: '2px solid',
                borderColor: 'info.main',
                backgroundColor: 'info.light',
                '& .MuiTypography-root': {
                  color: 'info.contrastText'
                }
              })
            }}
            onClick={() => handleFilterClick('low-cpu')}
          >
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Low CPU Pods
              </Typography>
              <Typography variant="h4" color={lowCpuPods > 0 ? "info.main" : "inherit"}>
                {lowCpuPods}
              </Typography>
            </CardContent>
          </Card>
          <Card 
            sx={{ 
              minWidth: 200, 
              flex: '1 1 200px',
              cursor: 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                transform: 'translateY(-2px)',
                boxShadow: 3
              },
              ...(activeFilter === 'low-memory' && {
                border: '2px solid',
                borderColor: 'info.main',
                backgroundColor: 'info.light',
                '& .MuiTypography-root': {
                  color: 'info.contrastText'
                }
              })
            }}
            onClick={() => handleFilterClick('low-memory')}
          >
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Low Memory Pods
              </Typography>
              <Typography variant="h4" color={lowMemoryPods > 0 ? "info.main" : "inherit"}>
                {lowMemoryPods}
              </Typography>
            </CardContent>
          </Card>
        </Box>

        <Paper sx={{ p: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 2, flexWrap: 'wrap' }}>
            <Typography variant="h5" component="h2">
              Pod Metrics
            </Typography>
            <NamespaceFilter
              namespaces={namespaces}
              selectedNamespace={selectedNamespace}
              onNamespaceChange={handleNamespaceChange}
              loading={loading || refreshing}
            />
            {activeFilter && (
              <Box sx={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: 1,
                px: 2,
                py: 1,
                borderRadius: 1,
                backgroundColor: activeFilter.includes('high') ? 'error.light' : 'info.light',
                color: activeFilter.includes('high') ? 'error.contrastText' : 'info.contrastText'
              }}>
                <Typography variant="body2" fontWeight="bold">
                  Filter: {activeFilter.replace('-', ' ').toUpperCase()}
                </Typography>
                <Typography variant="body2">
                  ({filteredPods.length} of {pods.length} pods)
                </Typography>
              </Box>
            )}
            {(loading || refreshing) && <CircularProgress size={24} />}
          </Box>

          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress size={40} />
            </Box>
          ) : (
            <>
              {pods.length === 0 ? (
                <Typography variant="body1" sx={{ textAlign: 'center', py: 4 }}>
                  {selectedNamespace 
                    ? `No pods found in namespace "${selectedNamespace}"`
                    : 'No pods found'
                  }
                </Typography>
              ) : filteredPods.length === 0 && activeFilter ? (
                <Typography variant="body1" sx={{ textAlign: 'center', py: 4 }}>
                  No pods match the current filter ({activeFilter.replace('-', ' ')})
                </Typography>
              ) : (
                <PodMetricsTable
                  pods={filteredPods}
                  sortBy={sortBy}
                  sortDirection={sortDirection}
                  onSortChange={handleSortChange}
                />
              )}
            </>
          )}
        </Paper>
      </Container>
    </Box>
  );
}

export default App;
