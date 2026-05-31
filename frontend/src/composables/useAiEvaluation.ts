import { ref } from 'vue';
import { API_BASE_URL } from '@/config';

export interface EvaluationMetrics {
  accuracy: number;
  precision_macro: number;
  recall_macro: number;
  f1_macro: number;
}

export interface BinaryEvaluationMetrics {
  accuracy: number;
  precision: number;
  recall: number;
  f1: number;
}

export interface EvaluationResponse {
  confusion_matrix: number[][];
  metrics: EvaluationMetrics;
  binary_metrics: BinaryEvaluationMetrics;
}

export function useAiEvaluation() {
  const loading = ref(false);

  const evaluate = async (path: string): Promise<EvaluationResponse | null> => {
    loading.value = true;
    try {
      const response = await fetch(`${API_BASE_URL}/ai/evaluate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataset_path: path }),
      });

      if (!response.ok) return null;
      return await response.json();
    } catch {
      return null;
    } finally {
      loading.value = false;
    }
  };

  return { evaluate, loading };
}
