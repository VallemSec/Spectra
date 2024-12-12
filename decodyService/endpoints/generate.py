import threading
from collections import defaultdict
import logging

from flask import Blueprint
import json

from helpers import Database, AI
from helpers.types import (
    DecodyDatabaseResultFormat, DecodyCategoryOutputFormat,
    DecodyFindingsOutputFormat, DecodyPromptFormat)

generate_app = Blueprint("generate_app", __name__)
logger = logging.getLogger(__name__)


class AIAdvice:
    def __init__(self, findings: list[str], ai: AI):
        self.findings = findings
        self.ai = ai
        self.advice = None

    def generate(self):
        self.advice = self.ai.generate_category_ai_advice(self.findings)


@generate_app.get("/generate/<request_id>")
def generate_endpoint(request_id: str):
    """
    Handles GET requests to generate AI advice
    and findings for a given request ID.
    :param request_id: An identifier to link data
    between requests.
    :return: 200 and a JSON object containing AI advice
    and findings for a given request ID.
    """
    db_entry = Database.KeyStorage.get(f"{request_id}-results")
    if db_entry is None:
        logger.error("request_id '%s' not found", request_id)
        return "request_id not found", 404
    db_results: list[DecodyDatabaseResultFormat] = json.loads(db_entry)

    category_findings: defaultdict[str, list[DecodyFindingsOutputFormat]] = defaultdict(list)
    for result in db_results:
        category_findings[result["category"]].append(
            DecodyFindingsOutputFormat(
                rule_name=result["rule_name"],
                rule_explanation=result["rule_explanation"],
                scanner_name=result["scanner_name"]
            )
        )

    db_prompts = Database.KeyStorage.get("decody-prompts")
    if db_prompts is None:
        return "", 500
    prompts: DecodyPromptFormat = json.loads(db_prompts)
    ai = AI(prompts)

    threads = []
    for category, findings in category_findings.items():
        ai_thread = AIAdvice([i["rule_explanation"] for i in findings], ai)
        thread = threading.Thread(target=ai_thread.generate)
        threads.append({
            "thread": thread,
            "ai": ai_thread,
            "category": category,
            "findings": findings
        })
        thread.start()

    thread_active = [True for _ in range(len(threads))]
    while True in thread_active:
        for i, thread in enumerate(threads):
            if not thread["thread"].is_alive():
                thread_active[i] = False

    results: list[DecodyCategoryOutputFormat] = list()
    for thread in threads:
        ai_thread = thread["ai"]
        category = thread["category"]
        findings = thread["findings"]
        results.append(DecodyCategoryOutputFormat(
            category=category,
            ai_advice=ai_thread.advice,
            results=findings
        ))

    ai_advice = ai.generate_complete_ai_advice(
        [result["ai_advice"] for result in results])

    return {
        "ai_advice": ai_advice,
        "results": results
    }
