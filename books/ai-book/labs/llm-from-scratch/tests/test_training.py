from pathlib import Path

import pytest
import torch

from config import TrainingConfig
from train import configure_reproducibility, load_checkpoint, train_text


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
    result = train_text("abc " * 100, tiny_training_config(max_steps=30), lab_output(tmp_path))

    assert result.final_train_loss < result.initial_train_loss


def test_checkpoint_restores_model_and_tokenizer(tmp_path):
    """Omitting model configuration or vocabulary from checkpoints must break restoration."""
    result = train_text("hello " * 60, tiny_training_config(max_steps=2), lab_output(tmp_path))

    model, config, tokenizer = load_checkpoint(result.checkpoint_path, torch.device("cpu"))

    assert config.vocab_size == len(tokenizer)
    assert model(torch.tensor([[0]], dtype=torch.long))[0].shape[-1] == len(tokenizer)
    saved_state = torch.load(result.checkpoint_path, map_location="cpu")["model_state_dict"]
    assert torch.equal(model.token_embedding.weight.cpu(), saved_state["token_embedding.weight"])


def lab_output(tmp_path) -> Path:
    return Path(__file__).resolve().parents[1] / "out" / tmp_path.name


def test_training_rejects_checkpoint_directory_outside_lab_output(tmp_path):
    """Writing checkpoints into source or arbitrary paths must be rejected."""
    with pytest.raises(ValueError, match="must be inside"):
        train_text("abc " * 100, tiny_training_config(max_steps=1), tmp_path)


def test_reproducibility_configuration_enables_deterministic_algorithms():
    """Dropping explicit deterministic configuration must make this observable guarantee fail."""
    configure_reproducibility(seed=23)

    assert torch.are_deterministic_algorithms_enabled()
