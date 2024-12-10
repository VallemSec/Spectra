# Import rules

This script automatically imports the rules from decodyConfig into decody's database.

## environment variables
if your database is not using the default values in the table make sure you pass them as a -e flag if you're using docker or make a .env file.

| Flag             | Description                  | Default      |
|------------------|------------------------------|--------------|
| MARIADB_HOST     | The host of the database     | localhost    |
| MARIADB_PORT     | The port of the database     | 3306         |
| MARIADB_USER     | The user of the database     | root         |
| MARIADB_PASSWORD | The password of the database | password     |
| MARIADB_DATABASE | The name of the database     | decody_devdb |

## Usage Python


```bash
python importRules.py <path to decodyConfig>
```

## Usage Docker
If you use docker for the database as well, make sure to include the database in the same network as the import_script.


```bash
docker build -t import_script .
```

```bash
 docker run -v <path to decodyConfig>:/config import_script
```

if you use docker for the database as well, make sure to include the database in the same network as the import_script.

```bash
docker run --network <network> -v <path to decodyConfig>:/config import_script
```
