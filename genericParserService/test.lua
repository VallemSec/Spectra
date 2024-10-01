function(input)
    output = {}

    for line in input:gmatch("([^\n]*)\n?") do
        if #line > 0 then  -- Ensure the line is not empty
            table.insert(output, {err_short = "random_err", err_long = line})
        end
    end

    return output
end
