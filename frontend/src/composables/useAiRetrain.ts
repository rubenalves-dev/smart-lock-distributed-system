import { ref } from 'vue';

export interface TrainingDiagnostics {
  train_accuracy: number;
  validation_accuracy: number;
  train_loss: number;
  validation_loss: number;
  underfitting_detected: boolean;
  overfitting_detected: boolean;
}

export interface RetrainResponse {
  success: boolean;
  message: string;
  diagnostics?: TrainingDiagnostics;
}

export function useAiRetrain() {
  const loading = ref(false);

  const retrain = async (epochs: number, datasetPath: string): Promise<RetrainResponse | null> => {
    loading.value = true;
    try {
      const response = await fetch('https://smartlock-api.raiiaa.dev/api/ai/retrain', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          epochs: epochs,
          dataset_path: datasetPath,
        }),
      });

      if (!response.ok) {
        return null;
      }
      return await response.json();
    } catch {
      return null;
    } finally {
      loading.value = false;
    }
  };

  return { retrain, loading };
}
