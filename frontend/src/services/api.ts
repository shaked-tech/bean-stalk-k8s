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

const API_BASE_URL = '/api';

export const fetchNamespaces = async (): Promise<string[]> => {
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
