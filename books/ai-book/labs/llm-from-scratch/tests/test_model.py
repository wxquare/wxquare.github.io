import torch

from config import ModelConfig
from model import GPTLanguageModel


def tiny_model_config() -> ModelConfig:
    return ModelConfig(vocab_size=17, block_size=8, n_embd=16, n_head=4, n_layer=2, dropout=0.0)


def test_model_returns_vocab_logits_and_differentiable_loss():
    """Removing the language-model head or loss calculation must fail this contract."""
    model = GPTLanguageModel(tiny_model_config())
    indices = torch.randint(0, 17, (2, 5))

    logits, loss = model(indices, indices)

    assert logits.shape == (2, 5, 17)
    assert loss is not None and torch.isfinite(loss)
    loss.backward()
    assert model.token_embedding.weight.grad is not None


def test_future_tokens_cannot_change_past_logits():
    """Removing the causal mask would let changed future tokens alter early logits."""
    torch.manual_seed(7)
    model = GPTLanguageModel(tiny_model_config()).eval()
    unchanged_prefix = torch.tensor([[1, 2, 3, 4]])
    changed_suffix = torch.tensor([[1, 2, 9, 8]])

    original_logits, _ = model(unchanged_prefix)
    changed_logits, _ = model(changed_suffix)

    assert torch.allclose(original_logits[:, :2], changed_logits[:, :2], atol=1e-5)
