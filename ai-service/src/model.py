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
