from collections import defaultdict
import logging

from flask import Blueprint
import json

from helpers import Database, AI
from helpers.types import DecodyDatabaseResultFormat, DecodyCategoryOutputFormat, \
    DecodyFindingsOutputFormat

generate_app = Blueprint("generate_app", __name__)
logger = logging.getLogger(__name__)


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

    ai = AI()

    results: list[DecodyCategoryOutputFormat] = list()
    for category, findings in category_findings.items():
        ai_category_advice = ai.generate_category_ai_advice(
            [i["rule_explanation"] for i in findings]
        )
        results.append(DecodyCategoryOutputFormat(
            category=category,
            ai_advice=ai_category_advice,
            results=findings,
        ))

    ai_advice = ai.generate_complete_ai_advice(
        [result["ai_advice"] for result in results])

    return {
        "ai_advice": ai_advice,
        "results": results
    }
