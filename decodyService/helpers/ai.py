class AI:
    def __init__(self):
        # Define some initialisation for AI methods
        pass

    def generate_ai_advice(self, error: str) -> str:
        # Generate an ELIA5 description for a single error
        return error

    def generate_complete_ai_advice(self, errors: list[str]) -> str:
        # Generate an ELIA5 report for everything in general
        return ",".join(errors)
