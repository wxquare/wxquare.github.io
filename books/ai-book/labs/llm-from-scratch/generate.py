from __future__ import annotations

import argparse
from pathlib import Path

import torch

from train import choose_device, load_checkpoint


def generate_text(
    checkpoint_path: Path,
    prompt: str,
    max_new_tokens: int,
    temperature: float,
    top_k: int | None,
    device: torch.device,
) -> str:
    model, _, tokenizer = load_checkpoint(checkpoint_path, device)
    if not prompt:
        raise ValueError("prompt cannot be empty")
    model.eval()
    prompt_tokens = torch.tensor([tokenizer.encode(prompt)], dtype=torch.long, device=device)
    generated = model.generate(prompt_tokens, max_new_tokens, temperature=temperature, top_k=top_k)
    return tokenizer.decode(generated[0].tolist())


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate text from a trained GPT checkpoint.")
    parser.add_argument("--checkpoint", type=Path, required=True)
    parser.add_argument("--prompt", required=True)
    parser.add_argument("--max-new-tokens", type=int, default=100)
    parser.add_argument("--temperature", type=float, default=1.0)
    parser.add_argument("--top-k", type=int, default=None)
    parser.add_argument("--device", default=None, help="cpu, mps, or cuda; defaults to automatic selection")
    args = parser.parse_args()

    print(
        generate_text(
            args.checkpoint,
            args.prompt,
            args.max_new_tokens,
            args.temperature,
            args.top_k,
            choose_device(args.device),
        )
    )


if __name__ == "__main__":
    main()
