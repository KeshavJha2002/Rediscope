package viewer

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>rediscope RDB viewer</title>
  <style>
    :root {
      --bg: #f6f6f2;
      --surface: #fff;
      --ink: #202322;
      --muted: #656d69;
      --line: #242827;
      --soft: #d9ddd8;
      --key: #79c8ff;
      --value: #c9edbd;
      --neutral: #e7e9e6;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    * { box-sizing: border-box; }
    body { margin: 0; background: var(--bg); color: var(--ink); }
    button { font: inherit; }
    .app { display: grid; grid-template-columns: 230px minmax(360px, 1fr) 320px; height: 100vh; }
    .left, .center, .right { overflow: auto; min-width: 0; padding: 14px; }
    .left { border-right: 2px solid var(--line); background: #fbfbf8; }
    .center { border-right: 2px solid var(--line); background: #fcfcfa; }
    h1 { margin: 0 0 12px; font-size: 18px; }
    h2 { margin: 18px 0 8px; color: var(--muted); text-transform: uppercase; font-size: 12px; letter-spacing: 0; }
    .file-btn, .section-btn {
      width: 100%;
      border: 1px solid var(--soft);
      border-radius: 6px;
      background: var(--surface);
      color: var(--ink);
      cursor: pointer;
      text-align: left;
    }
    .file-btn { padding: 9px; margin-bottom: 7px; }
    .file-btn.active, .section-btn.active { border-color: var(--line); box-shadow: inset 0 0 0 1px var(--line); }
    .muted { color: var(--muted); font-size: 12px; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
    .tabs { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 12px; }
    .tab { border: 1px solid var(--line); border-radius: 6px 6px 0 0; padding: 6px 10px; background: var(--surface); }
    .facts { display: grid; gap: 7px; margin-bottom: 12px; }
    .fact { display: grid; grid-template-columns: 120px 1fr; gap: 10px; border: 1px solid var(--soft); border-radius: 6px; background: var(--surface); padding: 8px; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-size: 12px; }
    .section-list { display: grid; gap: 7px; }
    .section-btn { padding: 8px; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; font-size: 12px; }
    .byte-grid { display: grid; grid-template-columns: repeat(10, 18px); grid-auto-rows: 18px; gap: 4px; width: max-content; }
    .byte { width: 18px; height: 18px; border: 1px solid var(--line); border-radius: 4px; background: var(--neutral); cursor: pointer; }
    .byte.key { background: var(--key); }
    .byte.value { background: var(--value); }
    .byte.neutral { background: var(--neutral); }
    .byte.selected { box-shadow: 0 0 0 2px var(--line); }
    .tooltip { position: fixed; display: none; pointer-events: none; border: 1px solid var(--line); border-radius: 6px; background: #fff; padding: 8px; box-shadow: 0 12px 26px rgba(0,0,0,.16); font-size: 12px; font-family: ui-monospace, SFMono-Regular, Consolas, monospace; white-space: pre; z-index: 10; }
    @media (max-width: 900px) { .app { grid-template-columns: 1fr; height: auto; } .left, .center { border-right: 0; border-bottom: 2px solid var(--line); } }
  </style>
</head>
<body>
  <main class="app">
    <aside class="left">
      <h1>rediscope</h1>
      <div id="files"></div>
    </aside>
    <section class="center">
      <div id="tabs" class="tabs"></div>
      <div id="summary"></div>
      <h2>RDB structure</h2>
      <div id="sections" class="section-list"></div>
    </section>
    <aside class="right">
      <h2>Byte view</h2>
      <div id="bytes" class="byte-grid"></div>
    </aside>
  </main>
  <div id="tooltip" class="tooltip"></div>
  <script>
    const files = {{.Payload}};
    let activeFile = 0;
    let selected = null;
    const els = {
      files: document.getElementById("files"),
      tabs: document.getElementById("tabs"),
      summary: document.getElementById("summary"),
      sections: document.getElementById("sections"),
      bytes: document.getElementById("bytes"),
      tooltip: document.getElementById("tooltip")
    };
    function classFor(kind) {
      if (kind === "key") return "key";
      if (kind === "value") return "value";
      return "neutral";
    }
    function sectionAt(file, offset) {
      return file.sections.find((s) => offset >= s.start && offset < s.end);
    }
    function setActive(index) {
      activeFile = index;
      selected = { start: 0, end: files[index].size, id: "file", source: "file" };
      render();
    }
    function selectSection(id) {
      const section = files[activeFile].sections.find((s) => s.id === id);
      selected = section ? { start: section.start, end: section.end, id: section.id, source: "section" } : null;
      renderCenter();
      renderBytes();
    }
    function render() {
      renderFiles();
      renderCenter();
      renderBytes();
    }
    function renderFiles() {
      els.files.innerHTML = "";
      files.forEach((file, index) => {
        const button = document.createElement("button");
        button.className = "file-btn" + (index === activeFile ? " active" : "");
        button.innerHTML = file.name + "<br><span class='muted'>" + file.size + "B / RDB " + file.version + "</span>";
        button.onclick = () => setActive(index);
        els.files.appendChild(button);
      });
    }
    function renderCenter() {
      const file = files[activeFile];
      els.tabs.innerHTML = "";
      files.forEach((item, index) => {
        const tab = document.createElement("button");
        tab.className = "tab" + (index === activeFile ? " active" : "");
        tab.textContent = item.name;
        tab.onclick = () => setActive(index);
        els.tabs.appendChild(tab);
      });
      els.summary.innerHTML = "<div class='facts'>" +
        "<div class='fact'><span>file</span><span>" + file.name + "</span></div>" +
        "<div class='fact'><span>size</span><span>" + file.size + "B</span></div>" +
        "<div class='fact'><span>version</span><span>" + file.version + "</span></div>" +
        "<div class='fact'><span>keys</span><span>" + file.keys.length + "</span></div>" +
        "</div>";
      els.sections.innerHTML = "";
      file.sections.forEach((section) => {
        const button = document.createElement("button");
        button.className = "section-btn" + (selected && selected.source === "section" && selected.id === section.id ? " active" : "");
        button.innerHTML = section.label + "<br><span class='muted'>" + section.kind + " offset=" + section.start + ".." + section.end + " size=" + section.size + "B</span>";
        button.onclick = () => selectSection(section.id);
        els.sections.appendChild(button);
      });
    }
    function renderBytes() {
      const file = files[activeFile];
      els.bytes.innerHTML = "";
      const byteCount = file.hex.length / 2;
      for (let i = 0; i < byteCount; i++) {
        const hex = file.hex.slice(i * 2, i * 2 + 2);
        const section = sectionAt(file, i);
        const box = document.createElement("button");
        const colorClass = section ? classFor(section.kind) : "neutral";
        box.className = "byte " + colorClass;
        if (selected && i >= selected.start && i < selected.end) box.className += " selected";
        box.onmouseenter = (event) => {
          const ascii = printable(parseInt(hex, 16));
          const group = contiguousRange(file, i, colorClass);
          els.tooltip.textContent = "offset: " + i + "\nbyte: 0x" + hex + "\nascii: " + quote(ascii) + "\nvalue: " + quote(bytesText(file, group.start, group.end));
          els.tooltip.style.display = "block";
          moveTooltip(event);
        };
        box.onmousemove = moveTooltip;
        box.onmouseleave = () => els.tooltip.style.display = "none";
        box.onclick = () => {
          selected = { ...contiguousRange(file, i, colorClass), id: "byte-" + i, source: "byte" };
          renderBytes();
        };
        els.bytes.appendChild(box);
      }
    }
    function printable(byte) {
      return byte >= 32 && byte <= 126 ? String.fromCharCode(byte) : ".";
    }
    function byteClass(file, offset) {
      const section = sectionAt(file, offset);
      return section ? classFor(section.kind) : "neutral";
    }
    function contiguousRange(file, offset, colorClass) {
      let start = offset;
      let end = offset + 1;
      while (start > 0 && byteClass(file, start - 1) === colorClass) start--;
      while (end < file.size && byteClass(file, end) === colorClass) end++;
      return { start, end };
    }
    function bytesText(file, start, end) {
      let text = "";
      for (let i = start; i < end; i++) {
        text += printable(parseInt(file.hex.slice(i * 2, i * 2 + 2), 16));
      }
      return text;
    }
    function quote(value) {
      return "\"" + String(value).replaceAll("\\", "\\\\").replaceAll("\"", "\\\"") + "\"";
    }
    function moveTooltip(event) {
      const margin = 14;
      const x = Math.min(event.clientX + margin, window.innerWidth - els.tooltip.offsetWidth - margin);
      const y = Math.min(event.clientY + margin, window.innerHeight - els.tooltip.offsetHeight - margin);
      els.tooltip.style.left = Math.max(margin, x) + "px";
      els.tooltip.style.top = Math.max(margin, y) + "px";
    }
    setActive(0);
  </script>
</body>
</html>`
