-- main.lua - entry point for ui-mcp Lua environment
-- mcp global is created by handleStart in Go AFTER this runs
-- mcp.lua (if present) is loaded by Go after creating the mcp global

-- session.metaTostring - enables Lua's tostring() to work with prototype methods
-- Checks if obj has a "tostring" property (direct or inherited) that's a function
-- If so, calls obj:tostring(); otherwise returns the object's type
function session.metaTostring(obj)
    local tostr = obj.tostring
    if type(tostr) == "function" then
        return tostr(obj)
    end
    return rawget(obj, "type") or "object"
end

-- Wrap session:prototype() to set __tostring on every prototype
-- GopherLua doesn't inherit __tostring through metatables properly
local originalPrototype = session.prototype
function session:prototype(name, init, parent)
    local proto = originalPrototype(self, name, init, parent)
    proto.__tostring = session.metaTostring
    return proto
end

-- Object prototype - default base for all prototypes
-- Provides common methods inherited by all objects
Object = session:prototype('Object', {})

-- Returns "a <Type>" or "an <Type>" with grammatically correct article
-- Examples: "a Person", "an Object", "an Item", "a Contact"
function Object:tostring()
    local t = self.type or "Object"
    if rawget(self, "type") ~= nil then
       return t
    end
    local first = t:sub(1, 1):lower()
    local article = first:match("[aeiou]") and "an" or "a"
    return article .. " " .. t
end
