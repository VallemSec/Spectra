from flask import Flask, g, make_response
import dotenv
import logging
import os
import keydb
import pymysql

import endpoints


dotenv.load_dotenv()

logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", handlers=[
    logging.StreamHandler(), logging.FileHandler(os.getenv("LOG_FILE", "decody.log"), encoding="utf-8")
], level=os.getenv("LOGLEVEL", "INFO").upper())

db_logger = logging.getLogger("db_appcontext")

app = Flask(__name__)
app.register_blueprint(endpoints.load_app)
app.register_blueprint(endpoints.generate_app)


@app.before_request
def open_db():
    if "keydb_conn" not in g:
        keydb_conn = keydb.KeyDB(host=os.getenv("KEYDB_HOST", "localhost"), port=int(os.getenv("KEYDB_PORT", "6379")))
        try:
            keydb_conn.ping()
        except keydb.ConnectionError:
            db_logger.debug("Failed to open connection to keydb")
            response = make_response("Internal server error")
            response.status_code = 500
            return response
        g.keydb_conn = keydb_conn
        db_logger.debug("Opened connection to keydb")
    if "mariadb_conn" not in g:
        try:
            mariadb_conn = pymysql.Connect(
                host=os.getenv("MARIADB_HOST", "localhost"),
                port=int(os.getenv("MARIADB_PORT", "3306")),
                user=os.getenv("MARIADB_USER", "root"),
                password=os.getenv("MARIADB_PASSWORD", "password")
            )
        except pymysql.err.OperationalError:
            db_logger.debug("Failed to open connection to maria db")
            response = make_response("Internal server error")
            response.status_code = 500
            return response
        g.mariadb_conn = mariadb_conn
        db_logger.debug("Opened connection to maria db")


@app.teardown_appcontext
def close_db(exception):
    keydb_conn = g.pop("keydb_conn", None)
    mariadb_conn = g.pop("mariadb_conn", None)
    if keydb_conn is not None:
        keydb_conn.close()
        db_logger.debug("Closed connection to keydb")
    else:
        db_logger.debug("No connection to keydb found, nothing to close")
    if mariadb_conn is not None:
        mariadb_conn.close()
        db_logger.debug("Closed connection to maria db")
    else:
        db_logger.debug("No connection to maria db found, nothing to close")


if __name__ == "__main__":
    app.run("0.0.0.0", port=5001, debug=True,
            ssl_context="adhoc" if os.getenv("RUN_TLS", "0") == "1" else None)
