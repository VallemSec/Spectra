import sys
import os
import pymysql
import yaml
import dotenv


def get_rules_from_file(file: str) -> list:
    try:
        with open(file, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
            return data.get('rules', [])
    except FileNotFoundError:
        return []


def db_connect(host: str = None,
               port: int = None,
               user: str = None,
               password: str = None,
               db: str = None):
    return pymysql.Connect(
        host=host or os.getenv("MARIADB_HOST") or "localhost",
        port=port or int(os.getenv("MARIADB_PORT", 0)) or 3306,
        user=user or os.getenv("MARIADB_USER") or "root",
        password=password or os.getenv("MARIADB_PASSWORD") or "password",
        database=db or os.getenv("MARIADB_DATABASE") or "decody_devdb",
        cursorclass=pymysql.cursors.DictCursor
    )


dotenv.load_dotenv()

parser_folder = sys.argv[1] if len(sys.argv) > 1 else None

if parser_folder is None:
    print("No folder provided")
    sys.exit(1)

if not os.path.exists(parser_folder):
    print("Folder does not exist")
    sys.exit(1)

print("Folder found")

parser_files = set()
# The parser folder is found, now we can start parsing the files
# and extracting the necessary information
print(f"Searching for yaml files in {parser_folder}")
for file in os.listdir(parser_folder):
    if os.path.isfile(os.path.join(parser_folder, file)):
        # if the file is a yaml file, we can parse it
        if file.endswith(".yaml"):
            print(f"Found: {file}")
            parser_files.add(file)
    elif os.path.isdir(os.path.join(parser_folder, file)):
        # if the file is a directory, we can recursively search for yaml files
        for root, _, files in os.walk(os.path.join(parser_folder, file)):
            for file in files:
                if file.endswith(".yaml"):
                    print(f"Found: " + os.path.relpath(os.path.join(root, file), parser_folder))
                    # add the path from the parser folder to the file
                    parser_files.add(os.path.relpath(os.path.join(root, file), parser_folder))

try:
    conn = db_connect()
except pymysql.err.OperationalError:
    print("Failed to open connection to maria db")
    sys.exit(1)

try:
    with conn.cursor() as cursor:
        for file in parser_files:
            cursor.execute("""
            SELECT * FROM files WHERE file_name = %s;
            """, (file,))

            if cursor.rowcount == 0:
                cursor.execute("""
                INSERT INTO files (file_name) VALUES (%s)ON DUPLICATE KEY UPDATE file_name = file_name;
                """, (file,))

            rules = get_rules_from_file(os.path.join(parser_folder, file))
            for rule in rules:
                cursor.execute("""
                SELECT * FROM rules
                WHERE category = %s AND explanation = %s AND `condition` = %s AND name = %s AND file_id = (SELECT id FROM files WHERE file_name = %s);
                """, (rule["category"], rule["explanation"], rule["condition"], rule["name"], file))

                if cursor.rowcount == 0:
                    cursor.execute("""
                    INSERT INTO rules (category, explanation, `condition`, name, file_id)
                    VALUES (%s, %s, %s, %s, (SELECT id FROM files WHERE file_name = %s));
                    """, (rule["category"], rule["explanation"], rule["condition"], rule["name"], file))

        conn.commit()
except Exception as e:
    print(f"An error occurred: {e}")
    conn.rollback()
finally:
    conn.close()
