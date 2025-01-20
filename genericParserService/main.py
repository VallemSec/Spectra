import logging
import sys

import pymysql

import helpers
from lua_parser import Parser, ParserFinder
from dotenv import load_dotenv
import argparse
import threading
import json
import os

load_dotenv()

PARSER_FOLDER = os.getenv("PARSER_FOLDER")
SQL_CONN_STRING = os.getenv("SQL_CONN_STRING")
if PARSER_FOLDER is None:
    raise ValueError("PARSER_FOLDER environment variable is not set")
if not os.path.exists(PARSER_FOLDER):
    raise ValueError("Expected PARSER_FOLDER to be an existing path")
if SQL_CONN_STRING is None:
    raise ValueError("SQL_STRING environment variable is not set")


connection = pymysql.connect(**helpers.convert_sql_str_to_connect_obj(SQL_CONN_STRING))

parser = argparse.ArgumentParser()
parser.add_argument("name", help="Name to give to the output")
parser.add_argument("target",
                    help="""Which parser file(s) to use.
                    You can specify either a file or directory""")
parser.add_argument("database_key", help="Key of the entry in the database. Format is 'parser-<UUID>'")
parser.add_argument("-v", "--verbose", action="store_true",
                    help="Enable verbose output, aka loglevel DEBUG")

args = parser.parse_args()


logging.basicConfig(
    level=logging.DEBUG if args.verbose else logging.INFO,
    format='{"time": "%(asctime)s", "logger_name": "%(name)s", "level": "%(levelname)s", "message": "%(message)s"}',
    stream=sys.stderr
)


pf = ParserFinder(args.target, PARSER_FOLDER)

lua_parser_threads = []
database_input = ""
for parser_file in pf.yield_parsers():
    c = connection.cursor()
    c.execute("SELECT kv.value FROM key_value kv WHERE kv.key = %s", args.database_key)
    database_input = c.fetchone()["value"]
    lua_parser = Parser(parser_file, database_input)
    thread = threading.Thread(target=lua_parser.parse)
    lua_parser_threads.append({"thread": thread, "parser": lua_parser})
    thread.start()
    logging.debug("Started a thread with id %s", thread.native_id)

thread_active = [True for _ in range(len(lua_parser_threads))]
while True in thread_active:
    for i, thread in enumerate(lua_parser_threads):
        if not thread["thread"].is_alive():
            thread_active[i] = False

results = []
panicked_files = []
for thread in lua_parser_threads:
    thread["parser"].cleanup()
    result = thread["parser"].result
    panicked = thread["parser"].panicked
    if panicked:
        panicked_files.append(thread["parser"].parser_file)
    elif result:
        results += result

if len(panicked_files) > 0:
    logging.error('{"panicked_parsers": %s}', json.dumps(panicked_files))

print(json.dumps({
    "name": args.name,
    "results": results
}))
