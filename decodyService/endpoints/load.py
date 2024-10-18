from flask import Blueprint, request
import jsonschema
import logging
import json
import os

from helpers import safe_eval, AI, Database


load_app = Blueprint("load_app", __name__)
logger = logging.getLogger(__name__)

with open(os.getenv("INPUTSCHEMA"), "r", encoding="utf-8") as f:
    schema = json.load(f)


@load_app.post("/load/<request_id>")
def load_endpoint(request_id: str) -> tuple[str, int]:
    """
    This endpoint loads the given data into the database
    after parsing it.
    :param request_id: An identifier that to link data
    between requests.
    :return: 201 created
    """
    # Validate request body
    if not request.is_json:
        return "Body not JSON", 400
    request_body: dict = request.json
    try:
        jsonschema.validate(instance=request_body, schema=schema)
    except jsonschema.ValidationError:
        logger.debug("Validation failed, body not properly formatted")
        return "Body not properly formatted", 400
    Database.KeyStorage.set(f"{request_id}-input", json.dumps(request_body))

    ai = AI()

    # Fetch all rulesets from the database based on the input
    rules = []
    for rule_file in request_body.get("rules"):
        rules += Database.fetch_rules(rule_file)

    value = Database.KeyStorage.get(f"{request_id}-results")
    results = json.loads(value) if value else list()
    for result in request_body.get("results"):
        result_body = dict()
        for rule in rules:
            if not safe_eval(
                    rule["condition"],
                     {
                         "err_short": result["err_short"]
                     }):
                continue

            result_body["category"] = rule["category"]
            result_body["description"] = rule["explanation"]
            result_body["name"] = rule["name"]
            result_body["ai_advice"] = ai.generate_ai_advice(
                rule["explanation"])
        results.append(result_body)

    Database.KeyStorage.set(f"{request_id}-results",
                            json.dumps(list(filter(None, results))))
    return "", 201
