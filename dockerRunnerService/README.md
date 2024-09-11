# Docker compose (Easiest)

```shell
docker compose up -d
```

if an update came out you might need to rebuild the image

```shell
docker compose up -d --build
```

# Docker run
### How to build

```shell
docker build -t dockerrunnerapi .
```

### How to run

```shell
docker run -p 127.0.0.1:8008:8080 -v /var/run/docker.sock:/var/run/docker.sock dockerrunnerapi
```


the API will be accessible on port 8008
This is an example JSON to send to start subfinder and scan hu.nl

```json
{
	"containerName": "projectdiscovery/subfinder",
	"containerCommand":[
		"-d", "hu.nl"
	]
}
```