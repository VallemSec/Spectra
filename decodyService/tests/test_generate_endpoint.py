import unittest
from unittest import mock
import json

class TestGenerateEndpoint(unittest.TestCase):
    def test_valid_request_id_returns_correct_advice_and_results(self):
        """
        Test the endpoint for a valid request ID to ensure correct advice and results are returned.
        """
        from main import app

        with app.app_context(), app.test_client() as client:
            request_id = "valid_id"
            mock_results = json.dumps([
                {"description": "error1"},
                {"description": "error2"}
            ])
            expected_advice = "error1,error2"

            with mock.patch("helpers.Database.KeyStorage.get", return_value=mock_results):
                response = client.get(f"/generate/{request_id}")
                data = response.get_json()

                self.assertEqual(response.status_code, 200)
                self.assertEqual(data["advice"], expected_advice)
                self.assertEqual(data["results"], json.loads(mock_results))

    def test_request_id_does_not_exist_in_database(self):
        """
        Test the behavior of the endpoint when a request ID that does not exist in the database is provided.
        Ensure that the response status code is 404 and the response data is 'request_id not found'.
        """
        from main import app

        with app.app_context(), app.test_client() as client, \
                mock.patch('helpers.Database.KeyStorage.get', return_value=None):
            request_id = "non_existent_id"
            response = client.get(f"/generate/{request_id}")

            self.assertEqual(response.status_code, 404)
            self.assertEqual(response.data, b"request_id not found")
