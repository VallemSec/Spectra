from flask import Blueprint, request
import jsonschema
import logging
import json
import os

from helpers import safe_eval, AI, Database
from helpers.types import LoadEndpointInputFormat, DecodyDatabaseRuleFormat

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
    request_body: LoadEndpointInputFormat = request.json
    try:
        jsonschema.validate(instance=request_body, schema=schema)
    except jsonschema.ValidationError:
        logger.error("Validation failed, body not properly formatted")
        return "Body not properly formatted", 400

    # Get all input objects and check if the request body is a duplicate
    aggregated_input_str = Database.KeyStorage.get(f"{request_id}-input")
    aggregated_input: list[LoadEndpointInputFormat] = json.loads(aggregated_input_str) if aggregated_input_str else []
    for ai in aggregated_input:
        if ai == request_body:
            logger.info("Request body already exists in aggregated input, returning early.")
            return "Duplicate request", 409
    aggregated_input.append(request_body)
    Database.KeyStorage.set(f"{request_id}-input", json.dumps(request_body, sort_keys=True))

    # Fetch all rulesets from the database based on the input
    rules: list[DecodyDatabaseRuleFormat] = []
    for rule_file in request_body.get("rules"):
        rules += Database.fetch_rules(rule_file)

    ai = AI()

    # Apply rulesets to the request_body and form an result object
    aggregated_results_str = Database.KeyStorage.get(f"{request_id}-results")
    aggregated_results = json.loads(aggregated_results_str) if aggregated_results_str else []
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
        aggregated_results.append(result_body)

    Database.KeyStorage.set(f"{request_id}-results",
                            json.dumps(list(filter(None, aggregated_results))))
    return "", 201
