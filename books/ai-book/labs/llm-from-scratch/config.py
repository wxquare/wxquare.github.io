from dataclasses import dataclass


@dataclass(frozen=True)
class ModelConfig:
    vocab_size: int
    block_size: int = 128
    n_embd: int = 384
    n_head: int = 6
    n_layer: int = 6
    dropout: float = 0.1


@dataclass(frozen=True)
class TrainingConfig:
    batch_size: int = 32
    learning_rate: float = 3e-4
    weight_decay: float = 0.1
    max_steps: int = 1_000
    eval_interval: int = 100
    eval_batches: int = 10
    grad_clip: float = 1.0
    seed: int = 1337
    validation_fraction: float = 0.1


def smoke_model_config(vocab_size: int) -> ModelConfig:
    return ModelConfig(vocab_size=vocab_size, block_size=32, n_embd=48, n_head=4, n_layer=2)
