import React, { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
  Typography,
  Box,
  useMediaQuery,
  useTheme,
  Button,
  CircularProgress
} from '@mui/material';
import {
  CheckCircle as OptimizedIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  Psychology as RecommendIcon
} from '@mui/icons-material';
import { PodMetrics, PodResourceRecommendation, fetchPodRecommendations } from '../services/api';
import RecommendationModal from './RecommendationModal';

interface PodMetricsTableProps {
  pods: PodMetrics[];
  sortBy: string;
  sortDirection: 'asc' | 'desc';
  onSortChange: (property: string) => void;
}

const PodMetricsTable: React.FC<PodMetricsTableProps> = ({
  pods,
  sortBy,
  sortDirection,
  onSortChange
}) => {
  const theme = useTheme();

  // Responsive breakpoints for column visibility
  const isExtraSmall = useMediaQuery(theme.breakpoints.down('sm')); // < 600px
  const isSmall = useMediaQuery(theme.breakpoints.down('md')); // < 900px
  const isMedium = useMediaQuery(theme.breakpoints.down('lg')); // < 1200px
  const isLarge = useMediaQuery(theme.breakpoints.down('xl')); // < 1536px

  // Recommendation state
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedPod, setSelectedPod] = useState<PodMetrics | null>(null);
  const [recommendation, setRecommendation] = useState<PodResourceRecommendation | undefined>(undefined);
  const [loadingRecommendation, setLoadingRecommendation] = useState(false);
  const [recommendationError, setRecommendationError] = useState<string>('');
  const [loadingPods, setLoadingPods] = useState<Set<string>>(new Set());

  const handleSort = (property: string) => {
    onSortChange(property);
  };

  // Define column visibility based on screen size/zoom level
  const getVisibleColumns = () => {
    if (isExtraSmall) {
      // Very small screens: only essential columns
      return {
        name: true,
        containerName: false,
        namespace: true,
        cpuUsage: false,
        cpuRequest: false,
        cpuLimit: false,
        cpuRequestPercentage: true,
        cpuLimitPercentage: false,
        memoryUsage: false,
        memoryRequest: false,
        memoryLimit: false,
        memoryRequestPercentage: true,
        memoryLimitPercentage: false
      };
    } else if (isSmall) {
      // Small screens: essential + some important columns
      return {
        name: true,
        containerName: false,
        namespace: true,
        cpuUsage: false,
        cpuRequest: false,
        cpuLimit: false,
        cpuRequestPercentage: true,
        cpuLimitPercentage: false,
        memoryUsage: false,
        memoryRequest: false,
        memoryLimit: false,
        memoryRequestPercentage: true,
        memoryLimitPercentage: false
      };
    } else if (isMedium) {
      // Medium screens: hide limit percentages and some raw values
      return {
        name: true,
        containerName: true,
        namespace: true,
        cpuUsage: true,
        cpuRequest: false,
        cpuLimit: false,
        cpuRequestPercentage: true,
        cpuLimitPercentage: false,
        memoryUsage: true,
        memoryRequest: false,
        memoryLimit: false,
        memoryRequestPercentage: true,
        memoryLimitPercentage: false
      };
    } else if (isLarge) {
      // Large screens: hide limit percentages
      return {
        name: true,
        containerName: true,
        namespace: true,
        cpuUsage: true,
        cpuRequest: true,
        cpuLimit: true,
        cpuRequestPercentage: true,
        cpuLimitPercentage: false,
        memoryUsage: true,
        memoryRequest: true,
        memoryLimit: true,
        memoryRequestPercentage: true,
        memoryLimitPercentage: false
      };
    } else {
      // Extra large screens: show all columns
      return {
        name: true,
        containerName: true,
        namespace: true,
        cpuUsage: true,
        cpuRequest: true,
        cpuLimit: true,
        cpuRequestPercentage: true,
        cpuLimitPercentage: true,
        memoryUsage: true,
        memoryRequest: true,
        memoryLimit: true,
        memoryRequestPercentage: true,
        memoryLimitPercentage: true
      };
    }
  };

  const visibleColumns = getVisibleColumns();

  const sortedPods = [...pods].sort((a, b) => {
    let comparison = 0;

    switch (sortBy) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'namespace':
        comparison = a.namespace.localeCompare(b.namespace);
        break;
      case 'containerName':
        comparison = a.containerName.localeCompare(b.containerName);
        break;
      case 'cpuUsage':
        comparison = a.cpu.usageValue - b.cpu.usageValue;
        break;
      case 'cpuRequest':
        comparison = a.cpu.requestValue - b.cpu.requestValue;
        break;
      case 'cpuLimit':
        comparison = a.cpu.limitValue - b.cpu.limitValue;
        break;
      case 'cpuRequestPercentage':
        comparison = a.cpu.requestPercentage - b.cpu.requestPercentage;
        break;
      case 'cpuLimitPercentage':
        comparison = a.cpu.limitPercentage - b.cpu.limitPercentage;
        break;
      case 'memoryUsage':
        comparison = a.memory.usageValue - b.memory.usageValue;
        break;
      case 'memoryRequest':
        comparison = a.memory.requestValue - b.memory.requestValue;
        break;
      case 'memoryLimit':
        comparison = a.memory.limitValue - b.memory.limitValue;
        break;
      case 'memoryRequestPercentage':
        comparison = a.memory.requestPercentage - b.memory.requestPercentage;
        break;
      case 'memoryLimitPercentage':
        comparison = a.memory.limitPercentage - b.memory.limitPercentage;
        break;
      default:
        comparison = 0;
    }

    return sortDirection === 'asc' ? comparison : -comparison;
  });

  const renderProgressBar = (value: number, color: string, hasValidTarget: boolean = true) => {
    // If the target (request/limit) is missing or invalid, or value is NaN, show grey bar
    if (!hasValidTarget || isNaN(value)) {
      return (
        <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
          <Box sx={{ width: '100%', mr: 1 }}>
            <LinearProgress
              variant="determinate"
              value={0}
              sx={{
                height: 10,
                borderRadius: 5,
                backgroundColor: '#e0e0e0',
                '& .MuiLinearProgress-bar': {
                  backgroundColor: '#bdbdbd'
                }
              }}
            />
          </Box>
          <Box sx={{ minWidth: 35 }}>
            <Typography variant="body2" color="text.secondary">
              -
            </Typography>
          </Box>
        </Box>
      );
    }

    return (
      <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
        <Box sx={{ width: '100%', mr: 1 }}>
          <LinearProgress
            variant="determinate"
            value={Math.min(value, 100)}
            color={color as "primary" | "secondary" | "error" | "info" | "success" | "warning" | undefined}
            sx={{ height: 10, borderRadius: 5 }}
          />
        </Box>
        <Box sx={{ minWidth: 35 }}>
          <Typography variant="body2" color="text.secondary">
            {`${Math.round(value)}%`}
          </Typography>
        </Box>
      </Box>
    );
  };

  const renderSortArrow = (property: string) => {
    if (sortBy !== property) return null;
    return sortDirection === 'asc' ? ' ↑' : ' ↓';
  };

  // Helper functions for recommendations
  const getPodKey = (pod: PodMetrics): string => {
    return `${pod.namespace}/${pod.name}/${pod.containerName}`;
  };

  const getPodOptimizationStatus = (pod: PodMetrics) => {
    const cpuUtilization = pod.cpu.requestPercentage;
    const memoryUtilization = pod.memory.requestPercentage;
    const hasOptimalCpuUsage = cpuUtilization >= 60 && cpuUtilization <= 75;
    const hasOptimalMemoryUsage = memoryUtilization >= 60 && memoryUtilization <= 75;
    const hasCpuLimit = pod.cpu.limitValue > 0;
    const memoryMisaligned = Math.abs(pod.memory.requestValue - pod.memory.limitValue) > pod.memory.requestValue * 0.01;

    // Check utilization FIRST (highest priority) - ONLY based on actual utilization
    const isOverUtilized = cpuUtilization > 75 || memoryUtilization > 75;
    const isUnderUtilized = cpuUtilization < 60 || memoryUtilization < 60;

    // Over-utilized pods (red) - ONLY when utilization is actually high
    if (isOverUtilized) {
      return { status: 'high_risk', color: 'error' as const, text: '🚨 Over Utilized', icon: <ErrorIcon /> };
    }

    // Optimized pods (green) - good utilization, no CPU limits, memory aligned
    if (hasOptimalCpuUsage && hasOptimalMemoryUsage && !hasCpuLimit && !memoryMisaligned) {
      return { status: 'optimized', color: 'success' as const, text: '✅ Optimized', icon: <OptimizedIcon /> };
    }

    // Under-utilized pods (blue) - low utilization
    if (isUnderUtilized) {
      return { status: 'low_risk', color: 'info' as const, text: '📉 Under-utilized', icon: <OptimizedIcon /> };
    }

    // Minor issues (yellow) - has CPU limit or memory misalignment but utilization is OK
    if (hasCpuLimit || memoryMisaligned) {
      return { status: 'minor_issues', color: 'warning' as const, text: '⚠️ Minor Issues', icon: <WarningIcon /> };
    }

    return { status: 'get_recommendation', color: 'primary' as const, text: 'Get Recommendation', icon: <RecommendIcon /> };
  };

  const handleRecommendationClick = async (pod: PodMetrics) => {
    const podKey = getPodKey(pod);
    setSelectedPod(pod);
    setModalOpen(true);
    setLoadingRecommendation(true);
    setRecommendationError('');
    setRecommendation(undefined);

    // Add to loading set
    setLoadingPods(prev => new Set(prev).add(podKey));

    try {
      const recommendationsResponse = await fetchPodRecommendations(pod.namespace);

      if (recommendationsResponse) {
        // Find the specific recommendation for this pod
        const podRecommendation = recommendationsResponse.recommendations.find(r =>
          r.podName === pod.name &&
          r.namespace === pod.namespace &&
          r.containerName === pod.containerName
        );

        setRecommendation(podRecommendation);

        if (!podRecommendation) {
          setRecommendationError('No recommendation found for this pod. This may indicate insufficient historical data.');
        }
      } else {
        setRecommendationError('Failed to fetch recommendations. Please try again.');
      }
    } catch (error) {
      console.error('Error fetching recommendation:', error);
      setRecommendationError('Failed to fetch recommendations. Please try again.');
    } finally {
      setLoadingRecommendation(false);
      // Remove from loading set
      setLoadingPods(prev => {
        const newSet = new Set(prev);
        newSet.delete(podKey);
        return newSet;
      });
    }
  };

  const handleModalClose = () => {
    setModalOpen(false);
    setSelectedPod(null);
    setRecommendation(undefined);
    setRecommendationError('');
  };

  return (
    <>
    <TableContainer component={Paper}>
      <Table stickyHeader aria-label="sticky table">
        <TableHead>
          <TableRow>
            {visibleColumns.name && (
              <TableCell
                onClick={() => handleSort('name')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 150 }}
              >
                Pod Name{renderSortArrow('name')}
              </TableCell>
            )}
            {visibleColumns.containerName && (
              <TableCell
                onClick={() => handleSort('containerName')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 120 }}
              >
                Container{renderSortArrow('containerName')}
              </TableCell>
            )}
            {visibleColumns.namespace && (
              <TableCell
                onClick={() => handleSort('namespace')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 100 }}
              >
                Namespace{renderSortArrow('namespace')}
              </TableCell>
            )}
            {visibleColumns.cpuUsage && (
              <TableCell
                onClick={() => handleSort('cpuUsage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 100 }}
              >
                CPU Usage{renderSortArrow('cpuUsage')}
              </TableCell>
            )}
            {visibleColumns.cpuRequest && (
              <TableCell
                onClick={() => handleSort('cpuRequest')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 110 }}
              >
                CPU Request{renderSortArrow('cpuRequest')}
              </TableCell>
            )}
            {visibleColumns.cpuLimit && (
              <TableCell
                onClick={() => handleSort('cpuLimit')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 100 }}
              >
                CPU Limit{renderSortArrow('cpuLimit')}
              </TableCell>
            )}
            {visibleColumns.cpuRequestPercentage && (
              <TableCell
                onClick={() => handleSort('cpuRequestPercentage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 140 }}
              >
                CPU Request %{renderSortArrow('cpuRequestPercentage')}
              </TableCell>
            )}
            {visibleColumns.cpuLimitPercentage && (
              <TableCell
                onClick={() => handleSort('cpuLimitPercentage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 130 }}
              >
                CPU Limit %{renderSortArrow('cpuLimitPercentage')}
              </TableCell>
            )}
            {visibleColumns.memoryUsage && (
              <TableCell
                onClick={() => handleSort('memoryUsage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 120 }}
              >
                Memory Usage{renderSortArrow('memoryUsage')}
              </TableCell>
            )}
            {visibleColumns.memoryRequest && (
              <TableCell
                onClick={() => handleSort('memoryRequest')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 130 }}
              >
                Memory Request{renderSortArrow('memoryRequest')}
              </TableCell>
            )}
            {visibleColumns.memoryLimit && (
              <TableCell
                onClick={() => handleSort('memoryLimit')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 120 }}
              >
                Memory Limit{renderSortArrow('memoryLimit')}
              </TableCell>
            )}
            {visibleColumns.memoryRequestPercentage && (
              <TableCell
                onClick={() => handleSort('memoryRequestPercentage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 160 }}
              >
                Memory Request %{renderSortArrow('memoryRequestPercentage')}
              </TableCell>
            )}
            {visibleColumns.memoryLimitPercentage && (
              <TableCell
                onClick={() => handleSort('memoryLimitPercentage')}
                sx={{ cursor: 'pointer', fontWeight: 'bold', minWidth: 150 }}
              >
                Memory Limit %{renderSortArrow('memoryLimitPercentage')}
              </TableCell>
            )}
            {/* Recommendation Actions Column */}
            <TableCell
              sx={{ fontWeight: 'bold', minWidth: 150, textAlign: 'center' }}
            >
              Actions
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedPods.map((pod, index) => (
            <TableRow key={`${pod.namespace}-${pod.name}-${pod.containerName}-${index}`} hover>
              {visibleColumns.name && (
                <TableCell sx={{ minWidth: 150 }}>
                  <Typography noWrap title={pod.name}>
                    {pod.name}
                  </Typography>
                </TableCell>
              )}
              {visibleColumns.containerName && (
                <TableCell sx={{ minWidth: 120 }}>
                  <Typography noWrap title={pod.containerName}>
                    {pod.containerName}
                  </Typography>
                </TableCell>
              )}
              {visibleColumns.namespace && (
                <TableCell sx={{ minWidth: 100 }}>
                  <Typography noWrap title={pod.namespace}>
                    {pod.namespace}
                  </Typography>
                </TableCell>
              )}
              {visibleColumns.cpuUsage && (
                <TableCell sx={{ minWidth: 100 }}>{pod.cpu.usage}</TableCell>
              )}
              {visibleColumns.cpuRequest && (
                <TableCell sx={{ minWidth: 110 }}>
                  {pod.cpu.requestValue > 0 ? pod.cpu.request : '-'}
                </TableCell>
              )}
              {visibleColumns.cpuLimit && (
                <TableCell sx={{ minWidth: 100 }}>
                  {pod.cpu.limitValue > 0 ? pod.cpu.limit : '-'}
                </TableCell>
              )}
              {visibleColumns.cpuRequestPercentage && (
                <TableCell sx={{ minWidth: 140 }}>
                  {renderProgressBar(pod.cpu.requestPercentage, pod.cpu.requestPercentage > 80 ? 'error' : 'primary', pod.cpu.requestValue > 0)}
                </TableCell>
              )}
              {visibleColumns.cpuLimitPercentage && (
                <TableCell sx={{ minWidth: 130 }}>
                  {renderProgressBar(pod.cpu.limitPercentage, pod.cpu.limitPercentage > 80 ? 'error' : 'info', pod.cpu.limitValue > 0)}
                </TableCell>
              )}
              {visibleColumns.memoryUsage && (
                <TableCell sx={{ minWidth: 120 }}>{pod.memory.usage}</TableCell>
              )}
              {visibleColumns.memoryRequest && (
                <TableCell sx={{ minWidth: 130 }}>
                  {pod.memory.requestValue > 0 ? pod.memory.request : '-'}
                </TableCell>
              )}
              {visibleColumns.memoryLimit && (
                <TableCell sx={{ minWidth: 120 }}>
                  {pod.memory.limitValue > 0 ? pod.memory.limit : '-'}
                </TableCell>
              )}
              {visibleColumns.memoryRequestPercentage && (
                <TableCell sx={{ minWidth: 160 }}>
                  {renderProgressBar(pod.memory.requestPercentage, pod.memory.requestPercentage > 80 ? 'error' : 'primary', pod.memory.requestValue > 0)}
                </TableCell>
              )}
              {visibleColumns.memoryLimitPercentage && (
                <TableCell sx={{ minWidth: 150 }}>
                  {renderProgressBar(pod.memory.limitPercentage, pod.memory.limitPercentage > 80 ? 'error' : 'info', pod.memory.limitValue > 0)}
                </TableCell>
              )}
              {/* Recommendation Actions Column */}
              <TableCell sx={{ minWidth: 150, textAlign: 'center' }}>
                {(() => {
                  const podKey = getPodKey(pod);
                  const status = getPodOptimizationStatus(pod);
                  const isLoading = loadingPods.has(podKey);

                  return (
                    <Button
                      size="small"
                      color={status.color}
                      variant={status.status === 'optimized' ? 'outlined' : 'contained'}
                      onClick={() => handleRecommendationClick(pod)}
                      disabled={isLoading}
                      startIcon={isLoading ? <CircularProgress size={16} /> : status.icon}
                      sx={{
                        minWidth: 'auto',
                        whiteSpace: 'nowrap',
                        fontSize: '0.75rem'
                      }}
                    >
                      {isLoading ? 'Loading...' : status.text}
                    </Button>
                  );
                })()}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>

    {/* Recommendation Modal */}
    {selectedPod && (
      <RecommendationModal
        open={modalOpen}
        onClose={handleModalClose}
        podName={selectedPod.name}
        namespace={selectedPod.namespace}
        containerName={selectedPod.containerName}
        recommendation={recommendation}
        loading={loadingRecommendation}
        error={recommendationError}
      />
    )}
    </>
  );
};

export default PodMetricsTable;
