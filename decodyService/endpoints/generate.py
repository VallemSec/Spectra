from flask import Blueprint, g
import json
import keydb


generate_app = Blueprint("generate_app", __name__)


@generate_app.get("/generate/<request_id>")
def generate_endpoint(request_id: str):
    g.keydb_conn: keydb.KeyDB
    explanations_str = g.keydb_conn.get(f"{request_id}-explanations")
    return json.loads(explanations_str), 200
