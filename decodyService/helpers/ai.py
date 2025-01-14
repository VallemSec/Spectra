import json
import os

import openai

from helpers.types import DecodyPromptFormat


class AI:
    """
    A class to interact with the OpenAI API for generating simplified explanations
    of errors and summaries in Dutch. It initializes with environment variables
    for API configuration and provides methods to generate advice for specific
    errors and a comprehensive report.

    Methods:
        generate_category_ai_advice(errors: list[str]) -> str:
            Generates a simplified explanation for a list of errors.

        generate_complete_ai_advice(category_advices: list[str]) -> str:
            Generates a comprehensive report based on multiple simplified
            explanations.
    """

    def __init__(self, prompts: DecodyPromptFormat):
        # Define some initialisation for AI methods
        self._url = os.getenv("OPENAI_API_URL")
        self._api_key = os.getenv("OPENAI_API_KEY")
        self._model = os.getenv("OPENAI_MODEL_NAME")
        self._client = openai.OpenAI(
            api_key=self._api_key, base_url=self._url, timeout=30)
        self._prompts = prompts

    def generate_category_ai_advice(self, findings: list[str], explanations: list[str]) -> str:
        # Generate an ELIA5 description for a single error
        if self._prompts.get("category_prompt") is None:
            return ""
        response = self._client.chat.completions.create(
            model=self._model,
            messages=[{
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        # "text": "Explain the findings like I am five but in Dutch"
                        "text": self._prompts["category_prompt"]
                    },
                    {"type": "text", "text": json.dumps(findings)},
                    {"type": "text", "text": json.dumps(explanations)}
                ]
            }]
        )
        return json.loads(
            response.choices[0].message.model_dump_json()).get("content")

    def generate_complete_ai_advice(self, category_advices: list[str]) -> str:
        # Generate an ELIA5 report for everything in general
        if self._prompts.get("summary_prompt") is None:
            return ""
        response = self._client.chat.completions.create(
            model=self._model,
            messages=[{
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        # "text": "Summarize like I am five but in Dutch"
                        "text": self._prompts["summary_prompt"]
                    },
                    {"type": "text", "text": json.dumps(category_advices)}
                ]
            }]
        )
        return json.loads(
            response.choices[0].message.model_dump_json()).get("content")
