from typing import TypedDict


class ResultObjectFormat(TypedDict):
    short: str
    long: str


class LoadEndpointInputFormat(TypedDict):
    scanner_name: str
    rules: list[str]
    results: list[ResultObjectFormat]


class DecodyDatabaseRuleFormat(TypedDict):
    id: int
    category: str
    explanation: str
    condition: str
    name: str


class DecodyDatabaseResultFormat(TypedDict):
    category: str
    rule_name: str
    rule_explanation: str
    scanner_name: str
    long_input: str
    short_input: str


class DecodyFindingsOutputFormat(TypedDict):
    rule_name: str
    rule_explanation: str
    scanner_name: str


class DecodyCategoryOutputFormat(TypedDict):
    category: str
    ai_advice: str
    results: list[DecodyFindingsOutputFormat]


class DecodyOutputFormat(TypedDict):
    ai_advice: str
    results: dict[str, DecodyCategoryOutputFormat]


class DecodyPromptFormat(TypedDict):
    category_prompt: str
    summary_prompt: str
