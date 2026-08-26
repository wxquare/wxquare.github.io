import math

import pytest
import torch

from config import TrainingConfig
from evaluate import evaluate_checkpoint
from generate import generate_text
from train import train_text


@pytest.fixture()
def trained_checkpoint(tmp_path):
    result = train_text(
        "abc " * 100,
        TrainingConfig(batch_size=8, learning_rate=1e-2, max_steps=3, eval_batches=2, seed=19),
        tmp_path,
    )
    return result.checkpoint_path


def test_generation_appends_requested_number_of_tokens(trained_checkpoint):
    """Ignoring max_new_tokens or failing to append sampled tokens breaks this API."""
    text = generate_text(
        trained_checkpoint,
        "ab",
        max_new_tokens=7,
        temperature=1.0,
        top_k=2,
        device=torch.device("cpu"),
    )

    assert len(text) == 9


def test_evaluation_reports_finite_loss_and_perplexity(trained_checkpoint):
    """Returning a stale or invalid loss must not produce a valid evaluation result."""
    metrics = evaluate_checkpoint(trained_checkpoint, "abc " * 100, torch.device("cpu"))

    assert math.isfinite(metrics.loss)
    assert metrics.perplexity > 0
