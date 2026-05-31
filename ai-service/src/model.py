import os
import csv
import random
import numpy as np
import tensorflow as tf
from keras import layers
from tensorflow import keras

def create_model():
    model = keras.Sequential([
        layers.Dense(16, activation="relu", input_shape=(3,)),
        layers.Dense(16, activation="relu"),
        layers.Dense(4, activation="softmax")
    ])

    model.compile(
        optimizer="adam",
        loss="sparse_categorical_crossentropy",
        metrics=[
            "accuracy",
        ]
    )
    return model

def normalize_features(fails, distance_cm, is_denied):
    """Normalize features to be in the [0, 1] range for better neural network convergence."""
    # fails: scale by 5.0 (5 or more fails = 1.0)
    norm_fails = min(float(fails) / 5.0, 1.0)
    # distance_cm: scale by 150.0 (150cm or more = 1.0)
    norm_dist = min(float(distance_cm) / 150.0, 1.0)
    return [norm_fails, norm_dist, float(is_denied)]

def generate_synthetic_data(filepath, num_samples=1000):
    """Generates a synthetic dataset for baseline training and saves it as a CSV."""
    os.makedirs(os.path.dirname(filepath), exist_ok=True)
    with open(filepath, mode='w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['fails', 'distance_cm', 'is_denied', 'severity'])
        
        for _ in range(num_samples):
            # Class 0: OK (no fails, normal/far distance, not denied)
            # Class 1: Irregular (1 fail, normal distance, or 0 fails but very close)
            # Class 2: Suspicious (2 fails, or 1 fail and close, or access denied)
            # Class 3: Critical (3+ fails, or close with multiple fails)
            
            c = random.choices([0, 1, 2, 3], weights=[0.50, 0.25, 0.15, 0.10])[0]
            if c == 0:
                fails = 0
                distance_cm = random.uniform(50.0, 150.0)
                is_denied = 0.0
            elif c == 1:
                if random.random() > 0.5:
                    fails = 1
                    distance_cm = random.uniform(30.0, 100.0)
                    is_denied = 1.0
                else:
                    fails = 0
                    distance_cm = random.uniform(5.0, 20.0)
                    is_denied = 0.0
            elif c == 2:
                if random.random() > 0.5:
                    fails = 2
                    distance_cm = random.uniform(10.0, 50.0)
                    is_denied = 1.0
                else:
                    fails = 1
                    distance_cm = random.uniform(5.0, 20.0)
                    is_denied = 1.0
            else: # c == 3
                fails = random.randint(3, 6)
                distance_cm = random.uniform(2.0, 15.0)
                is_denied = 1.0
                
            writer.writerow([fails, distance_cm, is_denied, c])
    print(f"Generated synthetic dataset at {filepath} with {num_samples} samples.")

def load_dataset(filepath):
    """Loads and normalizes the dataset from a CSV file."""
    X = []
    y = []
    if not os.path.exists(filepath):
        raise FileNotFoundError(f"Dataset file not found: {filepath}")
        
    with open(filepath, mode='r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            fails = float(row.get('fails', 0.0))
            distance = float(row.get('distance_cm', 150.0))
            is_denied = float(row.get('is_denied', 0.0))
            severity = int(row.get('severity', 0))
            
            norm_features = normalize_features(fails, distance, is_denied)
            X.append(norm_features)
            y.append(severity)
            
    return np.array(X), np.array(y)

def train_model(model, X, y, epochs=10):
    """Trains the model on the provided data and returns the training history."""
    history = model.fit(
        X,
        y,
        epochs=epochs,
        batch_size=32,
        validation_split=0.2,
        verbose=1
    )
    return history