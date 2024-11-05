from typing import TypedDict


class ResultObjectFormat(TypedDict):
    err_short: str
    err_long: str

class LoadEndpointInputFormat(TypedDict):
    name: str
    rules: list[str]
    results: list[ResultObjectFormat]

class DecodyDatabaseRuleFormat(TypedDict):
    id: int
    category: str
    explanation: str
    condition: str
    name: str
