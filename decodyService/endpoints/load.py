from flask import Blueprint, g, request
import pymysql
import jsonschema
import logging
import json
import os

from helpers.eval import safe_eval


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

    # Insert request body into table
    with g.mariadb_conn.cursor() as cursor:
        cursor.execute("INSERT INTO key_value (key, value) VALUES (%s, %s)",
                       (f"{request_id}-input", json.dumps(request_body)))

    # Fetch all rulesets from the database based on the input
    g.mariadb_conn: pymysql.Connection
    with g.mariadb_conn.cursor() as cursor:
        rules = []
        for rule in request_body.get("rules"):
            cursor.execute("""
                select r.condition, r.explanation from rules r, files f
                where f.file_name = %s and r.file_id = f.file_id;
            """, (rule,))
            rules += cursor.fetchall()

    # Parse all the rules and add the applicable explanations to keydb
    with g.mariadb_conn.cursor() as cursor:
        cursor.execute("SELECT value FROM key_value WHERE key = %s",
                       (f"{request_id}-explanations",))
        value = cursor.fetchone()[0]
    explanations = set(json.loads(value)) \
        if value is not None else set()
    for result in request_body.get("results"):
        for rule in rules:
            if safe_eval(rule["condition"],
                         {
                             "err_short": result["err_short"]
                         }
                         ):
                explanations.add(rule["explanation"])
    with g.mariadb_conn.cursor() as cursor:
        cursor.execute("INSERT INTO key_value (key, value) VALUES (%s, %s)",
                       (f"{request_id}-explanations", json.dumps(explanations)))
    return "", 201
