FROM python:3.12.6-alpine3.20
LABEL authors="NekoLuka"

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt && apk add lua

COPY . .

ENTRYPOINT ["python", "/app/main.py"]
