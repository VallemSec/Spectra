from typing import TypedDict

import pymysql
from flask import g

class RuleFormat(TypedDict):
    id: int
    category: str
    explanation: str
    condition: str
    name: str

class Database:
    """
    Class that contains all the static methods for
    interacting with the database.
    """
    @staticmethod
    def fetch_rules(rule_file_name: str) -> list[RuleFormat]:
        output = list()
        conn: pymysql.Connection = g.mariadb_conn
        with conn.cursor() as cursor:
            cursor.execute("""
            SELECT r.* FROM rules r, files f
            WHERE f.file_name = %s AND r.file_id = f.id;
            """, (rule_file_name,))
            results = cursor.fetchall()
            for result in results:
                output.append(
                    RuleFormat(
                        id = result["id"],
                        category = result["category"],
                        explanation = result["explanation"],
                        condition = result["condition"],
                        name = result["name"]
                    ))
        return output

    class KeyStorage:
        """
        Class that contains all the static methods for
        interacting with the key-value storage in the Database.
        """
        @staticmethod
        def set(key: str, value: str) -> None:
            conn: pymysql.Connection = g.mariadb_conn
            with conn.cursor() as cursor:
                cursor.execute("""
                INSERT INTO `key_value` (`key`, `value`)
                VALUES (%s, %s)
                ON DUPLICATE KEY UPDATE `value` = %s;
                """, (key, value, value))
            conn.commit()

        @staticmethod
        def get(key: str) -> str:
            conn: pymysql.Connection = g.mariadb_conn
            with conn.cursor() as cursor:
                cursor.execute("""
                SELECT `value` FROM `key_value`
                WHERE `key` = %s;
                """, (key,))
                result = cursor.fetchone()
                return result["value"] if result else ""
