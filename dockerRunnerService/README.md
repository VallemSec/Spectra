# The api
the API will be accessible on port 8008
This is an example JSON to send to start the latest version of subfinder and scan vallem.com.

```json
{
    "containerName": "projectdiscovery/subfinder",
    "containerTag": "latest",
    "containerCommand":[
      "-d", "vallem.com"
    ],
    "volume":[
        "C:\\absolute/path/to/folder:/internal/folder"
    ],
    "env":[
        "PARSER_FOLDER=/parsers"
    ]
}
```

It's best practice to define a version of the container you want to run ex 6.0.0 instead of latest so the scraper doesn't break if the container is updated.

## Docker compose (Easiest)

```shell
docker compose up -d
```

if an update came out you might need to rebuild the image

```shell
docker compose up -d --build
```

## Docker run
### How to build

```shell
docker build -t dockerrunnerapi .
```

### How to run

```shell
docker run -p 127.0.0.1:8008:8080 -v /var/run/docker.sock:/var/run/docker.sock dockerrunnerapi
```
