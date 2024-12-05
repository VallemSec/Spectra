import json
import logging
import sys
import threading

from lupa import LuaRuntime, LuaError
import os

from lua_parser.parser_finder import ParserFinder


class Parser:
    def __init__(self, parser_file: str, parser_input: str):
        """
        Initializes the Parser instance with the specified Lua parser file, input, and folder.

        Sets up the Lua runtime environment, configures error and warning handlers, and attempts
        to load the Lua parser function from the specified file. If loading fails, logs the error
        and sets the panic state.

        Args:
            parser_file (str): The name of the Lua parser file.
            parser_input (str): The input data to be parsed by the Lua parser.
        """
        self._panic_logger = logging.getLogger("lua_panic_reporter")

        self.parser_file = parser_file
        self._parser_input = parser_input

        self._lua = LuaRuntime(unpack_returned_tuples=True)
        self._lua.eval("function() python = nil; end")()
        self._lua.globals().report_error = self._report_lua_parser_error
        self._lua.globals().report_warning = self._report_lua_parser_warning
        self._lua.globals().panic = self._lua_panic

        self._result = None
        self._panic = False
        self._thread_id = None

        with open(self.parser_file,
                  "r", encoding="utf-8") as f:
            try:
                self._parser_func = self._lua.eval(f.read())
            except LuaError as e:
                self._panic = True
                self._panic_logger.error("Could not load parser with the following error: %s", e)


    @property
    def result(self):
        return self._result


    @property
    def panicked(self):
        return self._panic


    def parse(self):
        """
        Executes the Lua parser function with the provided input.

        If the parser is in a panic state, logs an error and exits without parsing.
        Attempts to parse the input using the Lua function loaded during initialization.
        If a LuaError occurs during parsing, sets the panic state and logs the error.
        Checks for a panic signal in the environment variables given from an in-script panic.
        Converts the Lua parsing result to a dictionary format and stores it in the result attribute.
        """
        if self._panic:
            self._panic_logger.error("Not parsing %s, already panicked", self.parser_file)
            return

        self._thread_id = threading.current_thread().native_id

        try:
            result = self._parser_func(self._parser_input)
        except LuaError as e:
            self._panic = True
            self._panic_logger.error("Parser crashed with the following error: %s", e)
            return

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
        """
        Handles a Lua panic event by setting an environment variable and logging the panic details.

        This method sets an environment variable specific to the current thread to indicate a panic state.
        It logs the panic details using a custom logger configured to output JSON-formatted messages
        to standard error.

        Args:
            *args: Variable length argument list containing panic details to be logged.
        """
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
