from __future__ import annotations

import argparse
from dataclasses import asdict, dataclass
from pathlib import Path

import torch

from config import ModelConfig, TrainingConfig, smoke_model_config
from data import CharacterTokenizer, TextDataset
from model import GPTLanguageModel


@dataclass(frozen=True)
class TrainingResult:
    checkpoint_path: Path
    initial_train_loss: float
    final_train_loss: float
    final_validation_loss: float


LAB_ROOT = Path(__file__).resolve().parent
OUTPUT_ROOT = (LAB_ROOT / "out").resolve()


def output_directory(requested: str | Path) -> Path:
    candidate = Path(requested)
    resolved = candidate.resolve() if candidate.is_absolute() else (LAB_ROOT / candidate).resolve()
    if not resolved.is_relative_to(OUTPUT_ROOT):
        raise ValueError(f"checkpoint output directory must be inside {OUTPUT_ROOT}")
    return resolved


def configure_reproducibility(seed: int) -> None:
    torch.manual_seed(seed)
    torch.use_deterministic_algorithms(True, warn_only=True)
    if torch.cuda.is_available():
        torch.backends.cudnn.benchmark = False
        torch.backends.cudnn.deterministic = True


def choose_device(requested: str | None = None) -> torch.device:
    if requested is not None:
        return torch.device(requested)
    if torch.cuda.is_available():
        return torch.device("cuda")
    if torch.backends.mps.is_available():
        return torch.device("mps")
    return torch.device("cpu")


def split_tokens(tokens: list[int], validation_fraction: float) -> tuple[list[int], list[int]]:
    if not 0 < validation_fraction < 1:
        raise ValueError("validation_fraction must be between 0 and 1")
    split = int(len(tokens) * (1 - validation_fraction))
    if split < 2 or len(tokens) - split < 2:
        raise ValueError("text is too short to create train and validation sets")
    return tokens[:split], tokens[split:]


@torch.no_grad()
def estimate_loss(
    model: GPTLanguageModel,
    dataset: TextDataset,
    batch_size: int,
    eval_batches: int,
    device: torch.device,
) -> float:
    if eval_batches < 1:
        raise ValueError("eval_batches must be positive")
    was_training = model.training
    model.eval()
    losses: list[float] = []
    for _ in range(eval_batches):
        inputs, targets = dataset.sample_batch(batch_size)
        _, loss = model(inputs.to(device), targets.to(device))
        assert loss is not None
        losses.append(loss.item())
    model.train(was_training)
    return sum(losses) / len(losses)


def save_checkpoint(
    path: Path,
    model: GPTLanguageModel,
    model_config: ModelConfig,
    tokenizer: CharacterTokenizer,
) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    torch.save(
        {
            "model_state_dict": model.state_dict(),
            "model_config": asdict(model_config),
            "vocabulary": list(tokenizer.vocabulary),
        },
        path,
    )


def load_checkpoint(path: Path, device: torch.device) -> tuple[GPTLanguageModel, ModelConfig, CharacterTokenizer]:
    payload = torch.load(path, map_location=device)
    config = ModelConfig(**payload["model_config"])
    tokenizer = CharacterTokenizer(tuple(payload["vocabulary"]))
    model = GPTLanguageModel(config).to(device)
    model.load_state_dict(payload["model_state_dict"])
    return model, config, tokenizer


def train_text(
    text: str,
    training_config: TrainingConfig,
    out_dir: Path,
    model_config: ModelConfig | None = None,
    device: torch.device | None = None,
) -> TrainingResult:
    out_dir = output_directory(out_dir)
    configure_reproducibility(training_config.seed)
    tokenizer = CharacterTokenizer.fit(text)
    config = model_config or smoke_model_config(len(tokenizer))
    if config.vocab_size != len(tokenizer):
        raise ValueError("model_config.vocab_size must equal tokenizer vocabulary size")
    encoded = tokenizer.encode(text)
    train_tokens, validation_tokens = split_tokens(encoded, training_config.validation_fraction)
    if min(len(train_tokens), len(validation_tokens)) < config.block_size + 1:
        raise ValueError("text split is shorter than model block_size + 1")

    generator = torch.Generator().manual_seed(training_config.seed)
    train_dataset = TextDataset(train_tokens, config.block_size, generator)
    validation_dataset = TextDataset(validation_tokens, config.block_size, generator)
    active_device = device or choose_device()
    model = GPTLanguageModel(config).to(active_device)
    optimizer = torch.optim.AdamW(
        model.parameters(), lr=training_config.learning_rate, weight_decay=training_config.weight_decay
    )

    initial_train_loss = estimate_loss(
        model, train_dataset, training_config.batch_size, training_config.eval_batches, active_device
    )
    for _ in range(training_config.max_steps):
        inputs, targets = train_dataset.sample_batch(training_config.batch_size)
        _, loss = model(inputs.to(active_device), targets.to(active_device))
        assert loss is not None
        optimizer.zero_grad(set_to_none=True)
        loss.backward()
        torch.nn.utils.clip_grad_norm_(model.parameters(), training_config.grad_clip)
        optimizer.step()

    final_train_loss = estimate_loss(
        model, train_dataset, training_config.batch_size, training_config.eval_batches, active_device
    )
    final_validation_loss = estimate_loss(
        model, validation_dataset, training_config.batch_size, training_config.eval_batches, active_device
    )
    checkpoint_path = out_dir / "checkpoint.pt"
    save_checkpoint(checkpoint_path, model, config, tokenizer)
    return TrainingResult(checkpoint_path, initial_train_loss, final_train_loss, final_validation_loss)


def _model_config(preset: str, vocab_size: int) -> ModelConfig:
    if preset == "smoke":
        return smoke_model_config(vocab_size)
    if preset == "default":
        return ModelConfig(vocab_size=vocab_size)
    raise ValueError(f"unknown preset: {preset}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Train a small GPT from random initialization.")
    parser.add_argument("--input", type=Path, required=True, help="UTF-8 plain-text training corpus")
    parser.add_argument("--out-dir", type=Path, default=Path("out/run"))
    parser.add_argument("--preset", choices=("smoke", "default"), default="smoke")
    parser.add_argument("--max-steps", type=int, default=100)
    parser.add_argument("--device", default=None, help="cpu, mps, or cuda; defaults to automatic selection")
    parser.add_argument("--seed", type=int, default=1337)
    args = parser.parse_args()

    text = args.input.read_text(encoding="utf-8")
    tokenizer = CharacterTokenizer.fit(text)
    result = train_text(
        text,
        TrainingConfig(max_steps=args.max_steps, seed=args.seed),
        output_directory(args.out_dir),
        model_config=_model_config(args.preset, len(tokenizer)),
        device=choose_device(args.device),
    )
    print(f"initial_train_loss={result.initial_train_loss:.4f}")
    print(f"final_train_loss={result.final_train_loss:.4f}")
    print(f"final_validation_loss={result.final_validation_loss:.4f}")
    print(f"checkpoint={result.checkpoint_path}")


if __name__ == "__main__":
    main()
