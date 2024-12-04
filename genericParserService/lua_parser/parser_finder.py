import os


class ParserFinder:
    """
    A class to locate and yield Lua parser files from a specified directory.

    Methods:
        yield_parsers(): Yields each parser file path in the _parsers list.
    """
    def __init__(self, parser_path: str, parser_folder: str):
        absolute_parser_path = os.path.join(
            parser_folder.rstrip("/"), parser_path)
        if parser_path.endswith(".lua"):
            self._parsers = [absolute_parser_path]
            return
        self._parsers = [os.path.join(absolute_parser_path.rstrip("/"), f)
                         for f in os.listdir(os.path.join(parser_folder,
                                                          absolute_parser_path))
                         if os.path.isfile(os.path.join(
                            absolute_parser_path.rstrip("/"), f))
                         and f.endswith(".lua")]


    def yield_parsers(self):
        for parser in self._parsers:
            yield parser
