function(input)
    local output = {}

    for line in input:gmatch("([^\n]*)\n?") do
        if #line > 0 then  -- Ensure the line is not empty
            table.insert(output, {short = "random_err", long = line})
        end
    end

    return output
end
