import os

import unittest
from unittest import mock

class TestLoadEndpoint(unittest.TestCase):
    @mock.patch.dict(os.environ, {"INPUTSCHEMA": "../../jsonSchemas/decody-input.schema.json"})
    def test_valid_json_body_processed_correctly(self):
        """
        Test the processing of a valid JSON body by the endpoint.
        Mock the environment variables and database calls for testing.
        Send a POST request to the endpoint with a valid JSON body.
        Check if the response status code is 201 and certain database calls are made.
        """
        from main import app
        with (app.app_context(),
              mock.patch('helpers.Database.fetch_rules', return_value=[{
                  "condition": "err_short == 'error1'",
                  "category": "cat1",
                  "explanation": "exp1",
                  "name": "name1"
              }]),
              mock.patch('helpers.Database.KeyStorage.set') as mock_set,
              mock.patch('helpers.Database.KeyStorage.get', return_value='[]')):

            with app.test_client() as client:
                request_id = "123"
                response = client.post(f"/load/{request_id}", json={
                    "name": "test",
                    "rules": ["rule1.json"],
                    "results": [{"err_short": "error1"}]
                })
            self.assertEqual(response.status_code, 201)
            mock_set.assert_any_call(f"{request_id}-input", mock.ANY)
            mock_set.assert_any_call(f"{request_id}-results", mock.ANY)

    @mock.patch.dict(os.environ, {"INPUTSCHEMA": "../../jsonSchemas/decody-input.schema.json"})
    def test_request_body_not_json_returns_400(self):
        """
        Test the behavior when the request body is not in JSON format.
        Mock the environment variables for testing.
        Send a POST request to the endpoint with an empty body.
        Check if the response data is 'Body not JSON' and the status code is 400.
        """
        from main import app
        with app.app_context():
            with app.test_client() as client:
                request_id = "123"
                response = client.post(f"/load/{request_id}", data="")
            self.assertEqual(response.data, b"Body not JSON")
            self.assertEqual(response.status_code, 400)

    @mock.patch.dict(os.environ, {"INPUTSCHEMA": "../../jsonSchemas/decody-input.schema.json"})
    def test_request_body_invalid_json_returns_400(self):
        """
        Test the behavior when the request body is an empty JSON object.
        Mock the environment variables for testing.
        Send a POST request to the endpoint with an empty JSON body.
        Check if the response data is 'Body not properly formatted' and the status code is 400.
        """
        from main import app
        with app.app_context():
            with app.test_client() as client:
                request_id = "123"
                response = client.post(f"/load/{request_id}", json={})
            self.assertEqual(response.data, b"Body not properly formatted")
            self.assertEqual(response.status_code, 400)
