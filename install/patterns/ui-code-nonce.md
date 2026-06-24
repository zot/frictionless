---
name: ui-code Nonce
description: Append a transition-bumped nonce to ui-code strings so the engine's change detection sees a new value on every intended fire, guaranteeing the JS executes.
---

# ui-code Nonce

`ui-code` re-fires its bound string as JavaScript **only when the
string value changes**. If the method backing a `ui-code` binding
returns the same JS text twice in a row, the engine treats it as
unchanged and does nothing — the JS never runs.

This is silent: there's no error, the page just doesn't behave.

## Symptom

You have a `ui-code` binding on a Lua method that should fire when
some state transitions:

```html
<span hidden ui-code="iframeBridgeCode()"></span>
```

```lua
function PinnedChunk:iframeBridgeCode()
    if self._editing then return "" end
    return "(function(){ /* mount iframe stuff */ })();"
end
```

On initial render the string is the same as on subsequent ticks,
so the JS never executes. Or it executes once on first mount but
doesn't re-fire when the state should cause re-mount.

## Solution

Embed a **nonce** in the emitted string that changes on every
intended fire. The JS payload stays the same; the comment-prefixed
nonce makes the engine see a fresh value.

### State-bumped nonce

Bump on the transitions that should cause the JS to re-fire:

```lua
Ark.PinnedChunk = session:prototype("Ark.PinnedChunk", {
    _bridgeNonce = 0,
    -- ...
})

function PinnedChunk:bumpBridgeNonce()
    self._bridgeNonce = (self._bridgeNonce or 0) + 1
end

function PinnedChunk:edit()
    -- ... transition logic ...
    self:bumpBridgeNonce()
end

function PinnedChunk:revert()
    -- ... transition logic ...
    self:bumpBridgeNonce()
end

function PinnedChunk:iframeBridgeCode()
    return string.format([[
// iframe-bridge-%d-%d
(function() {
    /* same JS body every time */
})();
]], self:chunkID(), self._bridgeNonce or 0)
end
```

The leading `// ...` comment line is the nonce carrier. JavaScript
ignores it; the engine sees a different string and runs the JS.

### Why a JS comment

Strictly any change anywhere in the string works. A `//` comment is
the cleanest:

- doesn't alter JS semantics
- is easy to grep / inspect
- pairs naturally with a class/id identifier so you can tell which
  card the bridge fired for

## When to apply

Any `ui-code` binding whose JS payload is **idempotent or stable
between transitions** but needs to fire again on a state change.
This shows up most often with:

- iframe mount/scrape/click-shield wiring
- editor mount/destroy via `createInkArkEditor` or similar
- per-card JS setup that depends on Lua-side state

If the JS payload is naturally different every time (e.g., embeds
a timestamp or dynamic content), you don't need an explicit nonce.

## Don't fire on every tick

Resist the temptation to use `os.time()` or a `nowNonce()` that
changes on every variable-check pass. That would re-fire the JS
constantly, including during unrelated state updates — and
re-mounting an iframe/editor every tick is exactly the problem
this whole skill is designed to avoid (see the message-flood
issue documented in personal patterns).

Bump explicitly on the transitions that should cause a fire, and
nowhere else.

## Pair with `?priority=low` when the JS reads sibling attributes

The nonce makes the engine *fire* the JS. But if the JS reads
attributes set by sibling `ui-attr-` bindings on the same parent
element, fire timing matters: variable evaluation is non-
deterministic by default, and the `ui-code` may run before the
`ui-attr-` bindings have applied. The JS would then see a missing
attribute and bail.

Add `?priority=low` to the `ui-code` binding so it evaluates after
the parent's `ui-attr-` bindings:

```html
<div ui-attr-data-cur-chunkid="chunkID()">
  ...
  <span hidden ui-code="iframeBridgeCode()?priority=low"></span>
</div>
```

Without the priority, the JS might run with `data-cur-chunkid`
not yet on the host div, so `document.querySelector('[data-cur-chunkid="X"]')`
returns null and the bridge silently no-ops.

The combination — nonce + low priority — gives you both
guaranteed firing and guaranteed ordering. This is the standard
recipe for any `ui-code` block that reaches outside its own
element to read sibling state.

## Related

- `ui` skill mentions this for `mcp.code`: *"Re-assigning the same
  value is a no-op (change detection); append a nonce to
  re-execute (e.g., `code .. "\n// " .. nonce`)"*. Same mechanism;
  this pattern extends it to per-element `ui-code` bindings driven
  by Lua method returns.
- `ui-basics`: *"ui-code re-fires on page reload. Clear ui-code
  properties when their action is complete to prevent stale
  re-fires."* — the cleanup half of the same issue.
- `js-to-lua-bridge` pattern: uses `?priority=high` on **trigger**
  spans that have side effects. `ui-code` bindings that *consume*
  sibling state want the opposite — low priority so the consumed
  state is settled first.
