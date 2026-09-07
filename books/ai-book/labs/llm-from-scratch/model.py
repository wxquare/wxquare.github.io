from __future__ import annotations

import math

import torch
from torch import Tensor, nn
from torch.nn import functional as F

from config import ModelConfig


class CausalSelfAttention(nn.Module):
    def __init__(self, config: ModelConfig):
        super().__init__()
        if config.n_embd % config.n_head != 0:
            raise ValueError("n_embd must be divisible by n_head")
        self.n_head = config.n_head
        self.head_size = config.n_embd // config.n_head
        self.query_key_value = nn.Linear(config.n_embd, 3 * config.n_embd)
        self.projection = nn.Linear(config.n_embd, config.n_embd)
        self.dropout = nn.Dropout(config.dropout)
        self.register_buffer(
            "causal_mask",
            torch.tril(torch.ones(config.block_size, config.block_size, dtype=torch.bool)),
        )

    def forward(self, inputs: Tensor) -> Tensor:
        batch_size, time, channels = inputs.shape
        query, key, value = self.query_key_value(inputs).chunk(3, dim=-1)
        query = query.view(batch_size, time, self.n_head, self.head_size).transpose(1, 2)
        key = key.view(batch_size, time, self.n_head, self.head_size).transpose(1, 2)
        value = value.view(batch_size, time, self.n_head, self.head_size).transpose(1, 2)

        attention = (query @ key.transpose(-2, -1)) / math.sqrt(self.head_size)
        attention = attention.masked_fill(~self.causal_mask[:time, :time], float("-inf"))
        attention = self.dropout(F.softmax(attention, dim=-1))
        output = attention @ value
        output = output.transpose(1, 2).contiguous().view(batch_size, time, channels)
        return self.dropout(self.projection(output))


class FeedForward(nn.Module):
    def __init__(self, config: ModelConfig):
        super().__init__()
        self.layers = nn.Sequential(
            nn.Linear(config.n_embd, 4 * config.n_embd),
            nn.GELU(),
            nn.Linear(4 * config.n_embd, config.n_embd),
            nn.Dropout(config.dropout),
        )

    def forward(self, inputs: Tensor) -> Tensor:
        return self.layers(inputs)


class TransformerBlock(nn.Module):
    def __init__(self, config: ModelConfig):
        super().__init__()
        self.attention_norm = nn.LayerNorm(config.n_embd)
        self.attention = CausalSelfAttention(config)
        self.feed_forward_norm = nn.LayerNorm(config.n_embd)
        self.feed_forward = FeedForward(config)

    def forward(self, inputs: Tensor) -> Tensor:
        inputs = inputs + self.attention(self.attention_norm(inputs))
        return inputs + self.feed_forward(self.feed_forward_norm(inputs))


class GPTLanguageModel(nn.Module):
    def __init__(self, config: ModelConfig):
        super().__init__()
        self.config = config
        self.token_embedding = nn.Embedding(config.vocab_size, config.n_embd)
        self.position_embedding = nn.Embedding(config.block_size, config.n_embd)
        self.dropout = nn.Dropout(config.dropout)
        self.blocks = nn.Sequential(*(TransformerBlock(config) for _ in range(config.n_layer)))
        self.final_norm = nn.LayerNorm(config.n_embd)
        self.language_model_head = nn.Linear(config.n_embd, config.vocab_size, bias=False)
        self.apply(self._init_weights)

    @staticmethod
    def _init_weights(module: nn.Module) -> None:
        if isinstance(module, nn.Linear):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)
            if module.bias is not None:
                nn.init.zeros_(module.bias)
        elif isinstance(module, nn.Embedding):
            nn.init.normal_(module.weight, mean=0.0, std=0.02)

    def forward(self, idx: Tensor, targets: Tensor | None = None) -> tuple[Tensor, Tensor | None]:
        if idx.ndim != 2:
            raise ValueError("idx must have shape [batch, time]")
        _, time = idx.shape
        if time > self.config.block_size:
            raise ValueError(f"input length {time} exceeds block_size {self.config.block_size}")

        positions = torch.arange(time, device=idx.device)
        embeddings = self.token_embedding(idx) + self.position_embedding(positions)
        logits = self.language_model_head(self.final_norm(self.blocks(self.dropout(embeddings))))

        loss = None
        if targets is not None:
            if targets.shape != idx.shape:
                raise ValueError("targets must have the same shape as idx")
            loss = F.cross_entropy(logits.reshape(-1, logits.size(-1)), targets.reshape(-1))
        return logits, loss

    @torch.no_grad()
    def generate(
        self,
        idx: Tensor,
        max_new_tokens: int,
        temperature: float = 1.0,
        top_k: int | None = None,
    ) -> Tensor:
        if max_new_tokens < 0:
            raise ValueError("max_new_tokens must not be negative")
        if temperature <= 0:
            raise ValueError("temperature must be positive")
        for _ in range(max_new_tokens):
            context = idx[:, -self.config.block_size :]
            logits, _ = self(context)
            next_logits = logits[:, -1, :] / temperature
            if top_k is not None:
                if top_k < 1:
                    raise ValueError("top_k must be positive")
                cutoff = torch.topk(next_logits, min(top_k, next_logits.size(-1))).values[:, -1:]
                next_logits = next_logits.masked_fill(next_logits < cutoff, float("-inf"))
            probabilities = F.softmax(next_logits, dim=-1)
            next_token = torch.multinomial(probabilities, num_samples=1)
            idx = torch.cat((idx, next_token), dim=1)
        return idx
