from flask import Flask, g
import dotenv
import logging
import os
import pymysql

from helpers import Database

import endpoints

dotenv.load_dotenv()

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", handlers=[
        logging.StreamHandler(),
        logging.FileHandler(
            os.getenv("LOG_FILE", "decody.log"), encoding="utf-8"
        )
    ],
    level=os.getenv("LOGLEVEL", "INFO").upper()
)

db_logger = logging.getLogger("db_appcontext")

app = Flask(__name__)
app.register_blueprint(endpoints.load_app)
app.register_blueprint(endpoints.generate_app)


@app.before_request
def open_db():
    if "mariadb_conn" not in g:
        try:
            mariadb_conn = Database.db_connect()
        except pymysql.err.OperationalError:
            db_logger.debug("Failed to open connection to maria db")
            return "Internal server error", 500
        g.mariadb_conn = mariadb_conn
        db_logger.debug("Opened connection to maria db")


@app.teardown_appcontext
def close_db(exception):
    mariadb_conn = g.pop("mariadb_conn", None)
    if mariadb_conn is not None:
        mariadb_conn.close()
        db_logger.debug("Closed connection to maria db")
    else:
        db_logger.debug("No connection to maria db found, nothing to close")


if __name__ == "__main__":
    app.run("0.0.0.0", port=5001, debug=True,
            ssl_context="adhoc" if os.getenv("RUN_TLS", "0") == "1" else None)
