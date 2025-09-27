import React from 'react';
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
  useTheme
} from '@mui/material';
import { PodMetrics } from '../services/api';

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

  return (
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
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default PodMetricsTable;
