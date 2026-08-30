import pytest
import torch

from data import CharacterTokenizer, TextDataset


def test_tokenizer_round_trips_fitted_text():
    """Dropping a character from either mapping must break this round trip."""
    tokenizer = CharacterTokenizer.fit("cab")

    assert tokenizer.decode(tokenizer.encode("abca")) == "abca"


def test_tokenizer_rejects_unknown_character():
    """Silently assigning an id to unseen text would corrupt training data."""
    tokenizer = CharacterTokenizer.fit("ab")

    with pytest.raises(ValueError, match="not in tokenizer vocabulary"):
        tokenizer.encode("ac")


def test_batch_targets_are_inputs_shifted_by_one_token():
    """Using the same position as the target would not train next-token prediction."""
    dataset = TextDataset(
        list(range(12)), block_size=4, generator=torch.Generator().manual_seed(1)
    )

    inputs, targets = dataset.sample_batch(3)

    assert inputs.shape == targets.shape == (3, 4)
    assert torch.equal(targets[:, :-1], inputs[:, 1:])
