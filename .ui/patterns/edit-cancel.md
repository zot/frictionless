---
name: Edit/Cancel Pattern
description: Save state before editing, restore on cancel
---

# Edit/Cancel Pattern

Save a snapshot of editable fields before opening an editor. On save, discard the snapshot. On cancel, restore from the snapshot.

## Lua Implementation

```lua
function Item:openEditor()
    self._snapshot = { name = self.name, email = self.email }
    self.editing = true
end

function Item:save()
    self._snapshot = nil
    self.editing = false
end

function Item:cancel()
    if self._snapshot then
        self.name = self._snapshot.name
        self.email = self._snapshot.email
        self._snapshot = nil
    end
    self.editing = false
end
```

## Key Points

- Use `_` prefix for snapshot field to exclude from serialization
- Copy all editable fields to snapshot
- Always check if snapshot exists before restoring
- Clear snapshot on both save and cancel
