from lupa import LuaRuntime


def parse_to_dict(lua_table):
    lua_dict, lua_list = dict(lua_table), list()
    for value in lua_dict.values():
        lua_list.append(dict(value))
    return lua_list


# Load the runtime and disable access to python objects
lua = LuaRuntime(unpack_returned_tuples=True)
lua.eval("function() python = nil; end")()

with open("test.lua", "r") as f:
    parser_func = lua.eval(f.read())

# print(lua_table_to_dict(parser_func(["test", "a", "b"], "test")))
print(parse_to_dict(parser_func("hello world\nBye world\nTesting 123")))