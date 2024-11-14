# Orchestrator service

## Features
- Organize services and kickoff order


```bash
docker build -t orchestrator_service .
```

```bash
docker run -p 8080:8080 -e DOCKER_RUNNER_SERVICE_URL="http://localhost:8008" -e CONFIG_FILE_PATH="config.yaml" -e PARSERS_FOLDER="/SpectraConfig/parsers" -e PARSER_IMAGE="nekoluka/spectra-scanner" -e PARSER_TAG="1.0.1" -e DECODY_SERVICE="http://localhost:5001" orchestrator_service
```
