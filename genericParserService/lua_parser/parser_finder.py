import os


class ParserFinder:
    def __init__(self, parser_path: str):
        if parser_path.endswith(".lua"):
            self._parsers = [parser_path]
            return
        self._parsers = [os.path.join(parser_path.rstrip("/"), f)
                         for f in os.listdir(parser_path)
                         if os.path.isfile(os.path.join(
                                parser_path.rstrip("/"), f))
                         and f.endswith(".lua")]


    def yield_parsers(self):
        for parser in self._parsers:
            yield parser
