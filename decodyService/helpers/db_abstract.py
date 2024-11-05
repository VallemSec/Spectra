from flask import g

import pymysql
import os

from helpers.types import DecodyDatabaseRuleFormat


class Database:
    """
    Class that contains all the static methods for
    interacting with the database.
    """
    @staticmethod
    def db_connect(host: str = None,
                   port: int = None,
                   user: str = None,
                   password: str = None,
                   db: str = None):
        return pymysql.Connect(
            host=host or os.getenv("MARIADB_HOST") or "localhost",
            port=port or int(os.getenv("MARIADB_PORT", 0)) or 3306,
            user=user or os.getenv("MARIADB_USER") or "root",
            password=password or os.getenv("MARIADB_PASSWORD") or "password",
            database=db or os.getenv("MARIADB_DATABASE") or "decody_devdb",
            cursorclass=pymysql.cursors.DictCursor
        )

    @staticmethod
    def fetch_rules(rule_file_name: str, connection: pymysql.Connection = None) -> list[DecodyDatabaseRuleFormat]:
        output = list()
        conn: pymysql.Connection = connection or g.mariadb_conn
        with conn.cursor() as cursor:
            cursor.execute("""
            SELECT r.id, r.category, r.explanation, r.`condition`, r.name 
            FROM rules r, files f
            WHERE f.file_name = %s AND r.file_id = f.id;
            """, (rule_file_name,))
            results = cursor.fetchall()
            for result in results:
                output.append(
                    DecodyDatabaseRuleFormat(
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
        def set(key: str, value: str, connection: pymysql.Connection = None) -> int:
            conn: pymysql.Connection = connection or g.mariadb_conn
            with conn.cursor() as cursor:
                affected_rows = cursor.execute("""
                INSERT INTO `key_value` (`key`, `value`)
                VALUES (%s, %s)
                ON DUPLICATE KEY UPDATE `value` = %s;
                """, (key, value, value))
            conn.commit()
            return affected_rows

        @staticmethod
        def get(key: str, connection: pymysql.Connection = None) -> str | None:
            conn: pymysql.Connection = connection or g.mariadb_conn
            with conn.cursor() as cursor:
                cursor.execute("""
                SELECT `value` FROM `key_value`
                WHERE `key` = %s;
                """, (key,))
                result = cursor.fetchone()
                return result["value"] if result else None
