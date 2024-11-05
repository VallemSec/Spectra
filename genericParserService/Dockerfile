FROM python:3.12.6-alpine3.20
LABEL authors="NekoLuka"

ARG LUA_CONFIG_FILE="{}"
ARG LUAROCKS_EXE="luarocks-5.4"

WORKDIR /app

COPY requirements.txt install_lua_packages.py ./
RUN pip install --no-cache-dir -r requirements.txt &&  \
    apk add lua luarocks && \
    echo "${LUA_CONFIG_FILE}" > config.lua.json && \
    python install_lua_packages.py ${LUAROCKS_EXE}

COPY . .

ENTRYPOINT ["python", "/app/main.py"]
