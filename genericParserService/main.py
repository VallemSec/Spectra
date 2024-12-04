import logging
import sys

from lua_parser import Parser, ParserFinder
from dotenv import load_dotenv
import argparse
import threading
import json
import os

load_dotenv()

PARSER_FOLDER = os.getenv("PARSER_FOLDER")
if PARSER_FOLDER is None:
    raise ValueError("PARSER_FOLDER environment variable is not set")
if not os.path.exists(PARSER_FOLDER):
    raise ValueError("Expected PARSER_FOLDER to be an existing path")


parser = argparse.ArgumentParser()
parser.add_argument("name", help="Name to give to the output")
parser.add_argument("target",
                    help="""Which parser file(s) to use.
                    You can specify either a file or directory""")
parser.add_argument("input", help="STDOUT of the tools that you want to parse")
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
for parser_file in pf.yield_parsers():
    lua_parser = Parser(parser_file, args.input)
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
        results.append(result)

if len(panicked_files) > 0:
    logging.error('{"panicked_parsers": %s', json.dumps(panicked_files))

print(json.dumps({
    "name": args.name,
    "results": results
}))
