-- Subscribe to the "job-tracker" publisher topic for bookmarklet page captures.
-- When a user clicks the bookmarklet on a job posting, the page content
-- arrives here and gets pushed as a page_received event for Claude to handle.
mcp:subscribe("job-tracker", function(data)
    mcp.pushState({
        app = "job-tracker",
        event = "page_received",
        url = data.url,
        title = data.title,
        text = data.text,
    })
end, {favicon = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNiIgaGVpZ2h0PSIxNiIgZmlsbD0iI0UwN0E0NyIgdmlld0JveD0iMCAwIDE2IDE2Ij4KICA8cGF0aCBkPSJNNi41IDFBMS41IDEuNSAwIDAgMCA1IDIuNVYzSDEuNUExLjUgMS41IDAgMCAwIDAgNC41djhBMS41IDEuNSAwIDAgMCAxLjUgMTRoMTNhMS41IDEuNSAwIDAgMCAxLjUtMS41di04QTEuNSAxLjUgMCAwIDAgMTQuNSAzSDExdi0uNUExLjUgMS41IDAgMCAwIDkuNSAxem0wIDFoM2EuNS41IDAgMCAxIC41LjVWM0g2di0uNWEuNS41IDAgMCAxIC41LS41bTEuODg2IDYuOTE0TDE1IDcuMTUxVjEyLjVhLjUuNSAwIDAgMS0uNS41aC0xM2EuNS41IDAgMCAxLS41LS41VjcuMTVsNi42MTQgMS43NjRhMS41IDEuNSAwIDAgMCAuNzcyIDBNMS41IDRoMTNhLjUuNSAwIDAgMSAuNS41djEuNjE2TDguMTI5IDcuOTQ4YS41LjUgMCAwIDEtLjI1OCAwTDEgNi4xMTZWNC41YS41LjUgMCAwIDEgLjUtLjUiLz4KPC9zdmc+Cg=="})
