import json
import os
import argparse


parser = argparse.ArgumentParser()
parser.add_argument("luarocks_exe")
args = parser.parse_args()


with open("config.lua.json", "r") as f:
    data: dict[str, dict] = json.load(f)

package_dict: dict[str, str] = {}
for value in data.values():
    package_list: list[dict[str, str]] = value.get("packages")
    for package_info in package_list:
        package, version = list(package_info.keys())[0], list(package_info.values())[0]
        package_dict[package] = version

for package, version in package_dict.items():
    os.system(f"{args.luarocks_exe} install {package} {version}")
