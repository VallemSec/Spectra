from flask import Flask
import dotenv
import logging
import os

import endpoints


dotenv.load_dotenv()

logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", handlers=[
    logging.StreamHandler(), logging.FileHandler(os.getenv("LOG_FILE", "decody.log"), encoding="utf-8")
], level=os.getenv("LOGLEVEL", "INFO").upper())

db_logger = logging.getLogger("db")

app = Flask(__name__)
app.register_blueprint(endpoints.load_app)
app.register_blueprint(endpoints.generate_app)

if __name__ == "__main__":
    app.run("0.0.0.0", port=5001, debug=True,
            ssl_context="adhoc" if os.getenv("RUN_TLS", "1") else None)
