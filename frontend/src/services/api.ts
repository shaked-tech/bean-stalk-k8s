export interface ResourceMetrics {
  usage: string;
  request: string;
  limit: string;
  usageValue: number;
  requestValue: number;
  limitValue: number;
  requestPercentage: number;
  limitPercentage: number;
}

export interface PodMetrics {
  name: string;
  namespace: string;
  containerName: string;
  cpu: ResourceMetrics;
  memory: ResourceMetrics;
  labels: Record<string, string>;
}

export interface NamespaceList {
  namespaces: string[];
}

export interface PodMetricsList {
  pods: PodMetrics[];
}

export interface PodSummaryResponse {
  totalPods: number;
  averageCpuUsage: number;
  averageMemoryUsage: number;
  highCpuPods: number;      // >80% usage
  highMemoryPods: number;   // >80% usage
  lowCpuPods: number;       // <40% usage
  lowMemoryPods: number;    // <40% usage
  generatedAt: string;
}

export interface ClusterInfo {
  clusterName: string;
}

// Recommendation types
export interface ResourceRecommendation {
  currentRequest: string;
  currentLimit: string;
  currentUsage: string;
  recommendedRequest?: string;
  recommendedLimit?: string;
  currentUtilization: number;
  targetUtilization: number;
  currentRequestValue: number;
  currentLimitValue: number;
  currentUsageValue: number;
  recommendedRequestValue: number;
  recommendedLimitValue?: number;
  resourceChange: string;
  percentageChange: number;
}

export interface PodResourceRecommendation {
  podName: string;
  namespace: string;
  containerName: string;
  cpu: ResourceRecommendation;
  memory: ResourceRecommendation;
  reasons: string[];
  priority: 'high' | 'medium' | 'low';
  confidenceScore: number;
  estimatedSavings: string;
  riskLevel: string;
  basedOnDays: number;
  dataQuality: string;
}

export interface RecommendationsSummary {
  totalPodsAnalyzed: number;
  podsNeedingOptimization: number;
  podsWellOptimized: number;
  totalCpuRequestIncrease: string;
  totalCpuRequestDecrease: string;
  totalMemoryRequestChange: string;
  estimatedCostSavings: string;
  highPriorityRecommendations: number;
  mediumPriorityRecommendations: number;
  lowPriorityRecommendations: number;
  overUtilizedPods: number;
  underUtilizedPods: number;
  podsWithCpuLimits: number;
  memoryMisalignedPods: number;
  podsMissingRequests: number;
}

export interface TimeRange {
  start: string;
  end: string;
}

export interface RecommendationsResponse {
  recommendations: PodResourceRecommendation[];
  summary: RecommendationsSummary;
  generatedAt: string;
  timeRange: TimeRange;
  analysisWindow: string;
  targetUtilization: number;
}

const API_BASE_URL = '/api';

// Enable mock data via environment variable for QA testing
// Set REACT_APP_USE_MOCK_DATA=true to enable mock data
// WARNING: NEVER set this environment variable in production deployments!
const USE_MOCK_DATA = process.env.REACT_APP_USE_MOCK_DATA === 'true';

// Production safeguard: Force disable mock data in production environment
const IS_PRODUCTION = process.env.NODE_ENV === 'production' ||
                     process.env.REACT_APP_ENV === 'production' ||
                     window.location.hostname !== 'localhost';

const SAFE_USE_MOCK_DATA = USE_MOCK_DATA && !IS_PRODUCTION;

// Log warning if mock data is enabled
if (SAFE_USE_MOCK_DATA) {
  console.warn('🔶 MOCK DATA ENABLED - Using mock pod data instead of real metrics');
} else if (USE_MOCK_DATA && IS_PRODUCTION) {
  console.error('🚨 MOCK DATA BLOCKED IN PRODUCTION - Using real API instead');
}

// Mock data for QA testing
const MOCK_NAMESPACES = ['default', 'kube-system', 'monitoring', 'app-prod', 'app-staging'];

const MOCK_PODS: PodMetrics[] = [
  // High CPU pods (>80%)
  {
    name: 'web-server-high-cpu-7d8f9c',
    namespace: 'app-prod',
    containerName: 'nginx',
    cpu: {
      usage: '950m', request: '1000m', limit: '2000m',
      usageValue: 950, requestValue: 1000, limitValue: 2000,
      requestPercentage: 95, limitPercentage: 47.5
    },
    memory: {
      usage: '512Mi', request: '1Gi', limit: '2Gi',
      usageValue: 512, requestValue: 1024, limitValue: 2048,
      requestPercentage: 50, limitPercentage: 25
    },
    labels: { app: 'web-server', tier: 'frontend' }
  },
  {
    name: 'api-service-heavy-load-6b4c8d',
    namespace: 'app-prod',
    containerName: 'java-app',
    cpu: {
      usage: '1.6', request: '2000m', limit: '4000m',
      usageValue: 1600, requestValue: 2000, limitValue: 4000,
      requestPercentage: 80, limitPercentage: 40
    },
    memory: {
      usage: '3.2Gi', request: '4Gi', limit: '6Gi',
      usageValue: 3277, requestValue: 4096, limitValue: 6144,
      requestPercentage: 80, limitPercentage: 53.3
    },
    labels: { app: 'api-service', tier: 'backend' }
  },
  {
    name: 'worker-intensive-task-9f3e1a',
    namespace: 'app-prod',
    containerName: 'python-worker',
    cpu: {
      usage: '3.4', request: '3000m', limit: '4000m',
      usageValue: 3400, requestValue: 3000, limitValue: 4000,
      requestPercentage: 113.3, limitPercentage: 85
    },
    memory: {
      usage: '1.8Gi', request: '2Gi', limit: '3Gi',
      usageValue: 1843, requestValue: 2048, limitValue: 3072,
      requestPercentage: 90, limitPercentage: 60
    },
    labels: { app: 'worker', type: 'batch-job' }
  },

  // High Memory pods (>80%)
  {
    name: 'redis-cache-memory-hog-5c7e2b',
    namespace: 'app-prod',
    containerName: 'redis',
    cpu: {
      usage: '200m', request: '500m', limit: '1000m',
      usageValue: 200, requestValue: 500, limitValue: 1000,
      requestPercentage: 40, limitPercentage: 20
    },
    memory: {
      usage: '7.2Gi', request: '8Gi', limit: '10Gi',
      usageValue: 7373, requestValue: 8192, limitValue: 10240,
      requestPercentage: 90, limitPercentage: 72
    },
    labels: { app: 'redis', tier: 'cache' }
  },
  {
    name: 'database-primary-8a1f4c',
    namespace: 'app-prod',
    containerName: 'postgres',
    cpu: {
      usage: '800m', request: '1000m', limit: '2000m',
      usageValue: 800, requestValue: 1000, limitValue: 2000,
      requestPercentage: 80, limitPercentage: 40
    },
    memory: {
      usage: '11.5Gi', request: '12Gi', limit: '16Gi',
      usageValue: 11776, requestValue: 12288, limitValue: 16384,
      requestPercentage: 95.8, limitPercentage: 71.9
    },
    labels: { app: 'postgres', role: 'primary' }
  },

  // Low CPU pods (<40%)
  {
    name: 'monitoring-agent-light-3d9b7e',
    namespace: 'monitoring',
    containerName: 'prometheus-agent',
    cpu: {
      usage: '50m', request: '200m', limit: '500m',
      usageValue: 50, requestValue: 200, limitValue: 500,
      requestPercentage: 25, limitPercentage: 10
    },
    memory: {
      usage: '128Mi', request: '256Mi', limit: '512Mi',
      usageValue: 128, requestValue: 256, limitValue: 512,
      requestPercentage: 50, limitPercentage: 25
    },
    labels: { app: 'prometheus', component: 'agent' }
  },
  {
    name: 'log-shipper-idle-2f6c8a',
    namespace: 'kube-system',
    containerName: 'fluentd',
    cpu: {
      usage: '30m', request: '100m', limit: '200m',
      usageValue: 30, requestValue: 100, limitValue: 200,
      requestPercentage: 30, limitPercentage: 15
    },
    memory: {
      usage: '64Mi', request: '128Mi', limit: '256Mi',
      usageValue: 64, requestValue: 128, limitValue: 256,
      requestPercentage: 50, limitPercentage: 25
    },
    labels: { app: 'fluentd', 'k8s-app': 'logging' }
  },

  // Low Memory pods (<40%)
  {
    name: 'config-reloader-minimal-4e8d1f',
    namespace: 'kube-system',
    containerName: 'config-reloader',
    cpu: {
      usage: '100m', request: '200m', limit: '500m',
      usageValue: 100, requestValue: 200, limitValue: 500,
      requestPercentage: 50, limitPercentage: 20
    },
    memory: {
      usage: '32Mi', request: '128Mi', limit: '256Mi',
      usageValue: 32, requestValue: 128, limitValue: 256,
      requestPercentage: 25, limitPercentage: 12.5
    },
    labels: { app: 'config-reloader', component: 'utility' }
  },
  {
    name: 'health-checker-lightweight-7b5a9c',
    namespace: 'default',
    containerName: 'health-check',
    cpu: {
      usage: '150m', request: '300m', limit: '600m',
      usageValue: 150, requestValue: 300, limitValue: 600,
      requestPercentage: 50, limitPercentage: 25
    },
    memory: {
      usage: '48Mi', request: '256Mi', limit: '512Mi',
      usageValue: 48, requestValue: 256, limitValue: 512,
      requestPercentage: 18.8, limitPercentage: 9.4
    },
    labels: { app: 'health-checker', type: 'utility' }
  },

  // Normal usage pods
  {
    name: 'web-app-balanced-9c2d4f',
    namespace: 'app-staging',
    containerName: 'node-app',
    cpu: {
      usage: '400m', request: '500m', limit: '1000m',
      usageValue: 400, requestValue: 500, limitValue: 1000,
      requestPercentage: 80, limitPercentage: 40
    },
    memory: {
      usage: '512Mi', request: '1Gi', limit: '2Gi',
      usageValue: 512, requestValue: 1024, limitValue: 2048,
      requestPercentage: 50, limitPercentage: 25
    },
    labels: { app: 'web-app', env: 'staging' }
  },
  {
    name: 'message-queue-stable-1a7f3e',
    namespace: 'app-staging',
    containerName: 'rabbitmq',
    cpu: {
      usage: '300m', request: '500m', limit: '1000m',
      usageValue: 300, requestValue: 500, limitValue: 1000,
      requestPercentage: 60, limitPercentage: 30
    },
    memory: {
      usage: '768Mi', request: '1Gi', limit: '2Gi',
      usageValue: 768, requestValue: 1024, limitValue: 2048,
      requestPercentage: 75, limitPercentage: 37.5
    },
    labels: { app: 'rabbitmq', tier: 'messaging' }
  },
  {
    name: 'auth-service-steady-6e9b2d',
    namespace: 'app-prod',
    containerName: 'oauth-server',
    cpu: {
      usage: '250m', request: '400m', limit: '800m',
      usageValue: 250, requestValue: 400, limitValue: 800,
      requestPercentage: 62.5, limitPercentage: 31.25
    },
    memory: {
      usage: '320Mi', request: '512Mi', limit: '1Gi',
      usageValue: 320, requestValue: 512, limitValue: 1024,
      requestPercentage: 62.5, limitPercentage: 31.25
    },
    labels: { app: 'auth-service', tier: 'security' }
  }
];

export const fetchClusterInfo = async (): Promise<string> => {
  if (SAFE_USE_MOCK_DATA) {
    // Simulate API delay for realistic testing
    await new Promise(resolve => setTimeout(resolve, 200));
    return 'Development Cluster';
  }

  try {
    const response = await fetch(`${API_BASE_URL}/cluster/info`);
    const data: ClusterInfo = await response.json();
    return data.clusterName;
  } catch (error) {
    console.error('Error fetching cluster info:', error);
    return 'Unknown';
  }
};

export const fetchNamespaces = async (): Promise<string[]> => {
  if (SAFE_USE_MOCK_DATA) {
    // Simulate API delay for realistic testing
    await new Promise(resolve => setTimeout(resolve, 300));
    return MOCK_NAMESPACES;
  }

  try {
    const response = await fetch(`${API_BASE_URL}/namespaces`);
    const data: NamespaceList = await response.json();
    return data.namespaces;
  } catch (error) {
    console.error('Error fetching namespaces:', error);
    return [];
  }
};

export const fetchPodMetrics = async (namespace?: string): Promise<PodMetrics[]> => {
  if (SAFE_USE_MOCK_DATA) {
    // Simulate API delay for realistic testing
    await new Promise(resolve => setTimeout(resolve, 500));

    // Filter by namespace if provided
    const filteredPods = namespace
      ? MOCK_PODS.filter(pod => pod.namespace === namespace)
      : MOCK_PODS;

    return filteredPods;
  }

  try {
    const url = namespace
      ? `${API_BASE_URL}/pods?namespace=${namespace}`
      : `${API_BASE_URL}/pods`;

    const response = await fetch(url);
    const data: PodMetricsList = await response.json();

    // Handle null pods case - ensure we always return an array
    return data.pods || [];
  } catch (error) {
    console.error('Error fetching pod metrics:', error);
    return [];
  }
};

export const fetchPodSummary = async (namespace?: string): Promise<PodSummaryResponse | null> => {
  if (SAFE_USE_MOCK_DATA) {
    // Simulate API delay for realistic testing
    await new Promise(resolve => setTimeout(resolve, 400));

    // Filter pods by namespace if provided
    const filteredPods = namespace
      ? MOCK_PODS.filter(pod => pod.namespace === namespace)
      : MOCK_PODS;

    // Calculate summary statistics from mock data
    const totalPods = filteredPods.length;
    const averageCpuUsage = totalPods > 0
      ? filteredPods.reduce((sum, pod) => sum + pod.cpu.requestPercentage, 0) / totalPods
      : 0;
    const averageMemoryUsage = totalPods > 0
      ? filteredPods.reduce((sum, pod) => sum + pod.memory.requestPercentage, 0) / totalPods
      : 0;

    const highCpuPods = filteredPods.filter(pod => pod.cpu.requestPercentage > 80).length;
    const highMemoryPods = filteredPods.filter(pod => pod.memory.requestPercentage > 80).length;
    const lowCpuPods = filteredPods.filter(pod => pod.cpu.requestPercentage < 40 && pod.cpu.requestPercentage > 0).length;
    const lowMemoryPods = filteredPods.filter(pod => pod.memory.requestPercentage < 40 && pod.memory.requestPercentage > 0).length;

    return {
      totalPods,
      averageCpuUsage,
      averageMemoryUsage,
      highCpuPods,
      highMemoryPods,
      lowCpuPods,
      lowMemoryPods,
      generatedAt: new Date().toISOString()
    };
  }

  try {
    const url = namespace
      ? `${API_BASE_URL}/pods/summary?namespace=${namespace}`
      : `${API_BASE_URL}/pods/summary`;

    const response = await fetch(url);
    const data: PodSummaryResponse = await response.json();

    return data;
  } catch (error) {
    console.error('Error fetching pod summary:', error);
    return null;
  }
};

export const fetchPodRecommendations = async (namespace?: string): Promise<RecommendationsResponse | null> => {
  if (SAFE_USE_MOCK_DATA) {
    // Simulate API delay for realistic testing
    await new Promise(resolve => setTimeout(resolve, 600));

    // Generate mock recommendations based on pod data
    const filteredPods = namespace
      ? MOCK_PODS.filter(pod => pod.namespace === namespace)
      : MOCK_PODS;

    const recommendations: PodResourceRecommendation[] = filteredPods.map(pod => {
      // Determine if pod needs optimization
      const cpuUtilization = pod.cpu.requestPercentage;
      const memoryUtilization = pod.memory.requestPercentage;
      const hasOptimalCpuUsage = cpuUtilization >= 60 && cpuUtilization <= 75;
      const hasOptimalMemoryUsage = memoryUtilization >= 60 && memoryUtilization <= 75;
      const hasCpuLimit = pod.cpu.limitValue > 0;
      const memoryMisaligned = Math.abs(pod.memory.requestValue - pod.memory.limitValue) > pod.memory.requestValue * 0.01;

      let reasons: string[] = [];
      let priority: 'high' | 'medium' | 'low' = 'low';

      // Determine reasons and priority
      if (!hasOptimalCpuUsage) {
        if (cpuUtilization > 75) {
          reasons.push('over_utilized');
          priority = 'high';
        } else if (cpuUtilization < 60) {
          reasons.push('under_utilized');
          priority = priority === 'low' ? 'medium' : priority;
        }
      } else {
        reasons.push('well_optimized');
      }

      if (hasCpuLimit) {
        reasons.push('cpu_limit_present');
        priority = 'high';
      }

      if (memoryMisaligned) {
        reasons.push('memory_misaligned');
        priority = priority === 'low' ? 'medium' : priority;
      }

      if (!hasOptimalMemoryUsage) {
        if (memoryUtilization > 75) {
          reasons.push('over_utilized');
          priority = 'high';
        } else if (memoryUtilization < 60) {
          reasons.push('under_utilized');
          priority = priority === 'low' ? 'medium' : priority;
        }
      }

      // Calculate recommended values
      const targetCpuRequest = hasOptimalCpuUsage ? pod.cpu.usageValue : (pod.cpu.usageValue / 0.7);
      const targetMemoryRequest = hasOptimalMemoryUsage ? pod.memory.usageValue : (pod.memory.usageValue / 0.7);

      return {
        podName: pod.name,
        namespace: pod.namespace,
        containerName: pod.containerName,
        cpu: {
          currentRequest: pod.cpu.request,
          currentLimit: pod.cpu.limit,
          currentUsage: pod.cpu.usage,
          recommendedRequest: hasOptimalCpuUsage && !hasCpuLimit ? undefined : `${Math.round(targetCpuRequest)}m`,
          recommendedLimit: hasCpuLimit ? undefined : undefined,
          currentUtilization: cpuUtilization,
          targetUtilization: 70,
          currentRequestValue: pod.cpu.requestValue,
          currentLimitValue: pod.cpu.limitValue,
          currentUsageValue: pod.cpu.usageValue,
          recommendedRequestValue: targetCpuRequest,
          recommendedLimitValue: undefined,
          resourceChange: hasCpuLimit ? 'remove_limit' : (hasOptimalCpuUsage ? 'no_change' : 'adjust'),
          percentageChange: hasOptimalCpuUsage && !hasCpuLimit ? 0 : ((targetCpuRequest - pod.cpu.requestValue) / pod.cpu.requestValue) * 100
        },
        memory: {
          currentRequest: pod.memory.request,
          currentLimit: pod.memory.limit,
          currentUsage: pod.memory.usage,
          recommendedRequest: hasOptimalMemoryUsage && !memoryMisaligned ? undefined : `${Math.round(targetMemoryRequest / 1024 / 1024)}Mi`,
          recommendedLimit: hasOptimalMemoryUsage && !memoryMisaligned ? undefined : `${Math.round(targetMemoryRequest / 1024 / 1024)}Mi`,
          currentUtilization: memoryUtilization,
          targetUtilization: 70,
          currentRequestValue: pod.memory.requestValue,
          currentLimitValue: pod.memory.limitValue,
          currentUsageValue: pod.memory.usageValue,
          recommendedRequestValue: targetMemoryRequest,
          recommendedLimitValue: targetMemoryRequest,
          resourceChange: hasOptimalMemoryUsage && !memoryMisaligned ? 'no_change' : 'align',
          percentageChange: hasOptimalMemoryUsage && !memoryMisaligned ? 0 : ((targetMemoryRequest - pod.memory.requestValue) / pod.memory.requestValue) * 100
        },
        reasons,
        priority,
        confidenceScore: 85 + Math.random() * 10, // 85-95%
        estimatedSavings: reasons.includes('under_utilized') ? 'medium' : 'low',
        riskLevel: priority === 'high' ? 'medium' : 'low',
        basedOnDays: 7,
        dataQuality: 'good'
      };
    });

    // Generate summary
    const summary: RecommendationsSummary = {
      totalPodsAnalyzed: recommendations.length,
      podsNeedingOptimization: recommendations.filter(r => !r.reasons.includes('well_optimized') || r.reasons.length > 1).length,
      podsWellOptimized: recommendations.filter(r => r.reasons.includes('well_optimized') && r.reasons.length === 1).length,
      totalCpuRequestIncrease: "500m",
      totalCpuRequestDecrease: "800m",
      totalMemoryRequestChange: "+200Mi",
      estimatedCostSavings: "medium",
      highPriorityRecommendations: recommendations.filter(r => r.priority === 'high').length,
      mediumPriorityRecommendations: recommendations.filter(r => r.priority === 'medium').length,
      lowPriorityRecommendations: recommendations.filter(r => r.priority === 'low').length,
      overUtilizedPods: recommendations.filter(r => r.reasons.includes('over_utilized')).length,
      underUtilizedPods: recommendations.filter(r => r.reasons.includes('under_utilized')).length,
      podsWithCpuLimits: recommendations.filter(r => r.reasons.includes('cpu_limit_present')).length,
      memoryMisalignedPods: recommendations.filter(r => r.reasons.includes('memory_misaligned')).length,
      podsMissingRequests: 0
    };

    return {
      recommendations,
      summary,
      generatedAt: new Date().toISOString(),
      timeRange: {
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString()
      },
      analysisWindow: "7 days",
      targetUtilization: 70
    };
  }

  try {
    const url = namespace
      ? `${API_BASE_URL}/pods/recommendations?namespace=${namespace}`
      : `${API_BASE_URL}/pods/recommendations`;

    const response = await fetch(url);

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data: RecommendationsResponse = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching pod recommendations:', error);
    return null;
  }
};
