from typing import Any
import re

import pymysql.cursors

def convert_sql_str_to_connect_obj(sql_str: str) -> dict[str, Any]:
    # Regex matches the pattern username:password@host:port/database
    pattern = re.compile(r"([\w]+):([\w\d!#$%^&]+)@([\w\d\.-_]+):(\d{1,5})/([\w_-]+)")
    match = pattern.fullmatch(sql_str)
    groups = match.groups()
    return {
        "user": groups[0],
        "password": groups[1],
        "host": groups[2],
        "port": int(groups[3]),
        "database": groups[4],
        "cursorclass": pymysql.cursors.DictCursor
    }
