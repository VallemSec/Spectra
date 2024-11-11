FROM python:3.12.6-alpine3.20
LABEL authors="NekoLuka"

WORKDIR /app
COPY . /app

EXPOSE 5001
ENV INPUTSCHEMA="/schemas/decody-input.schema.json"

RUN pip install --no-cache-dir -r requirements.txt

CMD ["python", "main.py"]
