from flask import Blueprint
import json

from helpers import Database, AI


generate_app = Blueprint("generate_app", __name__)


@generate_app.get("/generate/<request_id>")
def generate_endpoint(request_id: str):
    ai = AI()
    db_entry = Database.KeyStorage.get(f"{request_id}-results")
    if db_entry is None: return "request_id not found", 404
    results: list[dict] = json.loads(db_entry)

    description_list = [value["description"] for value in results]
    ai_advice = ai.generate_complete_ai_advice(description_list)

    return {
        "advice": ai_advice,
        "results": results
    }
