from lupa import LuaRuntime
import os

from lua_parser.parser_finder import ParserFinder


class Parser:
    def __init__(self, parser_file: str, parser_input: str, parser_folder: str):
        self._parser_input = parser_input

        self._lua = LuaRuntime(unpack_returned_tuples=True)
        self._lua.eval("function() python = nil; end")()

        with open(os.path.join(parser_folder.rstrip("/"), parser_file),
                  "r", encoding="utf-8") as f:
            self._parser_func = self._lua.eval(f.read())

        self._result = None


    @property
    def result(self):
        return self._result


    def parse(self):
        self._result = self._parse_to_dict(
            self._parser_func(self._parser_input))


    @staticmethod
    def _parse_to_dict(lua_table):
        lua_dict, lua_list = dict(lua_table), list()
        for value in lua_dict.values():
            lua_list.append(dict(value))
        return lua_list
