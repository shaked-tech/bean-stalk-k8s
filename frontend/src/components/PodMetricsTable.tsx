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
  Tooltip
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
  const handleSort = (property: string) => {
    onSortChange(property);
  };

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

  const getSortDirection = (property: string) => {
    return sortBy === property ? sortDirection : 'asc';
  };

  const renderSortArrow = (property: string) => {
    if (sortBy !== property) return null;
    return sortDirection === 'asc' ? ' ↑' : ' ↓';
  };

  return (
    <TableContainer component={Paper} sx={{ maxHeight: 600 }}>
      <Table stickyHeader aria-label="sticky table">
        <TableHead>
          <TableRow>
            <TableCell 
              onClick={() => handleSort('name')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Pod Name{renderSortArrow('name')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('containerName')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Container{renderSortArrow('containerName')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('namespace')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Namespace{renderSortArrow('namespace')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('cpuUsage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              CPU Usage{renderSortArrow('cpuUsage')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('cpuRequest')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              CPU Request{renderSortArrow('cpuRequest')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('cpuLimit')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              CPU Limit{renderSortArrow('cpuLimit')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('cpuRequestPercentage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              CPU Request %{renderSortArrow('cpuRequestPercentage')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('cpuLimitPercentage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              CPU Limit %{renderSortArrow('cpuLimitPercentage')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('memoryUsage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Memory Usage{renderSortArrow('memoryUsage')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('memoryRequest')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Memory Request{renderSortArrow('memoryRequest')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('memoryLimit')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Memory Limit{renderSortArrow('memoryLimit')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('memoryRequestPercentage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Memory Request %{renderSortArrow('memoryRequestPercentage')}
            </TableCell>
            <TableCell 
              onClick={() => handleSort('memoryLimitPercentage')}
              sx={{ cursor: 'pointer', fontWeight: 'bold' }}
            >
              Memory Limit %{renderSortArrow('memoryLimitPercentage')}
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedPods.map((pod, index) => (
            <TableRow key={`${pod.namespace}-${pod.name}-${pod.containerName}-${index}`} hover>
              <TableCell>
                <Typography>{pod.name}</Typography>
              </TableCell>
              <TableCell>{pod.containerName}</TableCell>
              <TableCell>{pod.namespace}</TableCell>
              <TableCell>{pod.cpu.usage}</TableCell>
              <TableCell>{pod.cpu.requestValue > 0 ? pod.cpu.request : '-'}</TableCell>
              <TableCell>{pod.cpu.limitValue > 0 ? pod.cpu.limit : '-'}</TableCell>
              <TableCell>
                {renderProgressBar(pod.cpu.requestPercentage, pod.cpu.requestPercentage > 80 ? 'error' : 'primary', pod.cpu.requestValue > 0)}
              </TableCell>
              <TableCell>
                {renderProgressBar(pod.cpu.limitPercentage, pod.cpu.limitPercentage > 80 ? 'error' : 'info', pod.cpu.limitValue > 0)}
              </TableCell>
              <TableCell>{pod.memory.usage}</TableCell>
              <TableCell>{pod.memory.requestValue > 0 ? pod.memory.request : '-'}</TableCell>
              <TableCell>{pod.memory.limitValue > 0 ? pod.memory.limit : '-'}</TableCell>
              <TableCell>
                {renderProgressBar(pod.memory.requestPercentage, pod.memory.requestPercentage > 80 ? 'error' : 'primary', pod.memory.requestValue > 0)}
              </TableCell>
              <TableCell>
                {renderProgressBar(pod.memory.limitPercentage, pod.memory.limitPercentage > 80 ? 'error' : 'info', pod.memory.limitValue > 0)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default PodMetricsTable;
