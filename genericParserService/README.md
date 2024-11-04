# Generic parsing service

## Features
- Dynamically load parsing scripts for inputted service

## Usage
Write a lua file with a function called parse, accepting an array of strings.
It needs to return an object according to specification.

## Build
`docker build -t parser --build-arg LUA_CONFIG_FILE="$(cat ../../SpectraConfig/parsers/config.lua.json)" .`  
This is assuming you have installed the config git repo next to this one.