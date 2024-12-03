import logging

from flask import Blueprint
import json

from helpers import Database, AI
from helpers.types import DecodyOutputResultFormat, DecodyCategoryOutputFormat

generate_app = Blueprint("generate_app", __name__)
logger = logging.getLogger(__name__)


@generate_app.get("/generate/<request_id>")
def generate_endpoint(request_id: str):
    db_entry = Database.KeyStorage.get(f"{request_id}-results")
    if db_entry is None:
        return "request_id not found", 404
    results: list[DecodyOutputResultFormat] = json.loads(db_entry)

    categories = {value["category"] for value in results}
    description_lists = dict()
    for category in categories:
        category_list = list()
        for result in results:
            if result["category"] == category:
                category_list.append(result["rule_explanation"])
        description_lists[category] = category_list

    ai = AI()
    results: list[DecodyCategoryOutputFormat] = list()
    for category, explanations in description_lists.items():
        ai_category_advice = ai.generate_category_ai_advice(explanations)
        results.append(DecodyCategoryOutputFormat(
            category=category,
            ai_advice=ai_category_advice,
            results=explanations,
        ))

    ai_advice = ai.generate_complete_ai_advice([result["ai_advice"] for result in results])

    return {
        "ai_advice": ai_advice,
        "results": results
    }
