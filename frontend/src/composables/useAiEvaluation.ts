import { ref } from 'vue';

export interface EvaluationResponse {
  binary_metrics: {
    accuracy: number;
    precision: number;
  };
}

export function useAiEvaluation() {
  const loading = ref(false);

  const evaluate = async (path: string): Promise<EvaluationResponse | null> => {
    loading.value = true;
    try {
  const response = await fetch('http://localhost:8080/api/ai/evaluate', {        
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataset_path: path })
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