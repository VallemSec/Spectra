# pylint: disable=eval-used
"""
helpers.eval contains all the code required to parse
boolean expressions written in python and return their result.

Example:
    x = 5
    result = helpers.eval.safe_eval("x == 5", {"x": x})
    print(result)
"""
import ast


def safe_eval(expr, variables=None) -> bool:
    """
    This function evaluates a boolean expression using `eval`.
    It is known that `eval` is unsafe, but the use of it here is
    justified since there will be no unauthorized input.
    :param expr: The boolean expression to be evaluated
    :param variables: A dictionary with the variable name as the key
    and the value as value
    :return: The result of the given expression
    """
    try:
        # Parse the expression
        tree = ast.parse(expr, mode="eval")

        # Compile and evaluate the expression safely
        code = compile(tree, "<string>", "eval")
        result = eval(code, {"__builtins__": {}}, variables)
        if not isinstance(result, bool):
            raise ValueError("Evaluated statement is not a boolean")
        return result

    except Exception as e:
        raise ValueError(f"Error evaluating expression: {e}") from e
