import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  Chip,
  Paper,
  IconButton,
  Divider,
  Tabs,
  Tab,
  CircularProgress,
  Alert,
  Tooltip,
  Card,
  CardContent
} from '@mui/material';
import {
  Close as CloseIcon,
  ContentCopy as CopyIcon,
  CheckCircle as OptimizedIcon,
  Warning as WarningIcon,
  Error as ErrorIcon
} from '@mui/icons-material';
import { PodResourceRecommendation } from '../services/api';

interface RecommendationModalProps {
  open: boolean;
  onClose: () => void;
  podName: string;
  namespace: string;
  containerName: string;
  recommendation?: PodResourceRecommendation;
  loading?: boolean;
  error?: string;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

const TabPanel: React.FC<TabPanelProps> = ({ children, value, index, ...other }) => {
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`recommendation-tabpanel-${index}`}
      aria-labelledby={`recommendation-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 0 }}>{children}</Box>}
    </div>
  );
};

const RecommendationModal: React.FC<RecommendationModalProps> = ({
  open,
  onClose,
  podName,
  namespace,
  containerName,
  recommendation,
  loading,
  error
}) => {
  const [tabValue, setTabValue] = useState(0);
  const [copySuccess, setCopySuccess] = useState(false);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const generateYAML = (rec: PodResourceRecommendation): string => {
    const yaml = `# Resource recommendations for ${rec.containerName} container
# Pod: ${rec.podName} (${rec.namespace})
# Generated: ${new Date().toISOString()}
# Analysis: ${rec.basedOnDays} days, Confidence: ${Math.round(rec.confidenceScore)}%

resources:${rec.cpu.recommendedRequest || rec.memory.recommendedRequest ? `
  requests:${rec.cpu.recommendedRequest ? `
    cpu: "${rec.cpu.recommendedRequest}"` : ''}${rec.memory.recommendedRequest ? `
    memory: "${rec.memory.recommendedRequest}"` : ''}` : ''}${rec.memory.recommendedLimit ? `
  limits:
    memory: "${rec.memory.recommendedLimit}"` : ''}${rec.cpu.currentLimit && rec.cpu.resourceChange === 'remove_limit' ? `
    # cpu: removed (limits can cause throttling)` : ''}

# Current configuration:
# resources:
#   requests:
#     cpu: "${rec.cpu.currentRequest}"
#     memory: "${rec.memory.currentRequest}"${rec.cpu.currentLimit !== '0' ? `
#   limits:
#     cpu: "${rec.cpu.currentLimit}"` : ''}${rec.memory.currentLimit !== '0' ? `
#     memory: "${rec.memory.currentLimit}"` : ''}`;

    return yaml;
  };

  const handleCopyYAML = async () => {
    if (!recommendation) return;

    try {
      await navigator.clipboard.writeText(generateYAML(recommendation));
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (err) {
      console.error('Failed to copy YAML:', err);
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'error';
      case 'medium':
        return 'warning';
      case 'low':
        return 'info';
      default:
        return 'default';
    }
  };

  const getPriorityIcon = (priority: string) => {
    switch (priority) {
      case 'high':
        return <ErrorIcon />;
      case 'medium':
        return <WarningIcon />;
      case 'low':
        return <OptimizedIcon />;
      default:
        return <OptimizedIcon />;
    }
  };

  const getStatusText = (rec?: PodResourceRecommendation) => {
    if (!rec) return { text: 'Loading...', color: 'default' as const };

    const isOptimized = rec.reasons.includes('well_optimized') && rec.reasons.length === 1;

    if (isOptimized) {
      return { text: 'Well Optimized', color: 'success' as const };
    }

    const hasMinorIssues = rec.reasons.some(r =>
      r === 'cpu_limit_present' || r === 'memory_misaligned'
    );

    if (hasMinorIssues && rec.priority === 'medium') {
      return { text: 'Minor Issues', color: 'warning' as const };
    }

    return { text: 'Needs Optimization', color: 'error' as const };
  };

  const formatReason = (reason: string): string => {
    const reasonMap: Record<string, string> = {
      'well_optimized': 'Well Optimized',
      'over_utilized': 'Over Utilized',
      'under_utilized': 'Under Utilized',
      'cpu_over_utilized': 'CPU Over Utilized',
      'cpu_under_utilized': 'CPU Under Utilized',
      'memory_over_utilized': 'Memory Over Utilized',
      'memory_under_utilized': 'Memory Under Utilized',
      'cpu_limit_present': 'CPU Limit Present',
      'memory_misaligned': 'Memory Misaligned',
      'missing_requests': 'Missing Resource Requests',
      'missing_limits': 'Missing Resource Limits'
    };

    return reasonMap[reason] || reason.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
  };

  const statusInfo = getStatusText(recommendation);

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { minHeight: '500px' }
      }}
    >
      <DialogTitle sx={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        pb: 1
      }}>
        <Box>
          <Typography variant="h6" component="div">
            Resource Recommendation
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {podName} → {containerName} ({namespace})
          </Typography>
        </Box>
        <IconButton onClick={onClose} size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>

      <DialogContent sx={{ pt: 1 }}>
        {loading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
            <Typography variant="body2" sx={{ ml: 2 }}>
              Analyzing resource usage patterns...
            </Typography>
          </Box>
        )}

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {recommendation && (
          <>
            {/* Status Header */}
            <Card sx={{ mb: 2, bgcolor: `${statusInfo.color}.50` }}>
              <CardContent sx={{ py: 2, '&:last-child': { pb: 2 } }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Chip
                    icon={getPriorityIcon(recommendation.priority)}
                    label={statusInfo.text}
                    color={statusInfo.color}
                    variant="outlined"
                  />
                  <Typography variant="body2">
                    Priority: {recommendation.priority.toUpperCase()}
                  </Typography>
                  <Typography variant="body2">
                    Confidence: {Math.round(recommendation.confidenceScore)}%
                  </Typography>
                  <Typography variant="body2">
                    Risk: {recommendation.riskLevel}
                  </Typography>
                </Box>
              </CardContent>
            </Card>

            {/* Reasons */}
            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" gutterBottom>
                Analysis Results:
              </Typography>
              <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                {recommendation.reasons.map((reason, index) => (
                  <Chip
                    key={index}
                    label={formatReason(reason)}
                    size="small"
                    color={reason === 'well_optimized' ? 'success' : 'default'}
                    variant={reason === 'well_optimized' ? 'filled' : 'outlined'}
                  />
                ))}
              </Box>
            </Box>

            {/* Tabs */}
            <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
              <Tabs value={tabValue} onChange={handleTabChange}>
                <Tab label="YAML Configuration" />
                <Tab label="Current vs Recommended" />
                <Tab label="Analysis Details" />
              </Tabs>
            </Box>

            {/* YAML Tab */}
            <TabPanel value={tabValue} index={0}>
              <Box sx={{ position: 'relative' }}>
                <Paper
                  sx={{
                    p: 2,
                    bgcolor: 'grey.50',
                    fontFamily: 'monospace',
                    position: 'relative',
                    maxHeight: '400px',
                    overflow: 'auto'
                  }}
                >
                  <Tooltip title={copySuccess ? "Copied!" : "Copy YAML"}>
                    <IconButton
                      onClick={handleCopyYAML}
                      sx={{
                        position: 'absolute',
                        top: 8,
                        right: 8,
                        bgcolor: 'background.paper',
                        '&:hover': { bgcolor: 'grey.100' }
                      }}
                      size="small"
                    >
                      <CopyIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <pre style={{
                    margin: 0,
                    fontSize: '0.875rem',
                    whiteSpace: 'pre-wrap',
                    paddingRight: '40px'
                  }}>
                    {generateYAML(recommendation)}
                  </pre>
                </Paper>
              </Box>
            </TabPanel>

            {/* Comparison Tab */}
            <TabPanel value={tabValue} index={1}>
              <Box sx={{ display: 'flex', gap: 2 }}>
                {/* Current Resources */}
                <Paper sx={{ flex: 1, p: 2 }}>
                  <Typography variant="subtitle2" gutterBottom color="text.secondary">
                    Current Resources
                  </Typography>
                  <Divider sx={{ mb: 2 }} />
                  <Typography variant="body2" gutterBottom>
                    <strong>CPU:</strong>
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Request: {recommendation.cpu.currentRequest}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Limit: {recommendation.cpu.currentLimit || 'None'}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Usage: {recommendation.cpu.currentUsage} ({Math.round(recommendation.cpu.currentUtilization)}%)
                  </Typography>

                  <Typography variant="body2" gutterBottom sx={{ mt: 2 }}>
                    <strong>Memory:</strong>
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Request: {recommendation.memory.currentRequest}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Limit: {recommendation.memory.currentLimit || 'None'}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Usage: {recommendation.memory.currentUsage} ({Math.round(recommendation.memory.currentUtilization)}%)
                  </Typography>
                </Paper>

                {/* Recommended Resources */}
                <Paper sx={{ flex: 1, p: 2, bgcolor: 'success.50' }}>
                  <Typography variant="subtitle2" gutterBottom color="success.dark">
                    Recommended Resources
                  </Typography>
                  <Divider sx={{ mb: 2 }} />
                  <Typography variant="body2" gutterBottom>
                    <strong>CPU:</strong>
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Request: {recommendation.cpu.recommendedRequest || 'No change'}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Limit: {recommendation.cpu.recommendedLimit !== undefined ? (recommendation.cpu.recommendedLimit || 'Remove') : 'No change'}
                  </Typography>
                  {recommendation.cpu.percentageChange !== 0 && (
                    <Typography variant="body2" sx={{ ml: 2, mb: 1, color: recommendation.cpu.percentageChange > 0 ? 'warning.main' : 'success.main' }}>
                      Change: {recommendation.cpu.percentageChange > 0 ? '+' : ''}{Math.round(recommendation.cpu.percentageChange)}%
                    </Typography>
                  )}

                  <Typography variant="body2" gutterBottom sx={{ mt: 2 }}>
                    <strong>Memory:</strong>
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Request: {recommendation.memory.recommendedRequest || 'No change'}
                  </Typography>
                  <Typography variant="body2" sx={{ ml: 2, mb: 1 }}>
                    Limit: {recommendation.memory.recommendedLimit || 'No change'}
                  </Typography>
                  {recommendation.memory.percentageChange !== 0 && (
                    <Typography variant="body2" sx={{ ml: 2, mb: 1, color: recommendation.memory.percentageChange > 0 ? 'warning.main' : 'success.main' }}>
                      Change: {recommendation.memory.percentageChange > 0 ? '+' : ''}{Math.round(recommendation.memory.percentageChange)}%
                    </Typography>
                  )}
                </Paper>
              </Box>
            </TabPanel>

            {/* Analysis Details Tab */}
            <TabPanel value={tabValue} index={2}>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Analysis Summary
                  </Typography>
                  <Typography variant="body2">
                    <strong>Data Quality:</strong> {recommendation.dataQuality}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Analysis Period:</strong> {recommendation.basedOnDays} days
                  </Typography>
                  <Typography variant="body2">
                    <strong>Estimated Savings:</strong> {recommendation.estimatedSavings}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Risk Level:</strong> {recommendation.riskLevel}
                  </Typography>
                </Paper>

                <Paper sx={{ p: 2 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Target Utilization
                  </Typography>
                  <Typography variant="body2">
                    Our recommendation engine targets 60-75% resource utilization for optimal efficiency.
                    This range provides good performance while allowing room for traffic spikes.
                  </Typography>
                  <Typography variant="body2" sx={{ mt: 1 }}>
                    <strong>CPU Strategy:</strong> Remove limits to prevent throttling, adjust requests for 70% utilization
                  </Typography>
                  <Typography variant="body2">
                    <strong>Memory Strategy:</strong> Align requests and limits, target 70% utilization based on P95 usage
                  </Typography>
                </Paper>
              </Box>
            </TabPanel>
          </>
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose}>
          Close
        </Button>
        {recommendation && (
          <Button
            variant="contained"
            startIcon={<CopyIcon />}
            onClick={handleCopyYAML}
            disabled={copySuccess}
          >
            {copySuccess ? 'Copied!' : 'Copy YAML'}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default RecommendationModal;
