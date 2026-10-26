from __future__ import annotations

import argparse
import math
from dataclasses import dataclass
from pathlib import Path

import torch

from data import TextDataset
from train import choose_device, estimate_loss, load_checkpoint, split_tokens


@dataclass(frozen=True)
class EvaluationMetrics:
    loss: float
    perplexity: float


def evaluate_checkpoint(checkpoint_path: Path, text: str, device: torch.device) -> EvaluationMetrics:
    model, config, tokenizer = load_checkpoint(checkpoint_path, device)
    _, validation_tokens = split_tokens(tokenizer.encode(text), validation_fraction=0.1)
    dataset = TextDataset(validation_tokens, config.block_size, torch.Generator().manual_seed(0))
    loss = estimate_loss(model, dataset, batch_size=8, eval_batches=4, device=device)
    return EvaluationMetrics(loss=loss, perplexity=math.exp(loss))


def main() -> None:
    parser = argparse.ArgumentParser(description="Evaluate a trained GPT checkpoint.")
    parser.add_argument("--checkpoint", type=Path, required=True)
    parser.add_argument("--input", type=Path, required=True, help="UTF-8 text evaluated with the checkpoint tokenizer")
    parser.add_argument("--device", default=None, help="cpu, mps, or cuda; defaults to automatic selection")
    args = parser.parse_args()

    metrics = evaluate_checkpoint(args.checkpoint, args.input.read_text(encoding="utf-8"), choose_device(args.device))
    print(f"loss={metrics.loss:.4f}")
    print(f"perplexity={metrics.perplexity:.4f}")


if __name__ == "__main__":
    main()
