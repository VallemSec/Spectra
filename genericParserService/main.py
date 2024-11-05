from lua_parser import Parser, ParserFinder
from dotenv import load_dotenv
import argparse
import threading
import json
import os

load_dotenv()

PARSER_FOLDER = os.getenv('PARSER_FOLDER')
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

args = parser.parse_args()

pf = ParserFinder(args.target, PARSER_FOLDER)

lua_parser_threads = []
for parser_file in pf.yield_parsers():
    lua_parser = Parser(parser_file, args.input, PARSER_FOLDER)
    thread = threading.Thread(target=lua_parser.parse)
    lua_parser_threads.append({"thread": thread, "parser": lua_parser})
    thread.start()

thread_active = [True for _ in range(len(lua_parser_threads))]
while True in thread_active:
    for i, thread in enumerate(lua_parser_threads):
        if not thread["thread"].is_alive():
            thread_active[i] = False

results = []
for thread in lua_parser_threads:
    results += thread["parser"].result

print(json.dumps({
    "name": args.name,
    "results": results
}))
