FROM python:3.12.6-alpine3.20
LABEL authors="NekoLuka"

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt && apk add nmap

COPY . .

ENTRYPOINT [ "python", "./main.py" ]
