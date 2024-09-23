from flask import Blueprint


load_app = Blueprint("load_app", __name__)


@load_app.post("/load")
def load_endpoint():
    pass
