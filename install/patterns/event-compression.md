# Event Compression Pattern

Compress high-frequency events (drag, scroll, resize) using a render loop. The event handler only stores state; a `requestAnimationFrame` loop applies changes when the state has changed, synced to the display refresh.

## How It Works

1. On start (e.g. mousedown): capture reference state and start a RAF loop
2. On each event (e.g. mousemove): store the latest value (e.g. mouse Y) — do no other work
3. The RAF loop checks if the value changed since last apply; if so, computes and applies the update
4. On end (e.g. mouseup): cancel the RAF loop, clean up

Capturing reference state at the start (not on each event) eliminates error propagation — every update is computed from the original reference point, not accumulated deltas.

## Template

```javascript
(function() {
  let startVal, currentVal, lastVal, rafId;

  element.addEventListener('mousedown', (e) => {
    startVal = currentVal = lastVal = e.clientY;
    loop();
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onRelease);
    e.preventDefault();
  });

  function onMove(e) {
    currentVal = e.clientY;  // Just store — no work
  }

  function loop() {
    rafId = requestAnimationFrame(() => {
      if (currentVal !== lastVal) {
        lastVal = currentVal;
        doWork(startVal, currentVal);
      }
      loop();
    });
  }

  function onRelease() {
    cancelAnimationFrame(rafId);
    document.removeEventListener('mousemove', onMove);
    document.removeEventListener('mouseup', onRelease);
  }
})();
```

## Example: Drag Resize Handle

```javascript
(function() {
  const panel = document.getElementById('my-panel');
  const handle = document.getElementById('my-handle');
  if (!panel || !handle) return;

  let startY, startHeight, currentY, lastY, rafId;

  handle.addEventListener('mousedown', (e) => {
    startY = currentY = lastY = e.clientY;
    startHeight = panel.offsetHeight;
    panel.querySelectorAll('iframe').forEach(f => f.style.pointerEvents = 'none');
    document.addEventListener('mousemove', onDrag);
    document.addEventListener('mouseup', onRelease);
    rafLoop();
    e.preventDefault();
  });

  function onDrag(e) {
    currentY = e.clientY;
  }

  function rafLoop() {
    rafId = requestAnimationFrame(() => {
      if (currentY !== lastY) {
        lastY = currentY;
        const height = Math.max(120, Math.min(window.innerHeight * 0.6, startHeight + (startY - currentY)));
        panel.style.height = height + 'px';
      }
      rafLoop();
    });
  }

  function onRelease() {
    cancelAnimationFrame(rafId);
    document.removeEventListener('mousemove', onDrag);
    document.removeEventListener('mouseup', onRelease);
    panel.querySelectorAll('iframe').forEach(f => f.style.pointerEvents = '');
  }
})();
```
