package viewer

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>RDB File Explorer</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f6f6f2;
      --surface: #ffffff;
      --key-byte: #79c8ff;
      --value-byte: #c9edbd;
      --type-byte: #e2e5e1;
      --ink: #202322;
      --muted: #666d69;
      --line: #252827;
      --soft-line: #d8dbd7;
      --header: #b9f2c8;
      --aux: #ffe49b;
      --db: #b7dcff;
      --set: #f7c1c5;
      --string: #bfe8dd;
      --stream: #d7c7ff;
      --hash: #ffd0a8;
      --array: #c8df9e;
      --zset: #c9d8ff;
      --list: #f5c6ec;
      --hll: #c8e4f0;
      --eof: #d6d6d6;
      --checksum: #b9b1a6;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      min-height: 100vh;
      background: var(--bg);
      color: var(--ink);
    }

    button {
      font: inherit;
    }

    .app {
      display: grid;
      grid-template-columns: var(--left-width, 210px) 5px minmax(320px, 1fr) 5px var(--right-width, 280px);
      min-height: 100vh;
      height: 100vh;
      overflow: hidden;
    }

    .files {
      background: #fbfbf8;
      padding: 18px 14px;
      min-width: 0;
      overflow-y: auto;
    }

    .files h1,
    .pane-title {
      margin: 0 0 14px;
      font-size: 13px;
      color: var(--muted);
      text-transform: uppercase;
      letter-spacing: 0;
      font-weight: 740;
    }

    .file-list {
      display: grid;
      gap: 6px;
    }

    .file-button {
      width: 100%;
      min-height: 34px;
      border: 1px solid var(--soft-line);
      border-radius: 6px;
      background: var(--surface);
      color: var(--ink);
      padding: 8px 10px;
      cursor: pointer;
      text-align: left;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 13px;
      font-weight: 660;
    }

    .file-button:hover,
    .file-button.active {
      border-color: var(--line);
      background: #f1f3ef;
    }

    .resizer {
      background: var(--line);
      cursor: col-resize;
      position: relative;
      user-select: none;
      touch-action: none;
      transition: background 0.12s ease;
      z-index: 5;
    }

    .resizer:hover,
    .resizer.dragging {
      background: #0066cc;
      box-shadow: 0 0 0 1px #0066cc;
    }

    .resizer::after {
      content: "";
      position: absolute;
      top: 0;
      bottom: 0;
      left: -4px;
      right: -4px;
    }

    .center {
      display: grid;
      grid-template-rows: auto 1fr;
      min-width: 0;
      background: #fcfcfa;
      overflow: hidden;
    }

    .tabs {
      display: flex;
      align-items: end;
      min-height: 48px;
      gap: 4px;
      padding: 12px 14px 0;
      border-bottom: 2px solid var(--line);
      overflow-x: auto;
      background: var(--surface);
    }

    .tab {
      border: 2px solid var(--line);
      border-bottom: 0;
      border-radius: 7px 7px 0 0;
      background: #f5f6f3;
      min-width: 132px;
      max-width: 220px;
      height: 36px;
      padding: 0 12px;
      cursor: pointer;
      color: var(--ink);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 13px;
      font-weight: 700;
    }

    .tab.active {
      background: #fcfcfa;
      height: 38px;
      transform: translateY(2px);
    }

    .records {
      overflow-y: auto;
      padding: 18px 18px 28px;
    }

    .file-heading {
      display: flex;
      align-items: baseline;
      justify-content: space-between;
      gap: 16px;
      padding-bottom: 14px;
      border-bottom: 1px solid var(--soft-line);
      margin-bottom: 14px;
    }

    .file-heading h2 {
      margin: 0;
      font-size: 22px;
      line-height: 1.2;
      letter-spacing: 0;
    }

    .file-stats {
      display: flex;
      flex-wrap: wrap;
      justify-content: flex-end;
      gap: 7px;
      color: var(--muted);
      font-size: 12px;
    }

    .stat {
      border: 1px solid var(--soft-line);
      border-radius: 999px;
      background: #ffffff;
      padding: 5px 8px;
      white-space: nowrap;
    }

    .section-group {
      margin-top: 18px;
      border: 1px solid var(--soft-line);
      border-left: 8px solid #d5d9d4;
      border-radius: 6px;
      background: #ffffff;
      padding: 10px;
    }

    .section-group-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      cursor: pointer;
      user-select: none;
      padding: 2px 2px 6px;
    }

    .section-group-header h3 {
      margin: 0;
      font-size: 13px;
      text-transform: uppercase;
      color: var(--muted);
      letter-spacing: 0;
      font-weight: 700;
    }

    .section-toggle-btn {
      color: var(--muted);
      font-size: 11px;
      display: inline-flex;
      align-items: center;
      transition: transform 0.15s ease;
    }

    .section-group.collapsed .section-toggle-btn {
      transform: rotate(-90deg);
    }

    .section-group.collapsed .record-row {
      display: none;
    }

    .section-group:hover,
    .section-group.active {
      border-color: var(--line);
      box-shadow: inset 0 0 0 1px var(--line);
    }

    .record-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 4px;
    }

    .record {
      width: auto;
      display: grid;
      grid-template-columns: minmax(170px, 1fr) minmax(96px, auto);
      gap: 8px;
      align-items: center;
      border: 1px solid var(--soft-line);
      border-radius: 6px;
      background: var(--surface);
      min-height: 36px;
      padding: 7px 8px;
      margin: 0;
      text-align: left;
      cursor: pointer;
      color: var(--ink);
    }

    .record:hover,
    .record.active {
      border-color: var(--line);
      box-shadow: inset 0 0 0 1px var(--line);
    }

    .record-name {
      display: none;
    }

    .record-value {
      font-family: ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace;
      overflow-wrap: anywhere;
      font-size: 12px;
      line-height: 1.32;
      color: #343937;
    }

    .record-side {
      font-family: ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace;
      color: var(--muted);
      font-size: 11px;
      justify-self: end;
      text-align: right;
      line-height: 1.28;
      white-space: pre-line;
    }

    .bytes-pane {
      display: grid;
      grid-template-rows: auto 1fr auto;
      min-width: 0;
      background: #ffffff;
      overflow: hidden;
    }

    .bytes-header {
      min-height: 52px;
      padding: 12px 16px 10px;
      border-bottom: 2px solid var(--line);
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .bytes-header-title h2 {
      margin: 0;
      font-size: 18px;
      letter-spacing: 0;
    }

    .byte-legend {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      font-size: 11px;
      color: var(--muted);
      font-family: ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace;
      line-height: 1.2;
    }

    .legend-item {
      display: inline-flex;
      align-items: center;
      gap: 4px;
    }

    .legend-dot {
      width: 10px;
      height: 10px;
      border-radius: 2px;
      border: 1px solid rgba(0, 0, 0, .5);
      display: inline-block;
    }

    .legend-dot.key {
      background: var(--key-byte);
    }

    .legend-dot.value {
      background: var(--value-byte);
    }

    .legend-dot.neutral {
      background: var(--type-byte);
    }

    .byte-wrap {
      overflow-y: auto;
      overflow-x: hidden;
      padding: 18px 16px;
      min-width: 0;
    }

    .byte-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, 18px);
      grid-auto-rows: 18px;
      gap: 4px;
      align-content: start;
      width: 100%;
    }

    .byte {
      width: 18px;
      height: 18px;
      border: 1px solid rgba(0, 0, 0, .62);
      border-radius: 4px;
      cursor: pointer;
      transition: transform .08s ease, box-shadow .08s ease;
    }

    .byte-grid.has-focus .byte:not(.focused) {
      opacity: .24;
    }

    .byte.neutral-part,
    .byte.nonprintable {
      background: #e7e9e6;
      border-color: #aeb4af;
    }

    .byte.key-part {
      background: var(--key-byte);
    }

    .byte.value-part {
      background: var(--value-byte);
    }

    .byte.type-part {
      background: #e7e9e6;
    }

    .byte:hover,
    .byte.focused {
      opacity: 1 !important;
      transform: translateY(-1px);
      box-shadow: 0 0 0 2px #202322;
      z-index: 2;
    }

    .byte.key-part:hover,
    .byte.key-part.focused {
      box-shadow: 0 0 0 2px #202322;
    }

    .byte.value-part:hover,
    .byte.value-part.focused {
      box-shadow: 0 0 0 2px #202322;
    }

    .byte.type-part:hover,
    .byte.type-part.focused {
      box-shadow: 0 0 0 2px #202322;
    }

    .byte-detail {
      border-top: 2px solid var(--line);
      padding: 12px 16px 14px;
      display: grid;
      grid-template-columns: 1fr auto;
      gap: 12px;
      align-items: start;
      min-height: 86px;
    }

    .byte-detail h3 {
      margin: 0 0 5px;
      font-size: 15px;
      letter-spacing: 0;
    }

    .byte-detail p {
      margin: 0;
      font-size: 13px;
      color: #3c403e;
      line-height: 1.4;
    }

    .range-text {
      font-family: ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace;
      color: var(--muted);
      font-size: 12px;
      white-space: nowrap;
    }

    .tooltip {
      position: fixed;
      pointer-events: none;
      display: none;
      max-width: min(340px, calc(100vw - 24px));
      border: 1px solid var(--line);
      background: #ffffff;
      color: var(--ink);
      padding: 8px 9px;
      border-radius: 6px;
      box-shadow: 0 12px 28px rgba(0, 0, 0, .16);
      font-size: 12px;
      line-height: 1.35;
      z-index: 10;
    }

    .tooltip strong {
      display: block;
      margin-bottom: 4px;
    }

    @media (max-width: 900px) {
      .app {
        grid-template-columns: 1fr;
        grid-template-rows: auto auto minmax(420px, 52vh);
        height: auto;
        overflow: auto;
      }

      .resizer {
        display: none;
      }

      .files {
        border-bottom: 2px solid var(--line);
      }

      .file-list {
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
      }

      .center {
        border-bottom: 2px solid var(--line);
      }

      .record {
        grid-template-columns: 1fr;
        gap: 6px;
      }

      .record-side {
        justify-self: start;
        text-align: left;
      }
    }

    .error-banner {
      display: none;
      background: #ffebe9;
      border: 1px solid #cf222e;
      color: #82071e;
      padding: 10px 14px;
      font-size: 13px;
      font-weight: 600;
      border-radius: 6px;
      margin: 12px 14px;
    }
  </style>
</head>
<body>
  <div class="error-banner" id="errorBanner" role="alert"></div>
  <div class="app" id="app">
    <aside class="files" id="paneFiles">
      <h1>Files</h1>
      <div class="file-list" id="fileList"></div>
    </aside>

    <div class="resizer" id="resizerLeft" title="Drag to resize panel"></div>

    <main class="center" id="paneCenter">
      <nav class="tabs" id="tabs" aria-label="Open files"></nav>
      <section class="records">
        <div class="file-heading">
          <h2 id="fileTitle"></h2>
          <div class="file-stats" id="fileStats"></div>
        </div>
        <div id="recordGroups"></div>
      </section>
    </main>

    <div class="resizer" id="resizerRight" title="Drag to resize panel"></div>

    <section class="bytes-pane" id="paneBytes">
      <header class="bytes-header">
        <div class="bytes-header-title">
          <h2>Byte view</h2>
        </div>
        <div class="byte-legend">
          <span class="legend-item"><span class="legend-dot key"></span>Key (Blue)</span>
          <span class="legend-item"><span class="legend-dot value"></span>Value (Green)</span>
          <span class="legend-item"><span class="legend-dot neutral"></span>Other (Grey)</span>
        </div>
      </header>
      <div class="byte-wrap">
        <div class="byte-grid" id="byteGrid" aria-label="RDB byte grid"></div>
      </div>
      <footer class="byte-detail">
        <div>
          <h3 id="activeTitle">Header</h3>
          <p id="activeSummary">REDIS file signature and RDB version.</p>
        </div>
        <div class="range-text" id="activeRange">0x0000-0x0008</div>
      </footer>
    </section>
  </div>
  <div class="tooltip" id="tooltip"></div>

  <script>
    const files = {{.Payload}};

    const fileList = document.getElementById("fileList");
    const tabs = document.getElementById("tabs");
    const fileTitle = document.getElementById("fileTitle");
    const fileStats = document.getElementById("fileStats");
    const recordGroups = document.getElementById("recordGroups");
    const byteGrid = document.getElementById("byteGrid");
    const activeTitle = document.getElementById("activeTitle");
    const activeSummary = document.getElementById("activeSummary");
    const activeRange = document.getElementById("activeRange");
    const tooltip = document.getElementById("tooltip");

    const openFileIds = [];
    const collapsedGroups = new Set();
    let activeFile = files && files.length > 0 ? files[0] : null;
    let activeRecord = activeFile ? fileTarget(activeFile) : null;
    let byteNodes = [];
    let spanByOffset = new Map();

    function hex2(n) {
      return n.toString(16).padStart(2, "0");
    }

    function hex4(n) {
      return "0x" + n.toString(16).padStart(4, "0");
    }

    function printable(byte) {
      return byte >= 32 && byte <= 126 ? String.fromCharCode(byte) : ".";
    }

    function isPrintable(byte) {
      return byte >= 32 && byte <= 126;
    }

    function allRecords(file) {
      if (!file.groups) return [];
      return file.groups.flatMap((group) => group.records);
    }

    function fileTarget(file) {
      return {
        id: "file-" + file.id,
        label: file.name,
        type: "file",
        encoding: file.version,
        size: file.bytes + "B",
        start: 0,
        end: file.bytes,
        summary: "Entire RDB file."
      };
    }

    function groupTarget(group) {
      const records = group.records || [];
      const start = records.length > 0 ? records[0].start : 0;
      const end = records.length > 0 ? records[records.length - 1].end : 0;
      const color = records.length > 0 ? records[0].color : "var(--db)";
      return {
        id: "group-" + group.title.toLowerCase().replaceAll(" ", "-"),
        label: group.title,
        type: "group",
        encoding: "mixed",
        size: records.reduce((sum, r) => sum + (r.end - r.start), 0) + "B",
        start: start,
        end: end,
        color: color,
        summary: group.title + " byte region."
      };
    }

    function toggleGroup(groupId) {
      if (collapsedGroups.has(groupId)) {
        collapsedGroups.delete(groupId);
      } else {
        collapsedGroups.add(groupId);
      }
      renderRecords();
    }

    let focusedNodes = [];
    let currentActiveRecordNode = null;
    let currentActiveGroupNode = null;

    function openFile(file) {
      if (!openFileIds.includes(file.id)) openFileIds.push(file.id);
      activeFile = file;
      activeRecord = fileTarget(file);
      render();
    }

    function setActiveRecord(record) {
      if (!record) return;
      activeRecord = record;
      activeTitle.textContent = record.label;
      activeSummary.textContent = record.summary;
      activeRange.textContent = hex4(record.start) + "-" + hex4(Math.max(0, record.end - 1));

      // Fast active class toggle for records
      if (currentActiveRecordNode) currentActiveRecordNode.classList.remove("active");
      currentActiveRecordNode = document.querySelector(".record[data-record-id=\"" + record.id + "\"]");
      if (currentActiveRecordNode) currentActiveRecordNode.classList.add("active");

      // Fast active class toggle for groups
      if (currentActiveGroupNode) currentActiveGroupNode.classList.remove("active");
      currentActiveGroupNode = document.querySelector(".section-group[data-group-id=\"" + record.id + "\"]");
      if (currentActiveGroupNode) currentActiveGroupNode.classList.add("active");

      // Clear previously focused byte nodes
      for (let i = 0; i < focusedNodes.length; i++) {
        focusedNodes[i].classList.remove("focused");
      }
      focusedNodes = [];

      const isFullFile = record.type === "file";
      if (!isFullFile && record.start < byteNodes.length) {
        byteGrid.classList.add("has-focus");
        const end = Math.min(record.end, byteNodes.length);
        for (let i = record.start; i < end; i++) {
          const node = byteNodes[i];
          if (node) {
            node.classList.add("focused");
            focusedNodes.push(node);
          }
        }
        // Auto scroll to first focused byte if not in view
        const firstByte = byteNodes[record.start];
        if (firstByte) {
          firstByte.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "nearest" });
        }
      } else {
        byteGrid.classList.remove("has-focus");
      }
    }

    function setActiveByteRun(offset) {
      if (!byteNodes[offset]) return;
      const targetClass = byteRunClass(byteNodes[offset]);
      let start = offset;
      let end = offset + 1;
      while (start > 0 && byteRunClass(byteNodes[start - 1]) === targetClass) start--;
      while (end < byteNodes.length && byteRunClass(byteNodes[end]) === targetClass) end++;

      activeTitle.textContent = "Byte run";
      activeSummary.textContent = targetClass === "key-part" ? "Continuous key bytes." :
        targetClass === "value-part" ? "Continuous value bytes." :
        "Continuous neutral bytes.";
      activeRange.textContent = hex4(start) + "-" + hex4(Math.max(0, end - 1));

      if (currentActiveRecordNode) currentActiveRecordNode.classList.remove("active");
      if (currentActiveGroupNode) currentActiveGroupNode.classList.remove("active");

      for (let i = 0; i < focusedNodes.length; i++) {
        focusedNodes[i].classList.remove("focused");
      }
      focusedNodes = [];

      byteGrid.classList.add("has-focus");
      for (let i = start; i < end; i++) {
        const node = byteNodes[i];
        if (node) {
          node.classList.add("focused");
          focusedNodes.push(node);
        }
      }
    }

    function byteRunClass(node) {
      if (node.classList.contains("key-part")) return "key-part";
      if (node.classList.contains("value-part")) return "value-part";
      return "neutral-part";
    }

    function renderFiles() {
      fileList.innerHTML = "";
      files.forEach((file) => {
        const button = document.createElement("button");
        button.className = "file-button";
        button.type = "button";
        button.textContent = file.name;
        button.title = file.path;
        button.classList.toggle("active", activeFile && file.id === activeFile.id);
        button.addEventListener("click", () => openFile(file));
        fileList.appendChild(button);
      });
    }

    function renderTabs() {
      tabs.innerHTML = "";
      openFileIds.forEach((id) => {
        const file = files.find((item) => item.id === id);
        if (!file) return;
        const tab = document.createElement("button");
        tab.className = "tab";
        tab.type = "button";
        tab.textContent = file.name;
        tab.classList.toggle("active", activeFile && file.id === activeFile.id);
        tab.addEventListener("click", () => openFile(file));
        tabs.appendChild(tab);
      });
    }

    function renderRecords() {
      if (!activeFile) return;
      fileTitle.textContent = activeFile.name;
      fileStats.innerHTML = "";
      [activeFile.version, activeFile.bytes + "B", activeFile.countLabel].forEach((text) => {
        if (!text) return;
        const stat = document.createElement("span");
        stat.className = "stat";
        stat.textContent = text;
        fileStats.appendChild(stat);
      });

      recordGroups.innerHTML = "";
      if (!activeFile.groups) return;
      activeFile.groups.forEach((group) => {
        const wrapper = document.createElement("section");
        wrapper.className = "section-group";
        const target = groupTarget(group);
        wrapper.dataset.groupId = target.id;
        wrapper.style.borderLeftColor = target.color;

        const isCollapsed = collapsedGroups.has(target.id);
        if (isCollapsed) {
          wrapper.classList.add("collapsed");
        }

        const header = document.createElement("div");
        header.className = "section-group-header";
        header.setAttribute("role", "button");
        header.setAttribute("tabindex", "0");
        header.setAttribute("aria-expanded", isCollapsed ? "false" : "true");

        const heading = document.createElement("h3");
        heading.textContent = group.title;

        const toggleBtn = document.createElement("span");
        toggleBtn.className = "section-toggle-btn";
        toggleBtn.textContent = "▼";

        header.append(heading, toggleBtn);
        header.addEventListener("click", (e) => {
          e.stopPropagation();
          toggleGroup(target.id);
          setActiveRecord(target);
        });

        wrapper.appendChild(header);

        const recordRow = document.createElement("div");
        recordRow.className = "record-row";

        (group.records || []).forEach((record) => {
          const row = document.createElement("button");
          row.className = "record";
          row.type = "button";
          row.dataset.recordId = record.id;
          row.style.borderLeftColor = record.color;

          const name = document.createElement("div");
          name.className = "record-name";
          name.textContent = record.label;

          const value = document.createElement("div");
          value.className = "record-value";
          if (record.json && record.json !== record.value) {
            const kv = document.createElement("div");
            kv.textContent = record.value;
            const json = document.createElement("div");
            json.textContent = record.json;
            value.append(kv, json);
          } else {
            value.textContent = record.value;
          }

          const side = document.createElement("div");
          side.className = "record-side";
          side.textContent = "type=" + record.type + "\nencoding=" + record.encoding + "\nsize=" + record.size;

          row.append(name, value, side);
          row.addEventListener("click", (event) => {
            event.stopPropagation();
            setActiveRecord(record);
          });
          recordRow.appendChild(row);
        });

        wrapper.appendChild(recordRow);
        recordGroups.appendChild(wrapper);
      });
    }

    function renderBytes() {
      byteGrid.innerHTML = "";
      byteGrid.classList.remove("has-focus");
      spanByOffset = new Map();
      byteNodes = [];
      focusedNodes = [];
      if (!activeFile || !activeFile.hex) return;

      const records = allRecords(activeFile);
      records.forEach((record) => {
        for (let i = record.start; i < record.end; i++) spanByOffset.set(i, record);
      });

      const hexMatches = activeFile.hex.match(/../g) || [];
      const byteLimit = 40000;
      const count = Math.min(hexMatches.length, byteLimit);
      const frag = document.createDocumentFragment();

      for (let offset = 0; offset < count; offset++) {
        const byte = parseInt(hexMatches[offset], 16);
        const record = spanByOffset.get(offset) || activeRecord;
        const node = document.createElement("button");
        node.className = "byte";
        node.dataset.offset = offset;
        node.dataset.byte = byte;
        const part = record && record.parts && record.parts.find((item) => offset >= item.start && offset < item.end);
        if (!isPrintable(byte)) {
          node.classList.add("neutral-part", "nonprintable");
        } else if (part && part.kind === "key") {
          node.classList.add("key-part");
        } else if (part && part.kind === "value") {
          node.classList.add("value-part");
        } else {
          node.classList.add("neutral-part");
        }
        node.type = "button";
        if (record && record.color) {
          node.style.borderColor = record.color;
        }
        node.setAttribute("aria-label", hex4(offset) + " byte " + hex2(byte) + (record ? " " + record.label : ""));
        frag.appendChild(node);
        byteNodes.push(node);
      }

      byteGrid.appendChild(frag);
    }

    function stringSegment(record, offset) {
      if (record && record.strings) {
        const found = record.strings.find((item) => offset >= item.start && offset < item.end);
        if (found) return found;
      }
      if (record && record.parts) {
        const part = record.parts.find((item) => offset >= item.start && offset < item.end);
        if (part) {
          if (part.kind === "key") {
            return { kind: "key", text: record.label };
          }
          if (part.kind === "value") {
            if (activeFile && activeFile.hex) {
              const hexMatches = activeFile.hex.match(/../g) || [];
              const isPrint = (idx) => {
                if (idx < part.start || idx >= part.end || idx >= hexMatches.length) return false;
                const b = parseInt(hexMatches[idx], 16);
                return isPrintable(b);
              };
              if (isPrint(offset)) {
                let start = offset;
                let end = offset + 1;
                while (start > part.start && isPrint(start - 1)) start--;
                while (end < part.end && isPrint(end)) end++;
                let text = "";
                for (let i = start; i < end; i++) {
                  text += String.fromCharCode(parseInt(hexMatches[i], 16));
                }
                return { kind: "value", text: text };
              }
            }
            return { kind: "value", text: record.label };
          }
        }
      }
      return null;
    }

    function tooltipText(offset, byte, record) {
      const segment = stringSegment(record, offset);
      const segmentLine = segment ? "<br>" + segment.kind + ": \"" + escapeHtml(segment.text) + "\"" : "";
      return "offset: " + hex4(offset) + "<br>" +
        "byte: 0x" + hex2(byte) + " (" + byte + ")<br>" +
        "ascii: " + escapeHtml(printable(byte)) +
        segmentLine;
    }

    function escapeHtml(value) {
      return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll("\"", "&quot;");
    }

    function positionTooltip(event) {
      const margin = 14;
      const x = Math.min(event.clientX + margin, window.innerWidth - tooltip.offsetWidth - margin);
      const y = Math.min(event.clientY + margin, window.innerHeight - tooltip.offsetHeight - margin);
      tooltip.style.left = Math.max(margin, x) + "px";
      tooltip.style.top = Math.max(margin, y) + "px";
    }

    function showErrorBanner(message) {
      try {
        const banner = document.getElementById("errorBanner");
        if (banner) {
          banner.textContent = message;
          banner.style.display = "block";
        }
      } catch {
        // Fallback if DOM error occurs
      }
    }

    window.addEventListener("error", (e) => {
      showErrorBanner("Viewer error: " + (e && e.message ? e.message : "Unknown error"));
    });

    window.addEventListener("unhandledrejection", (e) => {
      showErrorBanner("Async error: " + (e && e.reason ? (e.reason.message || String(e.reason)) : "Unknown rejection"));
    });

    function setupByteGridEvents() {
      try {
        if (!byteGrid) return;
        byteGrid.addEventListener("click", (e) => {
          const btn = e.target.closest(".byte");
          if (!btn) return;
          const offset = parseInt(btn.dataset.offset, 10);
          if (!isNaN(offset)) setActiveByteRun(offset);
        });

        byteGrid.addEventListener("mousemove", (e) => {
          const btn = e.target.closest(".byte");
          if (!btn) {
            tooltip.style.display = "none";
            return;
          }
          const offset = parseInt(btn.dataset.offset, 10);
          if (isNaN(offset)) return;
          const byte = parseInt(btn.dataset.byte, 10);
          const record = spanByOffset.get(offset) || activeRecord;
          tooltip.innerHTML = tooltipText(offset, byte, record);
          tooltip.style.display = "block";
          positionTooltip(e);
        });

        byteGrid.addEventListener("mouseleave", () => {
          tooltip.style.display = "none";
        });
      } catch (err) {
        console.error("ByteGrid events error:", err);
      }
    }

    function setupResizers() {
      try {
        const appEl = document.getElementById("app");
        const resizerLeft = document.getElementById("resizerLeft");
        const resizerRight = document.getElementById("resizerRight");
        if (!resizerLeft || !resizerRight || !appEl) return;

        let isDraggingLeft = false;
        let isDraggingRight = false;

        resizerLeft.addEventListener("mousedown", (e) => {
          isDraggingLeft = true;
          resizerLeft.classList.add("dragging");
          document.body.style.cursor = "col-resize";
          document.body.style.userSelect = "none";
          e.preventDefault();
        });

        resizerRight.addEventListener("mousedown", (e) => {
          isDraggingRight = true;
          resizerRight.classList.add("dragging");
          document.body.style.cursor = "col-resize";
          document.body.style.userSelect = "none";
          e.preventDefault();
        });

        window.addEventListener("mousemove", (e) => {
          if (isDraggingLeft) {
            const minLeft = 140;
            const maxLeft = Math.max(minLeft, window.innerWidth - 450);
            const newWidth = Math.max(minLeft, Math.min(e.clientX, maxLeft));
            appEl.style.setProperty("--left-width", newWidth + "px");
          } else if (isDraggingRight) {
            const minRight = 200;
            const maxRight = Math.max(minRight, window.innerWidth - 400);
            const newWidth = Math.max(minRight, Math.min(window.innerWidth - e.clientX, maxRight));
            appEl.style.setProperty("--right-width", newWidth + "px");
          }
        });

        window.addEventListener("mouseup", () => {
          if (isDraggingLeft || isDraggingRight) {
            isDraggingLeft = false;
            isDraggingRight = false;
            resizerLeft.classList.remove("dragging");
            resizerRight.classList.remove("dragging");
            document.body.style.cursor = "";
            document.body.style.userSelect = "";
          }
        });
      } catch (err) {
        console.error("Resizer setup error:", err);
      }
    }

    function render() {
      try {
        renderFiles();
        renderTabs();
        renderRecords();
        renderBytes();
        if (activeRecord) {
          setActiveRecord(activeRecord);
        }
      } catch (err) {
        console.error("Render error:", err);
        showErrorBanner("Failed to render RDB view: " + (err && err.message ? err.message : String(err)));
      }
    }

    try {
      setupResizers();
      setupByteGridEvents();
      if (!files || !Array.isArray(files) || files.length === 0) {
        showErrorBanner("No RDB files loaded in this viewer session.");
      } else if (activeFile) {
        openFile(activeFile);
      }
    } catch (err) {
      console.error("Initialization error:", err);
      showErrorBanner("Failed to initialize RDB viewer: " + (err && err.message ? err.message : String(err)));
    }
  </script>
</body>
</html>
`
