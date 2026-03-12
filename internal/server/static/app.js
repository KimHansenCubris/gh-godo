/* app.js – Workload Capacity Sheet UI */
(() => {
  "use strict";

  // ── DOM refs ──────────────────────────────────────────────────
  const filePathInput = document.getElementById("filePath");
  const loadBtn       = document.getElementById("loadBtn");
  const fileStatus    = document.getElementById("fileStatus");
  const entryForm     = document.getElementById("entryForm");
  const formTitle     = document.getElementById("formTitle");
  const saveBtn       = document.getElementById("saveBtn");
  const cancelBtn     = document.getElementById("cancelBtn");
  const formStatus    = document.getElementById("formStatus");
  const previewBody   = document.getElementById("previewBody");
  const rowCount      = document.getElementById("rowCount");

  const fields = ["project", "swtb", "version", "week", "type", "resources", "year", "quarter"];

  // ── State ─────────────────────────────────────────────────────
  let currentFile = "";
  let editingID   = null; // null = add mode, number = edit mode

  // ── Helpers ───────────────────────────────────────────────────
  function setStatus(el, msg, ok) {
    el.textContent = msg;
    el.className   = "status " + (ok ? "status--ok" : (msg ? "status--err" : ""));
  }

  function apiUrl(path) {
    return `${path}?file=${encodeURIComponent(currentFile)}`;
  }

  function formValues() {
    return {
      project:   document.getElementById("project").value.trim(),
      swtb:      document.getElementById("swtb").value.trim(),
      version:   document.getElementById("version").value.trim(),
      week:      document.getElementById("week").value.trim(),
      type:      document.getElementById("type").value,
      resources: document.getElementById("resources").value.trim(),
      year:      document.getElementById("year").value.trim(),
      quarter:   document.getElementById("quarter").value,
    };
  }

  function fillForm(row) {
    document.getElementById("project").value   = row.project   || "";
    document.getElementById("swtb").value      = row.swtb      || "";
    document.getElementById("version").value   = row.version   || "";
    document.getElementById("week").value      = row.week      || "";
    document.getElementById("type").value      = row.type      || "";
    document.getElementById("resources").value = row.resources || "";
    document.getElementById("year").value      = row.year      || "";
    document.getElementById("quarter").value   = row.quarter   || "";
  }

  function clearForm() {
    fields.forEach(id => { document.getElementById(id).value = ""; });
  }

  function enterEditMode(row) {
    editingID = row.id;
    formTitle.textContent = "Edit Entry";
    fillForm(row);
    cancelBtn.style.display = "";
    setStatus(formStatus, "");
    // Highlight the row being edited
    document.querySelectorAll("tbody tr").forEach(tr => {
      tr.classList.toggle("editing", tr.dataset.id === String(row.id));
    });
  }

  function exitEditMode() {
    editingID = null;
    formTitle.textContent = "Add Entry";
    clearForm();
    cancelBtn.style.display = "none";
    setStatus(formStatus, "");
    document.querySelectorAll("tbody tr").forEach(tr => tr.classList.remove("editing"));
  }

  // ── Render preview table ──────────────────────────────────────
  function renderRows(rows) {
    rowCount.textContent = `${rows.length} row${rows.length !== 1 ? "s" : ""}`;

    if (rows.length === 0) {
      previewBody.innerHTML = `<tr id="emptyRow"><td colspan="9" class="empty">No data in this sheet yet.</td></tr>`;
      return;
    }

    previewBody.innerHTML = rows.map(r => `
      <tr data-id="${r.id}">
        <td>${esc(r.project)}</td>
        <td>${esc(r.swtb)}</td>
        <td>${esc(r.version)}</td>
        <td>${esc(r.week)}</td>
        <td>${esc(r.type)}</td>
        <td>${esc(r.resources)}</td>
        <td>${esc(r.year)}</td>
        <td>${esc(r.quarter)}</td>
        <td class="actions">
          <button class="btn btn--edit" data-action="edit" data-id="${r.id}">Edit</button>
          <button class="btn btn--danger" data-action="delete" data-id="${r.id}">Delete</button>
        </td>
      </tr>`).join("");
  }

  function esc(str) {
    if (!str) return "";
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  // ── Load / refresh data ───────────────────────────────────────
  async function loadRows() {
    try {
      const res = await fetch(apiUrl("/api/rows"));
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || res.statusText);
      }
      const rows = await res.json();
      renderRows(rows || []);
      setStatus(fileStatus, `Loaded: ${currentFile}`, true);
    } catch (err) {
      setStatus(fileStatus, `Error: ${err.message}`);
      renderRows([]);
    }
  }

  // ── Event: Load button ─────────────────────────────────────────
  loadBtn.addEventListener("click", async () => {
    const p = filePathInput.value.trim();
    if (!p) { setStatus(fileStatus, "Please enter a file path."); return; }
    currentFile = p;
    saveBtn.disabled = false;
    exitEditMode();
    await loadRows();
  });

  // Also allow Enter in the file path input
  filePathInput.addEventListener("keydown", e => {
    if (e.key === "Enter") loadBtn.click();
  });

  // ── Event: Form submit (add or update) ────────────────────────
  entryForm.addEventListener("submit", async e => {
    e.preventDefault();
    if (!currentFile) { setStatus(formStatus, "Load a file first."); return; }

    const data = formValues();
    setStatus(formStatus, "");

    try {
      let res;
      if (editingID === null) {
        res = await fetch(apiUrl("/api/rows"), {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
        });
      } else {
        res = await fetch(apiUrl(`/api/rows/${editingID}`), {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(data),
        });
      }

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || res.statusText);
      }

      setStatus(formStatus, editingID === null ? "Entry saved!" : "Entry updated!", true);
      exitEditMode();
      await loadRows();
    } catch (err) {
      setStatus(formStatus, `Save failed: ${err.message}`);
    }
  });

  // ── Event: Cancel button ──────────────────────────────────────
  cancelBtn.addEventListener("click", exitEditMode);

  // ── Event: table action buttons (edit / delete) ───────────────
  previewBody.addEventListener("click", async e => {
    const btn = e.target.closest("button[data-action]");
    if (!btn) return;

    const id     = parseInt(btn.dataset.id, 10);
    const action = btn.dataset.action;

    if (action === "edit") {
      // Find row data from the table row
      const tr = btn.closest("tr");
      const cells = tr.querySelectorAll("td");
      enterEditMode({
        id,
        project:   cells[0].textContent,
        swtb:      cells[1].textContent,
        version:   cells[2].textContent,
        week:      cells[3].textContent,
        type:      cells[4].textContent,
        resources: cells[5].textContent,
        year:      cells[6].textContent,
        quarter:   cells[7].textContent,
      });
    }

    if (action === "delete") {
      if (!confirm("Delete this row from the Excel file?")) return;
      try {
        const res = await fetch(apiUrl(`/api/rows/${id}`), { method: "DELETE" });
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          throw new Error(body.error || res.statusText);
        }
        if (editingID === id) exitEditMode();
        await loadRows();
      } catch (err) {
        setStatus(fileStatus, `Delete failed: ${err.message}`);
      }
    }
  });
})();
