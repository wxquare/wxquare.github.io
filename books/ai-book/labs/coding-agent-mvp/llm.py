from typing import Protocol
import json
import urllib.error
import urllib.request

from config import LLMConfig


class LLMClient(Protocol):
    def complete(self, prompt: str) -> str:
        ...


class FakeLLM:
    def __init__(self, responses: list[str]):
        self.responses = responses
        self.index = 0

    def complete(self, prompt: str) -> str:
        if self.index >= len(self.responses):
            return json.dumps(
                {
                    "thought": "no more responses",
                    "final": {
                        "summary": "stopped",
                        "verification": "none",
                        "changed_files": [],
                    },
                },
                ensure_ascii=False,
            )
        value = self.responses[self.index]
        self.index += 1
        return value


class DeepSeekLLM:
    def __init__(self, config: LLMConfig):
        if not config.api_key:
            raise ValueError("llm.api_key is required in the config file")
        self.config = config

    def complete(self, prompt: str) -> str:
        payload: dict[str, object] = {
            "model": self.config.model,
            "messages": [{"role": "user", "content": prompt}],
            "temperature": self.config.temperature,
            "max_tokens": self.config.max_tokens,
            "response_format": {"type": "json_object"},
        }
        if self.config.thinking:
            payload["thinking"] = {"type": self.config.thinking}

        request = urllib.request.Request(
            f"{self.config.base_url.rstrip('/')}/chat/completions",
            data=json.dumps(payload).encode("utf-8"),
            headers={
                "Authorization": f"Bearer {self.config.api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )

        try:
            with urllib.request.urlopen(request, timeout=self.config.timeout) as response:
                body = response.read().decode("utf-8")
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")[:2000]
            raise RuntimeError(f"DeepSeek API error {exc.code}: {detail}") from exc
        except urllib.error.URLError as exc:
            raise RuntimeError(f"DeepSeek API request failed: {exc.reason}") from exc

        data = json.loads(body)
        return data["choices"][0]["message"].get("content") or ""


def build_llm(config: LLMConfig) -> LLMClient:
    if config.provider != "deepseek":
        raise ValueError(f"unsupported llm provider: {config.provider}")
    return DeepSeekLLM(config)
