from flask import Blueprint


generate_app = Blueprint("generate_app", __name__)


@generate_app.get("/generate")
def generate_endpoint():
    pass
