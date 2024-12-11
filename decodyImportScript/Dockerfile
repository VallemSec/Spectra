FROM python:3.12.6-alpine3.20

WORKDIR /app
COPY . /app

EXPOSE 5001

RUN pip install --no-cache-dir -r requirements.txt

CMD ["python", "import_rules.py", "/config"]
