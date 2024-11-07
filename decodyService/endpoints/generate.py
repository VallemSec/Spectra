from flask import Blueprint
import json

from helpers import Database, AI
from helpers.types import DecodyOutputResultFormat

generate_app = Blueprint("generate_app", __name__)


@generate_app.get("/generate/<request_id>")
def generate_endpoint(request_id: str):
    db_entry = Database.KeyStorage.get(f"{request_id}-results")
    if db_entry is None:
        return "request_id not found", 404
    results: list[DecodyOutputResultFormat] = json.loads(db_entry)

    ai = AI()
    description_list = [value["description"] for value in results]
    ai_advice = ai.generate_complete_ai_advice(description_list)

    Database.KeyStorage.delete(f"{request_id}-results")
    Database.KeyStorage.delete(f"{request_id}-input")

    return {
        "advice": ai_advice,
        "results": results
    }
