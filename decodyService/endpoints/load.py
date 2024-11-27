from flask import Blueprint, request
import jsonschema
import logging
import json
import os

from helpers import safe_eval, Database
from helpers.types import LoadEndpointInputFormat, DecodyDatabaseRuleFormat, DecodyOutputResultFormat

load_app = Blueprint("load_app", __name__)
logger = logging.getLogger(__name__)

with open(os.getenv("INPUTSCHEMA"), "r", encoding="utf-8") as f:
    logger.debug("Loaded input schema")
    schema = json.load(f)


@load_app.post("/load/<request_id>")
def load_endpoint(request_id: str) -> tuple[str, int]:
    """
    This endpoint loads the given data into the database
    after parsing it.
    :param request_id: An identifier to link data
    between requests.
    :return: 201 created
    """
    # Validate request body
    if not request.is_json:
        logger.error("Body is not JSON")
        return "Body is not JSON", 400
    request_body: LoadEndpointInputFormat = request.json
    try:
        jsonschema.validate(instance=request_body, schema=schema)
    except jsonschema.ValidationError as e:
        logger.error("Validation failed, body not properly formatted", e)
        return "Body not properly formatted", 400

    # Get all input objects and check if the request body is a duplicate
    aggregated_input_str = Database.KeyStorage.get(f"{request_id}-input")
    aggregated_input: list[LoadEndpointInputFormat] = json.loads(aggregated_input_str) \
        if aggregated_input_str else []
    for a_input in aggregated_input:
        if a_input == request_body:
            logger.info(
                "Request body already exists in aggregated input, returning early.")
            return "Duplicate request", 409
    aggregated_input.append(request_body)
    Database.KeyStorage.set(
        f"{request_id}-input", json.dumps(aggregated_input, sort_keys=True))
    logger.debug("Verified aggregated input, no duplicate requests made")

    # Fetch all rulesets from the database based on the input
    rules: list[DecodyDatabaseRuleFormat] = []
    for rule_file in request_body.get("rules"):
        rules += Database.fetch_rules(rule_file)
    logger.debug("Fetched rules from the database")

    # Apply rulesets to the request_body and form an result object
    aggregated_results_str = Database.KeyStorage.get(f"{request_id}-results")
    aggregated_results = json.loads(aggregated_results_str) \
        if aggregated_results_str else []
    logger.debug("Loaded aggregated results from the database")
    for rule in rules:
        for result in request_body["results"]:
            if not safe_eval(
                    rule["condition"],
                     {
                         "short": result["short"],
                         "long": result["long"],
                         "scanner_name": request_body["scanner_name"]
                     }):
                logger.debug(f"{rule["condition"]} did not evaluate")
                continue

            aggregated_results.append(DecodyOutputResultFormat(
                category=rule["category"],
                rule_name=rule["name"],
                rule_explanation=rule["explanation"],
                scanner_name=request_body["scanner_name"]
            ))
            logger.debug("Appended to aggregated results")

    Database.KeyStorage.set(
        f"{request_id}-results", json.dumps(aggregated_results))
    logger.debug("Succesfully loaded aggregated results into database")
    return "", 201
