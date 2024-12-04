import json
import logging
import sys
import threading

from lupa import LuaRuntime
import os

from lua_parser.parser_finder import ParserFinder


class Parser:
    def __init__(self, parser_file: str, parser_input: str, parser_folder: str):
        self.parser_file = parser_file
        self._parser_input = parser_input

        self._lua = LuaRuntime(unpack_returned_tuples=True)
        self._lua.eval("function() python = nil; end")()
        self._lua.globals().report_error = self._report_lua_parser_error
        self._lua.globals().report_warning = self._report_lua_parser_warning
        self._lua.globals().panic = self._lua_panic

        with open(os.path.join(parser_folder.rstrip("/"), self.parser_file),
                  "r", encoding="utf-8") as f:
            self._parser_func = self._lua.eval(f.read())

        self._result = None
        self._panic = False
        self._thread_id = None

    @property
    def result(self):
        return self._result

    @property
    def panicked(self):
        return self._panic

    def parse(self):
        self._thread_id = threading.current_thread().native_id
        result = self._parser_func(self._parser_input)
        if os.getenv(f"_PARSER_LUA_PANIC_{self._thread_id}", "0") == "1":
            self._panic = True
            return
        self._result = self._parse_to_dict(result)

    def cleanup(self):
        if f"_PARSER_LUA_PANIC_{self._thread_id}" in os.environ.keys():
            os.environ.pop(f"_PARSER_LUA_PANIC_{self._thread_id}")

    @staticmethod
    def _parse_to_dict(lua_table):
        lua_dict, lua_list = dict(lua_table), list()
        for value in lua_dict.values():
            lua_list.append(dict(value))
        return lua_list

    @staticmethod
    def _report_lua_parser_error(*args) -> None:
        logger = logging.getLogger("lua_error_reporter")
        for arg in args:
            logger.error(arg)

    @staticmethod
    def _report_lua_parser_warning(*args) -> None:
        logger = logging.getLogger("lua_error_reporter")
        for arg in args:
            logger.warning(arg)

    @staticmethod
    def _lua_panic(*args) -> None:
        thread_id = threading.current_thread().native_id
        os.environ[f"_PARSER_LUA_PANIC_{thread_id}"] = "1"
        logger = logging.getLogger("lua_panic_reporter")
        logger.propagate = False
        formatter = logging.Formatter(
            '{"time": "%(asctime)s", "logger_name": "%(name)s", "level": "%(levelname)s", "messages": %(message)s}'
        )
        handler = logging.StreamHandler(stream=sys.stderr)
        handler.setFormatter(formatter)
        logger.addHandler(handler)
        logger.error(json.dumps(args))
