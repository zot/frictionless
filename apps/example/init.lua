-- Subscribe to the "example" publisher topic for bookmarklet page captures.
-- When a user clicks the bookmarklet on a page, the content
-- arrives here and gets pushed as a page_received event for Claude to handle.
mcp:subscribe("example", function(data)
    mcp.pushState({
        app = "example",
        event = "page_received",
        url = data.url,
        title = data.title,
        text = data.text,
    })
end, {favicon = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNiIgaGVpZ2h0PSIxNiIgZmlsbD0iI0UwN0E0NyIgdmlld0JveD0iMCAwIDE2IDE2Ij4KICA8cGF0aCBkPSJNMTQuMzU0IDMuMzU0YTIgMiAwIDAgMC0yLjgzLTIuODNMMy43MDcgOC4zNDNhLjUuNSAwIDAgMCAwIC43MDdsMy41MzUgMy41MzZhLjUuNSAwIDAgMCAuNzA3IDBsMTAuODM1LTEwLjgzNnpNMy41IDhoLjc5M2wxLjc5NCAxLjc5NUwzIDEyLjg3OVY5LjVBMS41IDEuNSAwIDAgMSAzLjUgOCIvPgo8L3N2Zz4K"})
