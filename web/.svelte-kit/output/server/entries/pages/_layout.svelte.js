import { s as ssr_context, a as attr_style, e as escape_html, b as ensure_array_like, d as derived, c as stringify, f as attr_class, g as store_get, h as attr, u as unsubscribe_stores, i as store_set, j as bind_props, k as store_mutate } from "../../chunks/renderer.js";
import { w as writable } from "../../chunks/index.js";
import "clsx";
import { Compartment } from "@codemirror/state";
import "ag-grid-community";
function onDestroy(fn) {
  /** @type {SSRContext} */
  ssr_context.r.on_destroy(fn);
}
const connections = writable([]);
const activeConnectionId = writable(null);
const activeDatabase = writable("");
const tabs = writable([]);
const activeTabId = writable(null);
const explorerTree = writable([]);
const expandedNodes = writable(/* @__PURE__ */ new Set());
const sidebarCollapsed = writable(false);
const API_BASE = "/api";
async function request(path, options) {
  const resp = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...options?.headers
    },
    ...options
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: resp.statusText }));
    throw new Error(err.error || `HTTP ${resp.status}`);
  }
  return resp.json();
}
const api = {
  connections: {
    list: () => request("/connections"),
    get: (id) => request(`/connections/${id}`),
    create: (data) => request("/connections", { method: "POST", body: JSON.stringify(data) }),
    update: (id, data) => request(`/connections/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id) => request(`/connections/${id}`, { method: "DELETE" }),
    test: (id) => request(`/connections/${id}/test`, { method: "POST" })
  },
  metadata: {
    databases: (connId) => request(`/connections/${connId}/databases`),
    schemas: (connId, db) => request(`/connections/${connId}/schemas${db ? `?database=${db}` : ""}`),
    tables: (connId, schema) => request(`/connections/${connId}/schemas/${schema}/tables`),
    columns: (connId, schema, table) => request(`/connections/${connId}/schemas/${schema}/tables/${table}/columns`),
    indexes: (connId, schema, table) => request(`/connections/${connId}/schemas/${schema}/tables/${table}/indexes`),
    foreignKeys: (connId, schema, table) => request(`/connections/${connId}/schemas/${schema}/tables/${table}/fks`)
  },
  query: {
    execute: (connId, sql, database) => request(`/connections/${connId}/query`, {
      method: "POST",
      body: JSON.stringify({ sql, database })
    }),
    explain: (connId, sql, database) => request(`/connections/${connId}/explain`, {
      method: "POST",
      body: JSON.stringify({ sql, database })
    })
  },
  history: {
    list: (connId, limit) => request(`/connections/${connId}/history${limit ? `?limit=${limit}` : ""}`),
    clear: (connId) => request(`/connections/${connId}/history`, { method: "DELETE" }),
    saveFavorite: (connId, title, sql, database) => request(`/connections/${connId}/favorites`, {
      method: "POST",
      body: JSON.stringify({ title, sql, database })
    })
  }
};
function TreeNode_1($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { node, depth, toggleNode, expanded, onContextMenu } = $$props;
    let isExpanded = derived(() => expanded.has(node.id));
    let hasChildren = derived(() => node.children && node.children.length > 0);
    $$renderer2.push(`<div class="tree-node svelte-8rtq2n" tabindex="0"${attr_style("", { "padding-left": `${stringify(depth * 16 + 8)}px` })}><span class="tree-expand svelte-8rtq2n">`);
    if (hasChildren()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`${escape_html(isExpanded() ? "▾" : "▸")}`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`·`);
    }
    $$renderer2.push(`<!--]--></span> <span class="tree-icon svelte-8rtq2n">${escape_html(node.icon)}</span> <span class="tree-label svelte-8rtq2n">${escape_html(node.label)}</span> `);
    if (node.loading) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="tree-spinner svelte-8rtq2n">⟳</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (isExpanded() && node.children) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(node.children);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let child = each_array[$$index];
        TreeNode($$renderer2, {
          node: child,
          depth: depth + 1,
          toggleNode,
          expanded,
          onContextMenu
        });
      }
      $$renderer2.push(`<!--]-->`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}
function Sidebar($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let {
      onEditConnection,
      onCreateTable,
      onDesignTable,
      onViewDDL,
      onContextMenu
    } = $$props;
    let searchQuery = "";
    async function loadConnections() {
      try {
        const conns = await api.connections.list();
        store_set(connections, conns);
        store_set(explorerTree, conns.map((c) => ({
          id: `conn-${c.id}`,
          label: c.name,
          icon: c.driver === "mysql" ? "🐬" : c.driver === "postgres" ? "🐘" : "📁",
          type: "connection",
          children: [],
          loaded: false,
          loading: false,
          connectionId: c.id
        })));
      } catch (e) {
        console.error("Failed to load connections:", e);
      }
    }
    async function toggleNode(node) {
      const expanded = new Set(store_get($$store_subs ??= {}, "$expandedNodes", expandedNodes));
      if (expanded.has(node.id)) {
        expanded.delete(node.id);
        store_set(expandedNodes, expanded);
        return;
      }
      expanded.add(node.id);
      store_set(expandedNodes, expanded);
      if (!node.loaded && !node.loading) {
        await loadChildren(node);
      }
    }
    async function loadChildren(node) {
      if (node.type === "connection" && node.connectionId) {
        node.loading = true;
        try {
          const dbs = await api.metadata.databases(node.connectionId);
          node.children = dbs.map((db) => ({
            id: `${node.id}-db-${db.name}`,
            label: db.name,
            icon: "🗄️",
            type: "database",
            children: [],
            loaded: false,
            loading: false,
            connectionId: node.connectionId,
            database: db.name
          }));
          node.loaded = true;
        } catch (e) {
          console.error(e);
        } finally {
          node.loading = false;
        }
        store_set(explorerTree, [
          ...store_get($$store_subs ??= {}, "$explorerTree", explorerTree)
        ]);
      }
      if (node.type === "database" && node.connectionId) {
        node.loading = true;
        try {
          const tables = await api.metadata.tables(node.connectionId, node.database || "");
          node.children = tables.map((t) => ({
            id: `${node.id}-tbl-${t.name}`,
            label: t.name,
            icon: t.type === "VIEW" ? "👁️" : "📋",
            type: "table",
            children: [],
            loaded: false,
            loading: false,
            connectionId: node.connectionId,
            database: node.database,
            schema: node.database,
            tableName: t.name
          }));
          node.loaded = true;
        } catch (e) {
          console.error(e);
        } finally {
          node.loading = false;
        }
        store_set(explorerTree, [
          ...store_get($$store_subs ??= {}, "$explorerTree", explorerTree)
        ]);
      }
      if (node.type === "table" && node.connectionId) {
        node.loading = true;
        try {
          const [columns, indexes] = await Promise.all([
            api.metadata.columns(node.connectionId, node.schema || "", node.tableName || ""),
            api.metadata.indexes(node.connectionId, node.schema || "", node.tableName || "")
          ]);
          node.children = [
            {
              id: `${node.id}-cols`,
              label: `Columns (${columns.length})`,
              icon: "📂",
              type: "folder",
              children: columns.map((c) => ({
                id: `${node.id}-col-${c.name}`,
                label: `${c.name}  ${c.type}${c.isPrimary ? " 🔑" : ""}`,
                icon: c.isPrimary ? "🔑" : "📏",
                type: "column",
                connectionId: node.connectionId
              })),
              loaded: true
            },
            {
              id: `${node.id}-idxs`,
              label: `Indexes (${indexes.length})`,
              icon: "📂",
              type: "folder",
              children: indexes.map((idx) => ({
                id: `${node.id}-idx-${idx.name}`,
                label: `${idx.name} (${idx.columns.join(", ")})`,
                icon: idx.unique ? "💎" : "📇",
                type: "index",
                connectionId: node.connectionId
              })),
              loaded: true
            }
          ];
          node.loaded = true;
        } catch (e) {
          console.error(e);
        } finally {
          node.loading = false;
        }
        store_set(explorerTree, [
          ...store_get($$store_subs ??= {}, "$explorerTree", explorerTree)
        ]);
      }
    }
    function handleNodeContextMenu(node, event) {
      event.preventDefault();
      event.stopPropagation();
      const items = [];
      if (node.type === "connection" && node.connectionId) {
        items.push({
          label: "Edit Connection",
          icon: "✏️",
          action: () => onEditConnection(node.connectionId)
        });
        items.push({
          label: "Delete Connection",
          icon: "🗑️",
          action: async () => {
            if (confirm("Delete this connection?")) {
              await api.connections.delete(node.connectionId);
              loadConnections();
            }
          },
          danger: true
        });
      }
      if (node.type === "database" && node.connectionId) {
        items.push({
          label: "New Table",
          icon: "➕",
          action: () => onCreateTable(node.connectionId, node.database)
        });
        items.push({ separator: true, label: "" });
        items.push({
          label: "Refresh",
          icon: "↻",
          action: () => {
            node.loaded = false;
            loadChildren(node);
          }
        });
      }
      if (node.type === "table" && node.connectionId) {
        items.push({
          label: "Design Table",
          icon: "🔧",
          action: () => onDesignTable(node.connectionId, node.schema, node.tableName)
        });
        items.push({
          label: "View DDL",
          icon: "📜",
          action: () => onViewDDL(node.connectionId, node.schema, node.tableName)
        });
        items.push({ separator: true, label: "" });
        items.push({
          label: "Refresh",
          icon: "↻",
          action: () => {
            node.loaded = false;
            loadChildren(node);
          }
        });
        items.push({ separator: true, label: "" });
        items.push({
          label: "Drop Table",
          icon: "🗑️",
          action: async () => {
            if (confirm(`Drop table "${node.tableName}"? This cannot be undone.`)) {
              try {
                await fetch(`/api/connections/${node.connectionId}/schemas/${node.schema}/tables/${node.tableName}`, { method: "DELETE" });
                loadConnections();
              } catch (e) {
                alert("Failed to drop table: " + e.message);
              }
            }
          },
          danger: true
        });
      }
      if (items.length > 0) {
        onContextMenu(event.clientX, event.clientY, items);
      }
    }
    $$renderer2.push(`<div${attr_class("sidebar svelte-6dohdz", void 0, {
      "collapsed": store_get($$store_subs ??= {}, "$sidebarCollapsed", sidebarCollapsed)
    })}><div class="sidebar-header svelte-6dohdz"><span class="sidebar-title svelte-6dohdz">Explorer</span> <div class="sidebar-actions svelte-6dohdz"><button class="btn-icon" title="New Connection">⊕</button> <button class="btn-icon" title="Refresh">↻</button> <button class="btn-icon">${escape_html(store_get($$store_subs ??= {}, "$sidebarCollapsed", sidebarCollapsed) ? "▸" : "◂")}</button></div></div> `);
    if (!store_get($$store_subs ??= {}, "$sidebarCollapsed", sidebarCollapsed)) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="sidebar-search svelte-6dohdz"><input type="text" placeholder="Search objects..."${attr("value", searchQuery)} class="svelte-6dohdz"/></div> <div class="sidebar-tree svelte-6dohdz">`);
      if (store_get($$store_subs ??= {}, "$explorerTree", explorerTree).length === 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="empty-state svelte-6dohdz"><p>No connections yet</p> <button class="btn-primary">New Connection</button></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<!--[-->`);
        const each_array = ensure_array_like(store_get($$store_subs ??= {}, "$explorerTree", explorerTree));
        for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
          let node = each_array[$$index];
          TreeNode_1($$renderer2, {
            node,
            depth: 0,
            toggleNode,
            expanded: store_get($$store_subs ??= {}, "$expandedNodes", expandedNodes),
            onContextMenu: handleNodeContextMenu
          });
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function TopBar($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    $$renderer2.push(`<header class="topbar svelte-11yu8dz"><div class="topbar-left svelte-11yu8dz"><div class="app-logo svelte-11yu8dz"><span class="logo-icon svelte-11yu8dz">⬡</span> <span class="logo-text svelte-11yu8dz">SQLMG</span></div></div> <div class="topbar-center svelte-11yu8dz"><button class="topbar-btn svelte-11yu8dz"><span class="icon">⊕</span> New Connection</button></div> <div class="topbar-right svelte-11yu8dz"><button class="btn-icon" title="Settings">⚙</button></div></header>`);
  });
}
function StatusBar($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let activeConn = derived(() => store_get($$store_subs ??= {}, "$activeConnectionId", activeConnectionId));
    let activeDb = derived(() => store_get($$store_subs ??= {}, "$activeDatabase", activeDatabase));
    let tabCount = derived(() => store_get($$store_subs ??= {}, "$tabs", tabs).length);
    $$renderer2.push(`<footer class="statusbar svelte-g9asya"><div class="statusbar-left svelte-g9asya">`);
    if (activeConn()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="status-item status-connected svelte-g9asya">● Connected</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<span class="status-item status-disconnected svelte-g9asya">○ Disconnected</span>`);
    }
    $$renderer2.push(`<!--]--> `);
    if (activeDb()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="status-item svelte-g9asya">DB: ${escape_html(activeDb())}</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <div class="statusbar-right svelte-g9asya"><span class="status-item svelte-g9asya">Tabs: ${escape_html(tabCount())}</span> <span class="status-item svelte-g9asya">v0.1.0</span></div></footer>`);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function SQLEditor($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let {
      value = "",
      onExecute,
      onExecuteCurrent,
      schemaTables = [],
      schemaColumns = {},
      placeholder = "Enter SQL query (Ctrl+Enter to execute)...",
      readOnly = false
    } = $$props;
    new Compartment();
    new Compartment();
    $$renderer2.push(`<div class="editor-wrapper svelte-1r1fdy1"><div class="editor-toolbar svelte-1r1fdy1"><button class="btn-primary btn-sm svelte-1r1fdy1" title="Execute current statement (Ctrl+Enter)">▶ Run</button> <button class="btn-secondary btn-sm svelte-1r1fdy1" title="Execute all (Ctrl+Shift+Enter)">▶▶ Run All</button> <button class="btn-secondary btn-sm svelte-1r1fdy1" title="Format SQL">⟳ Format</button></div> <div class="editor-container svelte-1r1fdy1"></div></div>`);
    bind_props($$props, { value });
  });
}
function ResultGrid($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let totalRows = 0;
    onDestroy(() => {
    });
    $$renderer2.push(`<div class="result-grid-wrapper svelte-vnlmq8"><div class="grid-toolbar svelte-vnlmq8"><div class="grid-toolbar-left svelte-vnlmq8"><button class="btn-icon" title="Copy selected">📋</button> <button class="btn-icon" title="Export CSV">💾</button></div> <div class="grid-toolbar-right svelte-vnlmq8">`);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <span class="grid-info svelte-vnlmq8">${escape_html(totalRows)} rows</span></div></div> <div class="grid-container ag-theme-alpine-dark svelte-vnlmq8"></div></div>`);
  });
}
function QueryHistory($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let histories = [];
    let searchQuery = "";
    let filtered = derived(() => histories);
    function formatTime(ts) {
      try {
        return new Date(ts).toLocaleString();
      } catch {
        return ts;
      }
    }
    function truncate(s, len = 100) {
      return s.length > len ? s.slice(0, len) + "..." : s;
    }
    $$renderer2.push(`<div class="history-panel svelte-1o4uuze"><div class="history-header svelte-1o4uuze"><span class="history-title svelte-1o4uuze">Query History</span> <div class="history-actions svelte-1o4uuze"><button class="btn-icon" title="Refresh">↻</button> <button class="btn-icon" title="Clear">🗑</button> <button class="btn-icon" title="Close">×</button></div></div> <div class="history-search svelte-1o4uuze"><input type="text" placeholder="Search history..."${attr("value", searchQuery)} class="svelte-1o4uuze"/></div> <div class="history-list svelte-1o4uuze">`);
    if (filtered().length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="history-empty svelte-1o4uuze">No query history</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(filtered());
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let item = each_array[$$index];
        $$renderer2.push(`<div class="history-item svelte-1o4uuze" role="button" tabindex="0"><div class="history-sql svelte-1o4uuze">${escape_html(truncate(item.sql))}</div> <div class="history-meta svelte-1o4uuze"><span${attr_class("history-status svelte-1o4uuze", void 0, {
          "ok": item.status === "success",
          "fail": item.status === "error"
        })}>${escape_html(item.status)}</span> `);
        if (item.duration_ms) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span>${escape_html(item.duration_ms)}ms</span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> `);
        if (item.row_count) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span>${escape_html(item.row_count)} rows</span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <span class="history-time svelte-1o4uuze">${escape_html(formatTime(item.created_at))}</span></div></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div></div>`);
  });
}
function WorkArea($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    var $$store_subs;
    let showHistory = null;
    let splitPos = 50;
    let tabResults = {};
    let tabErrors = {};
    let tabExecuting = {};
    let tabSchemaTables = {};
    let tabSchemaColumns = {};
    async function executeQuery(tabId, sql) {
      if (!sql.trim()) return;
      const tab = store_get($$store_subs ??= {}, "$tabs", tabs).find((t) => t.id === tabId);
      if (!tab) return;
      tabExecuting[tabId] = true;
      tabErrors[tabId] = "";
      delete tabResults[tabId];
      const connId = tab.connectionId || store_get($$store_subs ??= {}, "$activeConnectionId", activeConnectionId);
      if (!connId) {
        tabErrors[tabId] = "No connection selected. Please connect to a database first.";
        tabExecuting[tabId] = false;
        return;
      }
      try {
        const result = await api.query.execute(connId, sql, tab.database || store_get($$store_subs ??= {}, "$activeDatabase", activeDatabase));
        tabResults[tabId] = {
          columns: result.columns,
          rows: result.rows,
          total: result.total,
          durationMs: result.durationMs
        };
        const idx = store_get($$store_subs ??= {}, "$tabs", tabs).findIndex((t) => t.id === tabId);
        if (idx >= 0) {
          store_mutate($$store_subs ??= {}, "$tabs", tabs, store_get($$store_subs ??= {}, "$tabs", tabs)[idx].resultSet = result);
          store_mutate($$store_subs ??= {}, "$tabs", tabs, store_get($$store_subs ??= {}, "$tabs", tabs)[idx].durationMs = result.durationMs);
          store_mutate($$store_subs ??= {}, "$tabs", tabs, store_get($$store_subs ??= {}, "$tabs", tabs)[idx].error = void 0);
        }
      } catch (e) {
        tabErrors[tabId] = e.message || "Query execution failed";
        const idx = store_get($$store_subs ??= {}, "$tabs", tabs).findIndex((t) => t.id === tabId);
        if (idx >= 0) {
          store_mutate($$store_subs ??= {}, "$tabs", tabs, store_get($$store_subs ??= {}, "$tabs", tabs)[idx].error = e.message);
        }
      } finally {
        tabExecuting[tabId] = false;
      }
    }
    let activeTab = derived(() => store_get($$store_subs ??= {}, "$tabs", tabs).find((t) => t.id === store_get($$store_subs ??= {}, "$activeTabId", activeTabId)));
    let $$settled = true;
    let $$inner_renderer;
    function $$render_inner($$renderer3) {
      $$renderer3.push(`<div class="workarea svelte-pm1qvb">`);
      if (store_get($$store_subs ??= {}, "$tabs", tabs).length === 0) {
        $$renderer3.push("<!--[0-->");
        $$renderer3.push(`<div class="welcome svelte-pm1qvb"><div class="welcome-content svelte-pm1qvb"><div class="welcome-icon svelte-pm1qvb">⬡</div> <h1 class="svelte-pm1qvb">SQLMG</h1> <p class="svelte-pm1qvb">Database Management System</p> <div class="welcome-actions svelte-pm1qvb"><button class="btn-primary">📝 New Query</button></div> <div class="shortcuts svelte-pm1qvb"><div class="shortcut svelte-pm1qvb"><kbd class="svelte-pm1qvb">Ctrl</kbd>+<kbd class="svelte-pm1qvb">Enter</kbd> Execute current statement</div> <div class="shortcut svelte-pm1qvb"><kbd class="svelte-pm1qvb">Ctrl</kbd>+<kbd class="svelte-pm1qvb">Shift</kbd>+<kbd class="svelte-pm1qvb">Enter</kbd> Execute all</div> <div class="shortcut svelte-pm1qvb"><kbd class="svelte-pm1qvb">Ctrl</kbd>+<kbd class="svelte-pm1qvb">S</kbd> Save to favorites</div></div></div></div>`);
      } else {
        $$renderer3.push("<!--[-1-->");
        $$renderer3.push(`<div class="tab-bar svelte-pm1qvb"><!--[-->`);
        const each_array = ensure_array_like(store_get($$store_subs ??= {}, "$tabs", tabs));
        for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
          let tab = each_array[$$index];
          $$renderer3.push(`<div${attr_class("tab svelte-pm1qvb", void 0, {
            "active": store_get($$store_subs ??= {}, "$activeTabId", activeTabId) === tab.id
          })} role="tab" tabindex="0"><span class="tab-icon svelte-pm1qvb">📝</span> <span class="tab-title svelte-pm1qvb">${escape_html(tab.title)}</span> `);
          if (tabExecuting[tab.id]) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<span class="tab-spinner svelte-pm1qvb">⟳</span>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <span class="tab-close svelte-pm1qvb" role="button" tabindex="0">×</span></div>`);
        }
        $$renderer3.push(`<!--]--> <button class="tab-add svelte-pm1qvb">+</button></div> `);
        if (activeTab()) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="tab-content svelte-pm1qvb"${attr("key", activeTab().id)}>`);
          if (activeTab().type === "query") {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div class="query-panel svelte-pm1qvb"><div class="split-view svelte-pm1qvb"${attr_style(`--split-pos: ${stringify(splitPos)}%`)}><div class="editor-pane svelte-pm1qvb"${attr_style(`flex: ${stringify(splitPos)}; min-height: 120px;`)}>`);
            SQLEditor($$renderer3, {
              onExecuteCurrent: (sql) => executeQuery(activeTab().id, sql),
              onExecute: (sql) => executeQuery(activeTab().id, sql),
              schemaTables: tabSchemaTables[activeTab().id] || [],
              schemaColumns: tabSchemaColumns[activeTab().id] || {},
              get value() {
                return activeTab().sql;
              },
              set value($$value) {
                activeTab().sql = $$value;
                $$settled = false;
              }
            });
            $$renderer3.push(`<!----></div> <div class="split-handle svelte-pm1qvb" role="separator" tabindex="0"></div> <div class="result-pane svelte-pm1qvb"${attr_style(`flex: ${stringify(100 - splitPos)}; min-height: 100px;`)}>`);
            if (tabExecuting[activeTab().id]) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<div class="result-status executing svelte-pm1qvb"><span class="spinner svelte-pm1qvb">⟳</span> Executing query...</div>`);
            } else if (tabErrors[activeTab().id]) {
              $$renderer3.push("<!--[1-->");
              $$renderer3.push(`<div class="result-status error svelte-pm1qvb">❌ ${escape_html(tabErrors[activeTab().id])}</div>`);
            } else if (tabResults[activeTab().id]) {
              $$renderer3.push("<!--[2-->");
              $$renderer3.push(`<div class="result-status success svelte-pm1qvb">✅ ${escape_html(tabResults[activeTab().id].total)} rows · ${escape_html(tabResults[activeTab().id].durationMs)}ms</div>`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<div class="result-status empty svelte-pm1qvb">Execute a query to see results</div>`);
            }
            $$renderer3.push(`<!--]--> `);
            if (tabResults[activeTab().id]) {
              $$renderer3.push("<!--[0-->");
              ResultGrid($$renderer3, {
                columns: tabResults[activeTab().id].columns,
                rows: tabResults[activeTab().id].rows
              });
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<div class="result-placeholder svelte-pm1qvb"><div class="placeholder-icon svelte-pm1qvb">📊</div> <p class="svelte-pm1qvb">Query results will appear here</p></div>`);
            }
            $$renderer3.push(`<!--]--></div></div> `);
            if (showHistory === activeTab().id) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<div class="history-drawer svelte-pm1qvb">`);
              QueryHistory($$renderer3, {
                connectionId: activeTab().connectionId || store_get($$store_subs ??= {}, "$activeConnectionId", activeConnectionId) || ""
              });
              $$renderer3.push(`<!----></div>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--> <div class="query-footer svelte-pm1qvb"><div class="footer-left svelte-pm1qvb">`);
            if (activeTab().connectionId) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<span class="footer-badge svelte-pm1qvb">🔗 ${escape_html(activeTab().connectionId.slice(0, 8))}</span>`);
            } else if (store_get($$store_subs ??= {}, "$activeConnectionId", activeConnectionId)) {
              $$renderer3.push("<!--[1-->");
              $$renderer3.push(`<span class="footer-badge svelte-pm1qvb">🔗 ${escape_html(store_get($$store_subs ??= {}, "$activeConnectionId", activeConnectionId).slice(0, 8))}</span>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--> `);
            if (activeTab().database || store_get($$store_subs ??= {}, "$activeDatabase", activeDatabase)) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<span class="footer-badge svelte-pm1qvb">🗄️ ${escape_html(activeTab().database || store_get($$store_subs ??= {}, "$activeDatabase", activeDatabase))}</span>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--></div> <div class="footer-right svelte-pm1qvb"><button class="btn-icon footer-btn svelte-pm1qvb" title="Query History">🕐</button></div></div></div>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
      $$renderer3.push(`<!--]--></div>`);
    }
    do {
      $$settled = true;
      $$inner_renderer = $$renderer2.copy();
      $$render_inner($$inner_renderer);
    } while (!$$settled);
    $$renderer2.subsume($$inner_renderer);
    if ($$store_subs) unsubscribe_stores($$store_subs);
  });
}
function ConnectionModal($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { connectionId = null } = $$props;
    let form = {
      name: "",
      driver: "mysql",
      host: "127.0.0.1",
      port: 3306,
      username: "root",
      password: "",
      database: ""
    };
    let testing = false;
    const driverDefaults = {
      mysql: { port: 3306, host: "127.0.0.1" },
      postgres: { port: 5432, host: "127.0.0.1" },
      sqlite: { port: 0, host: "" }
    };
    function onDriverChange() {
      const defaults = driverDefaults[form.driver];
      if (defaults) {
        form.port = defaults.port;
        form.host = defaults.host;
      }
    }
    $$renderer2.push(`<div class="overlay"><div class="modal"><div class="modal-header"><h2>${escape_html(connectionId ? "Edit Connection" : "New Connection")}</h2> <button class="btn-icon">×</button></div> <div class="modal-body"><div class="form-group"><label>Connection Name</label> <input type="text"${attr("value", form.name)} placeholder="My Database"/></div> <div class="form-group"><label>Driver</label> `);
    $$renderer2.select({ value: form.driver, onchange: onDriverChange }, ($$renderer3) => {
      $$renderer3.option({ value: "mysql" }, ($$renderer4) => {
        $$renderer4.push(`MySQL`);
      });
      $$renderer3.option({ value: "postgres" }, ($$renderer4) => {
        $$renderer4.push(`PostgreSQL`);
      });
      $$renderer3.option({ value: "sqlite" }, ($$renderer4) => {
        $$renderer4.push(`SQLite`);
      });
    });
    $$renderer2.push(`</div> `);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="form-row"><div class="form-group"><label>Host</label> <input type="text"${attr("value", form.host)}/></div> <div class="form-group"><label>Port</label> <input type="number"${attr("value", form.port)}/></div></div> <div class="form-row"><div class="form-group"><label>Username</label> <input type="text"${attr("value", form.username)}/></div> <div class="form-group"><label>Password</label> <input type="password"${attr("value", form.password)}/></div></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="form-group"><label>Database</label> <input type="text"${attr("value", form.database)} placeholder="(optional)"/></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <div class="modal-footer"><button class="btn-secondary"${attr("disabled", testing, true)}>${escape_html("Test Connection")}</button> <button class="btn-primary"${attr("disabled", true, true)}>${escape_html("Save")}</button> <button class="btn-secondary">Cancel</button></div></div></div>`);
  });
}
function TableDesigner($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { tableName, mode } = $$props;
    const SQL_TYPES = [
      "INTEGER",
      "BIGINT",
      "SMALLINT",
      "TINYINT",
      "INT",
      "VARCHAR",
      "CHAR",
      "TEXT",
      "MEDIUMTEXT",
      "LONGTEXT",
      "DECIMAL",
      "FLOAT",
      "DOUBLE",
      "NUMERIC",
      "DATE",
      "DATETIME",
      "TIMESTAMP",
      "TIME",
      "BOOLEAN",
      "BIT",
      "JSON",
      "JSONB",
      "UUID",
      "BLOB",
      "BINARY",
      "VARBINARY",
      "ENUM",
      "SERIAL",
      "BIGSERIAL"
    ];
    let activeTab = "columns";
    let localTableName = tableName ?? "";
    let tableComment = "";
    let columns = [];
    let validColumns = derived(() => columns.filter((c) => c.name.trim() !== ""));
    let canSave = derived(() => validColumns().length > 0 && localTableName.trim() !== "");
    $$renderer2.push(`<div class="overlay"><div class="modal wide svelte-7ssgok"><div class="modal-header"><h2>${escape_html(mode === "create" ? "Create Table" : "Alter Table: " + tableName)}</h2> <button class="btn-icon">×</button></div> `);
    {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="modal-body">`);
      if (mode === "create") {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="form-group"><label>Table Name</label> <input type="text"${attr("value", localTableName)} placeholder="table_name"/></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> <div class="form-group"><label>Table Comment</label> <input type="text"${attr("value", tableComment)} placeholder="(optional)"/></div> <div class="tabs svelte-7ssgok"><button${attr_class("tab svelte-7ssgok", void 0, { "active": activeTab === "columns" })}>Columns</button> <button${attr_class("tab svelte-7ssgok", void 0, { "active": activeTab === "indexes" })}>Indexes</button> <button${attr_class("tab svelte-7ssgok", void 0, { "active": activeTab === "foreignKeys" })}>Foreign Keys</button></div> `);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="columns-toolbar svelte-7ssgok"><button class="btn-secondary btn-sm svelte-7ssgok">+ Add Column</button></div> <div class="columns-table-wrap svelte-7ssgok"><table class="columns-table svelte-7ssgok"><thead><tr><th class="th-narrow svelte-7ssgok"></th><th class="svelte-7ssgok">Name</th><th class="svelte-7ssgok">Type</th><th class="th-length svelte-7ssgok">Length</th><th class="th-check svelte-7ssgok">Nullable</th><th class="th-check svelte-7ssgok">PK</th><th class="th-check svelte-7ssgok">AI</th><th class="svelte-7ssgok">Default</th><th class="svelte-7ssgok">Comment</th><th class="th-narrow svelte-7ssgok"></th></tr></thead><tbody><!--[-->`);
        const each_array = ensure_array_like(columns);
        for (let i = 0, $$length = each_array.length; i < $$length; i++) {
          let col = each_array[i];
          $$renderer2.push(`<tr><td class="td-reorder svelte-7ssgok"><button class="btn-icon btn-tiny svelte-7ssgok"${attr("disabled", i === 0, true)} title="Move up">↑</button> <button class="btn-icon btn-tiny svelte-7ssgok"${attr("disabled", i === columns.length - 1, true)} title="Move down">↓</button></td><td class="svelte-7ssgok"><input type="text"${attr("value", col.name)} placeholder="column_name" class="svelte-7ssgok"/></td><td class="svelte-7ssgok">`);
          $$renderer2.select(
            { value: col.type, class: "" },
            ($$renderer3) => {
              $$renderer3.push(`<!--[-->`);
              const each_array_1 = ensure_array_like(SQL_TYPES);
              for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
                let t = each_array_1[$$index];
                $$renderer3.option({ value: t }, ($$renderer4) => {
                  $$renderer4.push(`${escape_html(t)}`);
                });
              }
              $$renderer3.push(`<!--]-->`);
            },
            "svelte-7ssgok"
          );
          $$renderer2.push(`</td><td class="svelte-7ssgok"><input type="text"${attr("value", col.length)} placeholder="—" class="svelte-7ssgok"/></td><td class="td-center svelte-7ssgok"><input type="checkbox"${attr("checked", col.nullable, true)} class="svelte-7ssgok"/></td><td class="td-center svelte-7ssgok"><input type="checkbox"${attr("checked", col.isPrimary, true)} class="svelte-7ssgok"/></td><td class="td-center svelte-7ssgok"><input type="checkbox"${attr("checked", col.autoIncrement, true)} class="svelte-7ssgok"/></td><td class="svelte-7ssgok"><input type="text"${attr("value", col.defaultValue)} placeholder="—" class="svelte-7ssgok"/></td><td class="svelte-7ssgok"><input type="text"${attr("value", col.comment)} placeholder="—" class="svelte-7ssgok"/></td><td class="td-center svelte-7ssgok"><button class="btn-icon btn-tiny btn-danger-text svelte-7ssgok" title="Remove">×</button></td></tr>`);
        }
        $$renderer2.push(`<!--]-->`);
        if (columns.length === 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<tr><td colspan="10" class="empty-row svelte-7ssgok">No columns defined. Click "Add Column" to begin.</td></tr>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></tbody></table></div>`);
      }
      $$renderer2.push(`<!--]--></div> <div class="modal-footer"><button class="btn-secondary"${attr("disabled", !canSave(), true)}>${escape_html("Preview SQL")}</button> <button class="btn-primary"${attr("disabled", !canSave(), true)}>${escape_html(mode === "create" ? "Create Table" : "Save Changes")}</button> <button class="btn-secondary">Cancel</button></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div>`);
  });
}
function DDLViewer($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { schema, tableName } = $$props;
    $$renderer2.push(`<div class="overlay"><div class="modal"><div class="modal-header"><h2>DDL: ${escape_html(schema)}.${escape_html(tableName)}</h2> <button class="btn-icon">×</button></div> <div class="modal-body">`);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="loading-state svelte-3lzu7h">Loading DDL...</div>`);
    }
    $$renderer2.push(`<!--]--></div> <div class="modal-footer"><button class="btn-secondary"${attr("disabled", true, true)}>${escape_html("Copy")}</button> <button class="btn-secondary">Close</button></div></div></div>`);
  });
}
function ContextMenu($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { x = 0, y = 0, items = [] } = $$props;
    $$renderer2.push(`<div class="context-menu svelte-8ua90b"${attr_style("", { left: `${stringify(x)}px`, top: `${stringify(y)}px` })}><!--[-->`);
    const each_array = ensure_array_like(items);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let item = each_array[$$index];
      if (item.separator) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="menu-separator svelte-8ua90b"></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<button${attr_class("menu-item svelte-8ua90b", void 0, { "danger": item.danger })}>`);
        if (item.icon) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="menu-icon svelte-8ua90b">${escape_html(item.icon)}</span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <span class="menu-label">${escape_html(item.label)}</span></button>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function _layout($$renderer, $$props) {
  let { children } = $$props;
  let showConnectionModal = false;
  let editingConnectionId = null;
  let showTableDesigner = false;
  let designerTableName = "";
  let designerMode = "create";
  let showDDLViewer = false;
  let ddlSchema = "";
  let ddlTableName = "";
  let contextMenu = null;
  function openEditConnection(id) {
    editingConnectionId = id;
    showConnectionModal = true;
  }
  function openCreateTable(connId, schema) {
    designerTableName = "";
    designerMode = "create";
    showTableDesigner = true;
  }
  function openDesignTable(connId, schema, tableName) {
    designerTableName = tableName;
    designerMode = "alter";
    showTableDesigner = true;
  }
  function openDDLViewer(connId, schema, tableName) {
    ddlSchema = schema;
    ddlTableName = tableName;
    showDDLViewer = true;
  }
  function showContextMenu(x, y, items) {
    contextMenu = { x, y, items };
  }
  function getActions() {
    return {
      openCreateTable,
      openDesignTable,
      openDDLViewer,
      showContextMenu
    };
  }
  $$renderer.push(`<div id="app">`);
  TopBar($$renderer);
  $$renderer.push(`<!----> <div class="main-content svelte-12qhfyh">`);
  Sidebar($$renderer, {
    onEditConnection: openEditConnection,
    onCreateTable: openCreateTable,
    onDesignTable: openDesignTable,
    onViewDDL: openDDLViewer,
    onContextMenu: showContextMenu
  });
  $$renderer.push(`<!----> `);
  WorkArea($$renderer);
  $$renderer.push(`<!----></div> `);
  StatusBar($$renderer);
  $$renderer.push(`<!----></div> `);
  if (showConnectionModal) {
    $$renderer.push("<!--[0-->");
    ConnectionModal($$renderer, {
      connectionId: editingConnectionId
    });
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> `);
  if (showTableDesigner) {
    $$renderer.push("<!--[0-->");
    TableDesigner($$renderer, {
      tableName: designerTableName,
      mode: designerMode
    });
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> `);
  if (showDDLViewer) {
    $$renderer.push("<!--[0-->");
    DDLViewer($$renderer, {
      schema: ddlSchema,
      tableName: ddlTableName
    });
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> `);
  if (contextMenu) {
    $$renderer.push("<!--[0-->");
    ContextMenu($$renderer, {
      x: contextMenu.x,
      y: contextMenu.y,
      items: contextMenu.items
    });
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]-->`);
  bind_props($$props, { getActions });
}
export {
  _layout as default
};
