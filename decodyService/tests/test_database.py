import unittest

import pymysql

from helpers import Database


class DatabaseTest(unittest.TestCase):
    def test_open_db(self):
        self.assertIsInstance(Database.db_connect(), pymysql.Connection)

    def test_fail_open_db(self):
        with self.assertRaises(pymysql.err.OperationalError):
            Database.db_connect(port=99999)

    def test_set_key(self):
        conn = Database.db_connect()
        self.assertEqual(Database.KeyStorage.set("testing_set", "testing", conn), 1)
        with conn.cursor() as cursor:
            cursor.execute("DELETE FROM `key_value` WHERE `key` = 'testing_set'")
        conn.commit()

    def test_get_key(self):
        conn = Database.db_connect()
        with conn.cursor() as cursor:
            cursor.execute("INSERT INTO `key_value` (`key`, `value`) VALUES('testing_get', 'testing')")
        self.assertEqual(Database.KeyStorage.get("testing_get", conn), "testing")
        with conn.cursor() as cursor:
            cursor.execute("DELETE FROM `key_value` WHERE `key` = 'testing_get'")
        conn.commit()
