import json
import os

import openai


class AI:
    def __init__(self):
        # Define some initialisation for AI methods
        self._url = os.getenv("OPENAI_API_URL")
        self._api_key = os.getenv("OPENAI_API_KEY")
        self._model = os.getenv("OPENAI_MODEL_NAME")
        self._client = openai.OpenAI(
            api_key=self._api_key, base_url=self._url)

    def generate_category_ai_advice(self, errors: list[str]) -> str:
        # Generate an ELIA5 description for a single error
        response = self._client.chat.completions.create(
            model=self._model,
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": "Explain the errors like I am five but in Dutch"},
                    {"type": "text", "text": json.dumps(errors)}
                ]
            }]
        )
        return json.loads(response.choices[0].message.model_dump_json()).get("content")

    def generate_complete_ai_advice(self, category_advices: list[str]) -> str:
        # Generate an ELIA5 report for everything in general
        response = self._client.chat.completions.create(
            model=self._model,
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": "Summarize like I am five but in Dutch"},
                    {"type": "text", "text": json.dumps(category_advices)}
                ]
            }]
        )
        return json.loads(response.choices[0].message.model_dump_json()).get("content")
