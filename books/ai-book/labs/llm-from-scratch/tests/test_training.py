import torch

from config import TrainingConfig
from train import load_checkpoint, train_text


def tiny_training_config(max_steps: int) -> TrainingConfig:
    return TrainingConfig(
        batch_size=8,
        learning_rate=1e-2,
        max_steps=max_steps,
        eval_interval=max_steps,
        eval_batches=4,
        seed=11,
    )


def test_training_reduces_loss_on_repeated_pattern(tmp_path):
    """A broken optimizer step or target alignment must prevent this loss reduction."""
    result = train_text("abc " * 100, tiny_training_config(max_steps=30), tmp_path)

    assert result.final_train_loss < result.initial_train_loss


def test_checkpoint_restores_model_and_tokenizer(tmp_path):
    """Omitting model configuration or vocabulary from checkpoints must break restoration."""
    result = train_text("hello " * 60, tiny_training_config(max_steps=2), tmp_path)

    model, config, tokenizer = load_checkpoint(result.checkpoint_path, torch.device("cpu"))

    assert config.vocab_size == len(tokenizer)
    assert model(torch.tensor([[0]], dtype=torch.long))[0].shape[-1] == len(tokenizer)
