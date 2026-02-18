---
description: Handle image paste, drag-and-drop, thumbnails, lightbox preview, and file persistence using the JS-to-Lua bridge pattern. Images are base64-decoded to disk for agent access.
---

# Image Paste/Drop/Display

Accept images via clipboard paste or drag-and-drop, show thumbnails, send file paths to the AI agent, and display full-size previews in a lightbox.

## When to Use

- Chat interfaces where users attach images for the AI agent
- Any form that accepts image input via paste or drag-and-drop
- Apps that need to display image thumbnails with lightbox preview

## Architecture

```
Browser (JS)                    Lua                         Disk
-----------                     ---                         ----
paste/drop → FileReader      →  bridge span (updateValue)
             base64 + thumb  →  processImageUpload()     → base64 decode to file
                                ImageAttachment created      storage/uploads/img-*.ext

sendChat() collects paths    →  pushState({images: [...]})  agent reads files
```

Three layers:
1. **JS event handlers** - capture paste/drop, read as base64, generate thumbnail, send via bridge
2. **Lua processing** - decode base64 to disk, create attachment objects for UI display
3. **Viewdef rendering** - thumbnail previews, remove buttons, lightbox overlay

## Prototypes

### ImageAttachment (pending attachment before send)

```lua
MyApp.ImageAttachment = session:prototype("MyApp.ImageAttachment", {
    path = "",          -- absolute path to decoded file on disk
    filename = "",      -- original filename
    thumbnailUri = "",  -- data:image/jpeg;base64,... (small, for preview)
    fullUri = "",       -- data:image/...;base64,... (full size, for lightbox)
    _mcp = EMPTY        -- parent reference for remove callback
})

function ImageAttachment:remove()
    if self._mcp then
        self._mcp:removeAttachment(self)
    end
end
```

### ChatThumbnail (thumbnail in a sent message)

```lua
MyApp.ChatThumbnail = session:prototype("MyApp.ChatThumbnail", {
    uri = "",       -- thumbnail data URI
    fullUri = "",   -- full-size data URI (for lightbox)
    filename = ""
})

function ChatThumbnail:showFull()
    mcp.lightboxUri = self.fullUri ~= "" and self.fullUri or self.uri
end
```

## Lua State & Methods

```lua
-- State properties
mcp._imageAttachments = {}   -- list of ImageAttachment objects
mcp.imageUploadData = ""     -- bridge value (JSON from JS)
mcp.lightboxUri = ""         -- non-empty = lightbox visible
```

### processImageUpload() - Bridge handler

Called automatically via `priority=high` binding when JS sets bridge value.

```lua
function mcp:processImageUpload()
    if self.imageUploadData == "" then return end
    local payload = self.imageUploadData
    self.imageUploadData = ""  -- clear immediately

    local ok, data = pcall(json.decode, payload)
    if not ok or not data then return end

    local filename = data.filename or "image.png"
    local base64Data = data.base64 or ""
    local thumbnailUri = data.thumbnail or ""
    local fullUri = data.fullUri or ""
    if base64Data == "" then return end

    -- Decode base64 to file on disk
    local status = self:status()
    local uploadDir = (status and status.base_dir or "/tmp") .. "/storage/uploads"
    os.execute('mkdir -p "' .. uploadDir .. '"')

    local ext = filename:match("%.(%w+)$") or "png"
    local outPath = uploadDir .. "/img-" .. os.time() .. "-" .. math.random(10000) .. "." .. ext

    local tmpB64 = os.tmpname()
    local f = io.open(tmpB64, "wb")
    if not f then return end
    f:write(base64Data)
    f:close()
    os.execute('base64 -d < "' .. tmpB64 .. '" > "' .. outPath .. '"')
    os.remove(tmpB64)

    table.insert(self._imageAttachments, session:create(ImageAttachment, {
        path = outPath,
        filename = filename,
        thumbnailUri = thumbnailUri,
        fullUri = fullUri,
        _mcp = self
    }))
end
```

### Attachment management

```lua
function mcp:removeAttachment(att)
    for i, a in ipairs(self._imageAttachments) do
        if a == att then
            table.remove(self._imageAttachments, i)
            if att.path ~= "" then os.remove(att.path) end
            break
        end
    end
end

function mcp:clearAttachments()
    for _, att in ipairs(self._imageAttachments) do
        if att.path ~= "" then os.remove(att.path) end
    end
    self._imageAttachments = {}
end
```

### Sending images with chat

```lua
function mcp:sendChat()
    if self.chatInput == "" and #self._imageAttachments == 0 then return end

    -- Convert attachments to ChatThumbnails for message history
    local imagePaths, thumbnails = nil, nil
    if #self._imageAttachments > 0 then
        imagePaths = {}
        thumbnails = {}
        for _, att in ipairs(self._imageAttachments) do
            table.insert(imagePaths, att.path)
            table.insert(thumbnails, session:create(ChatThumbnail, {
                uri = att.thumbnailUri,
                fullUri = att.fullUri,
                filename = att.filename
            }))
        end
    end

    table.insert(self.messages, ChatMessage:new("You", self.chatInput, nil, thumbnails))

    -- Push event with file paths (agent reads files from disk)
    mcp.pushState({
        event = "chat",
        text = self.chatInput,
        images = imagePaths,  -- absolute paths to decoded files
    })

    self.chatInput = ""
    self._imageAttachments = {}  -- clear (files persist for agent)
end
```

## Viewdef: Bridge & Preview Row

```html
<!-- JS-to-Lua bridge for image data -->
<span id="image-bridge" style="display:none" ui-value="imageUploadData?access=rw"></span>
<span style="display:none" ui-value="processImageUpload()?priority=high"></span>

<!-- Pending attachment thumbnails (above chat input) -->
<div class="mcp-image-preview-row" ui-class-hidden="noImages()">
  <div ui-view="imageAttachments()?wrapper=lua.ViewList"></div>
</div>
```

### ImageAttachment list-item viewdef

```html
<template>
  <div class="mcp-image-thumb">
    <img ui-attr-src="thumbnailUri" ui-attr-alt="filename" ui-attr-title="filename">
    <sl-icon-button name="x-circle-fill" label="Remove" ui-event-click="remove()"></sl-icon-button>
  </div>
</template>
```

### ChatThumbnail list-item viewdef (in sent messages)

```html
<template>
  <div class="chat-thumb-item" ui-event-click="showFull()">
    <img ui-attr-src="uri" ui-attr-alt="filename" ui-attr-title="filename">
  </div>
</template>
```

### ChatMessage list-item viewdef (with optional thumbnails)

```html
<template>
  <div class="chat-message" ui-class-user-message="isUser()">
    <span class="chat-message-text" ui-value="text"></span>
    <div class="chat-thumbs" ui-class-hidden="noThumbnails()">
      <div ui-view="chatThumbnails()?wrapper=lua.ViewList"></div>
    </div>
  </div>
</template>
```

## Viewdef: Lightbox

```html
<div id="image-lightbox" class="image-lightbox"
     ui-class-visible="lightboxVisible()"
     ui-event-click="hideLightbox()">
  <img ui-attr-src="lightboxUri" alt="Preview">
</div>
```

With Escape key support:
```html
<script>
document.addEventListener('keydown', (e) => {
    const lb = document.getElementById('image-lightbox');
    if (e.key === 'Escape' && lb && lb.classList.contains('visible')) {
        lb.click();  // triggers hideLightbox() binding
    }
});
</script>
```

## Viewdef: Drag & Drop + Paste Handlers

Place this `<script>` in the DEFAULT viewdef, targeting the panel container.

```html
<script>
(function() {
    const panel = document.getElementById('mcp-chat-panel');
    if (!panel) return;

    let dragCounter = 0;

    panel.addEventListener('dragenter', (e) => {
        e.preventDefault();
        dragCounter++;
        panel.classList.add('drag-over');
    });

    panel.addEventListener('dragleave', () => {
        if (--dragCounter <= 0) {
            dragCounter = 0;
            panel.classList.remove('drag-over');
        }
    });

    panel.addEventListener('dragover', (e) => e.preventDefault());

    panel.addEventListener('drop', (e) => {
        e.preventDefault();
        dragCounter = 0;
        panel.classList.remove('drag-over');
        if (e.dataTransfer && e.dataTransfer.files) {
            handleImageFiles(e.dataTransfer.files);
        }
    });

    panel.addEventListener('paste', (e) => {
        const items = e.clipboardData && e.clipboardData.items;
        if (!items) return;
        const files = [];
        for (const item of items) {
            if (item.type.startsWith('image/')) {
                const file = item.getAsFile();
                if (file) files.push(file);
            }
        }
        if (files.length > 0) handleImageFiles(files);
    });

    function handleImageFiles(files) {
        for (const file of files) {
            if (!file.type.startsWith('image/')) continue;
            const reader = new FileReader();
            reader.onload = (e) => {
                const dataUri = e.target.result;
                const base64Data = dataUri.split(',')[1];
                generateThumbnail(dataUri, 150).then((thumbnail) => {
                    const payload = JSON.stringify({
                        filename: file.name || 'image.png',
                        mime: file.type,
                        base64: base64Data,
                        thumbnail: thumbnail,
                        fullUri: dataUri
                    });
                    const tryUpdate = () => {
                        if (window.uiApp) {
                            window.uiApp.updateValue('image-bridge', payload);
                        } else {
                            setTimeout(tryUpdate, 50);
                        }
                    };
                    tryUpdate();
                });
            };
            reader.readAsDataURL(file);
        }
    }

    function generateThumbnail(dataUri, maxDim) {
        return new Promise((resolve) => {
            const img = new Image();
            img.onload = () => {
                const canvas = document.createElement('canvas');
                let w = img.width, h = img.height;
                if (w > maxDim || h > maxDim) {
                    if (w > h) { h = Math.round(h * maxDim / w); w = maxDim; }
                    else { w = Math.round(w * maxDim / h); h = maxDim; }
                }
                canvas.width = w;
                canvas.height = h;
                canvas.getContext('2d').drawImage(img, 0, 0, w, h);
                resolve(canvas.toDataURL('image/jpeg', 0.7));
            };
            img.onerror = () => resolve('');
            img.src = dataUri;
        });
    }
})();
</script>
```

## CSS

```css
/* Drop zone overlay */
.my-panel.drag-over::after {
    content: 'Drop image here';
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(224, 122, 71, 0.1);
    border: 2px dashed var(--term-accent);
    border-radius: 6px;
    color: var(--term-accent);
    font-family: var(--term-mono);
    font-size: 14px;
    font-weight: 600;
    z-index: 100;
    pointer-events: none;
}

/* Panel needs position: relative for the overlay */
.my-panel { position: relative; }

/* Pending attachment thumbnails */
.image-preview-row {
    display: flex;
    gap: 8px;
    padding-bottom: 8px;
    flex-wrap: wrap;
}

.image-thumb {
    position: relative;
    display: inline-flex;
    border: 1px solid var(--term-accent);
    border-radius: 4px;
    overflow: hidden;
}

.image-thumb img {
    display: block;
    max-height: 60px;
    max-width: 100px;
    object-fit: contain;
}

.image-thumb sl-icon-button {
    position: absolute;
    top: -2px;
    right: -2px;
    font-size: 0.7rem;
    color: var(--term-danger);
    background: var(--term-bg-elevated);
    border-radius: 50%;
}

/* Chat thumbnails (in sent messages) */
.chat-thumbs {
    margin-top: 6px;
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.chat-thumb-item img {
    max-width: 150px;
    max-height: 150px;
    border-radius: 4px;
    border: 1px solid var(--term-border);
    cursor: pointer;
}

/* Lightbox */
.image-lightbox {
    display: none;
    position: fixed;
    inset: 0;
    z-index: 9999;
    background: rgba(0, 0, 0, 0.8);
    align-items: center;
    justify-content: center;
    cursor: pointer;
}

.image-lightbox.visible { display: flex; }

.image-lightbox img {
    max-width: 90vw;
    max-height: 90vh;
    border-radius: 6px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.5);
    object-fit: contain;
}
```

## Data Flow Summary

| Step | Layer | What Happens |
|------|-------|--------------|
| 1. User pastes/drops image | JS | `paste`/`drop` event fires |
| 2. Read as data URI | JS | `FileReader.readAsDataURL()` |
| 3. Generate thumbnail | JS | Canvas resize to 150px, JPEG 0.7 quality |
| 4. Send JSON payload | JS | `uiApp.updateValue('image-bridge', json)` |
| 5. Bridge triggers Lua | Engine | `imageUploadData` set, `processImageUpload()` called (priority=high) |
| 6. Decode to disk | Lua | base64 written to temp file, `base64 -d` to output path |
| 7. Create ImageAttachment | Lua | Added to `_imageAttachments`, UI updates with thumbnail |
| 8. User clicks Send | Lua | File paths collected, pushed in event, attachments cleared |
| 9. Agent processes | Agent | Reads image files from `storage/uploads/` paths |

## Key Points

- **Thumbnail is a data URI** - no server round-trip for preview, generated client-side via Canvas
- **Full image decoded to disk** - agent needs file paths, not data URIs
- **Files persist after send** - `_imageAttachments` is cleared but files stay for agent access
- **Bridge uses JSON** - single payload with filename, mime, base64, thumbnail, and fullUri
- **Two thumbnail contexts**: `ImageAttachment` (pending, with remove button) and `ChatThumbnail` (in sent message history, clickable for lightbox)
- **Lightbox is global** - single `lightboxUri` property on mcp; set non-empty to show, empty to hide

## See Also

- `js-to-lua-bridge.md` - The underlying bridge pattern used here
- `lua-json.md` - JSON encode/decode for the bridge payload
