from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence

import torch


@dataclass(frozen=True)
class CharacterTokenizer:
    vocabulary: tuple[str, ...]

    def __post_init__(self) -> None:
        if not self.vocabulary:
            raise ValueError("tokenizer vocabulary cannot be empty")
        if len(set(self.vocabulary)) != len(self.vocabulary):
            raise ValueError("tokenizer vocabulary contains duplicate characters")
        object.__setattr__(self, "_to_id", {char: index for index, char in enumerate(self.vocabulary)})

    @classmethod
    def fit(cls, text: str) -> CharacterTokenizer:
        if not text:
            raise ValueError("cannot fit a tokenizer on empty text")
        return cls(tuple(sorted(set(text))))

    def __len__(self) -> int:
        return len(self.vocabulary)

    def encode(self, text: str) -> list[int]:
        encoded: list[int] = []
        for char in text:
            try:
                encoded.append(self._to_id[char])
            except KeyError as error:
                raise ValueError(f"character {char!r} is not in tokenizer vocabulary") from error
        return encoded

    def decode(self, tokens: Sequence[int]) -> str:
        decoded: list[str] = []
        for token in tokens:
            if token < 0 or token >= len(self.vocabulary):
                raise ValueError(f"token id {token} is outside tokenizer vocabulary")
            decoded.append(self.vocabulary[token])
        return "".join(decoded)


class TextDataset:
    def __init__(self, tokens: Sequence[int], block_size: int, generator: torch.Generator | None = None):
        if block_size < 1:
            raise ValueError("block_size must be positive")
        if len(tokens) < block_size + 1:
            raise ValueError("tokens must contain at least block_size + 1 items")
        self.tokens = torch.tensor(tokens, dtype=torch.long)
        self.block_size = block_size
        self.generator = generator

    def sample_batch(self, batch_size: int) -> tuple[torch.Tensor, torch.Tensor]:
        if batch_size < 1:
            raise ValueError("batch_size must be positive")
        starts = torch.randint(
            0,
            len(self.tokens) - self.block_size,
            (batch_size,),
            generator=self.generator,
        )
        inputs = torch.stack([self.tokens[start : start + self.block_size] for start in starts])
        targets = torch.stack([self.tokens[start + 1 : start + self.block_size + 1] for start in starts])
        return inputs, targets
