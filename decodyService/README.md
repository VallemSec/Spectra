# Decody service

Dcodey collects all parsed outputs from the ran services and
creates a rapport that is understandable for people with little technical knowledge

## Features
- Collect results from scans
- Put results together in a rapport
- if no results are found, ask chatGPT to generate a description of the vulnerability

```bash
docker build -t decody .
```

```bash
docker run -p 5001:5001 -v <path to spectra repo>/jsonSchemas:/schemas decody
```
