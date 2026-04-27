<script lang="ts">
	import { tabs, activeTabId, activeConnectionId, activeDatabase, activeSchema } from '$lib/stores';
	import { api, type ResultSet } from '$lib/api/client';
	import SQLEditor from '$lib/components/editor/SQLEditor.svelte';
	import ResultGrid from '$lib/components/datagrid/ResultGrid.svelte';
	import QueryHistory from '$lib/components/editor/QueryHistory.svelte';

	let showHistory = $state<string | null>(null);
	let splitPos = $state(50);

	interface TabResult {
		columns: { name: string; type: string }[];
		rows: Record<string, any>[];
		total: number;
		durationMs: number;
	}

	let tabResults = $state<Record<string, TabResult>>({});
	let tabErrors = $state<Record<string, string>>({});
	let tabExecuting = $state<Record<string, boolean>>({});
	let tabSchemaTables = $state<Record<string, string[]>>({});
	let tabSchemaColumns = $state<Record<string, Record<string, string[]>>>({});

	async function executeQuery(tabId: string, sql: string) {
		if (!sql.trim()) return;

		const tab = $tabs.find(t => t.id === tabId);
		if (!tab) return;

		tabExecuting[tabId] = true;
		tabErrors[tabId] = '';
		delete tabResults[tabId];

		const connId = tab.connectionId || $activeConnectionId;
		if (!connId) {
			tabErrors[tabId] = 'No connection selected. Please connect to a database first.';
			tabExecuting[tabId] = false;
			return;
		}

		try {
			const result = await api.query.execute(connId, sql, tab.database || $activeDatabase);
			tabResults[tabId] = {
				columns: result.columns,
				rows: result.rows,
				total: result.total,
				durationMs: result.durationMs
			};

			const idx = $tabs.findIndex(t => t.id === tabId);
			if (idx >= 0) {
				$tabs[idx].resultSet = result;
				$tabs[idx].durationMs = result.durationMs;
				$tabs[idx].error = undefined;
			}
		} catch (e: any) {
			tabErrors[tabId] = e.message || 'Query execution failed';
			const idx = $tabs.findIndex(t => t.id === tabId);
			if (idx >= 0) {
				$tabs[idx].error = e.message;
			}
		} finally {
			tabExecuting[tabId] = false;
		}
	}

	async function loadSchemaForTab(tabId: string) {
		const tab = $tabs.find(t => t.id === tabId);
		if (!tab || !tab.connectionId) return;

		try {
			const connId = tab.connectionId;
			const dbs = await api.metadata.databases(connId);
			const db = tab.database || (dbs.length > 0 ? dbs[0].name : '');
			if (!db) return;

			const tables = await api.metadata.tables(connId, db);
			const tableNames = tables.map(t => t.name);
			tabSchemaTables[tabId] = tableNames;

			const colMap: Record<string, string[]> = {};
			for (const t of tables.slice(0, 50)) {
				try {
					const cols = await api.metadata.columns(connId, db, t.name);
					colMap[t.name] = cols.map(c => c.name);
				} catch { /* skip */ }
			}
			tabSchemaColumns[tabId] = colMap;
		} catch (e) {
			console.error('Failed to load schema:', e);
		}
	}

	function closeTab(tabId: string, e: MouseEvent) {
		e.stopPropagation();
		$tabs = $tabs.filter(t => t.id !== tabId);
		delete tabResults[tabId];
		delete tabErrors[tabId];
		delete tabExecuting[tabId];
		if ($activeTabId === tabId) {
			$activeTabId = $tabs.length > 0 ? $tabs[$tabs.length - 1].id : null;
		}
	}

	function addQueryTab(connId?: string, db?: string, sql?: string) {
		const id = `tab-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
		const conn = connId || $activeConnectionId;
		$tabs = [...$tabs, {
			id,
			title: sql ? sql.slice(0, 20) + (sql.length > 20 ? '...' : '') : 'New Query',
			type: 'query' as const,
			connectionId: conn || '',
			database: db || $activeDatabase || '',
			schema: $activeSchema || '',
			sql: sql || ''
		}];
		$activeTabId = id;
		if (conn) {
			loadSchemaForTab(id);
		}
	}

	function handleCellEdit(row: Record<string, any>, field: string, newValue: any) {
		console.log('Cell edited:', { row, field, newValue });
	}

	function handleHistorySelect(tabId: string, sql: string) {
		const idx = $tabs.findIndex(t => t.id === tabId);
		if (idx >= 0) {
			$tabs[idx].sql = sql;
		}
		showHistory = null;
	}

	let activeTab = $derived($tabs.find(t => t.id === $activeTabId));
</script>

<div class="workarea">
	{#if $tabs.length === 0}
		<div class="welcome">
			<div class="welcome-content">
				<div class="welcome-icon">⬡</div>
				<h1>SQLMG</h1>
				<p>Database Management System</p>
				<div class="welcome-actions">
					<button class="btn-primary" onclick={() => addQueryTab()}>
						📝 New Query
					</button>
				</div>
				<div class="shortcuts">
					<div class="shortcut"><kbd>Ctrl</kbd>+<kbd>Enter</kbd> Execute current statement</div>
					<div class="shortcut"><kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Enter</kbd> Execute all</div>
					<div class="shortcut"><kbd>Ctrl</kbd>+<kbd>S</kbd> Save to favorites</div>
				</div>
			</div>
		</div>
	{:else}
		<div class="tab-bar">
			{#each $tabs as tab (tab.id)}
				<div
					class="tab"
					class:active={$activeTabId === tab.id}
					onclick={() => ($activeTabId = tab.id)}
					role="tab"
					tabindex="0"
				>
					<span class="tab-icon">📝</span>
					<span class="tab-title">{tab.title}</span>
					{#if tabExecuting[tab.id]}
						<span class="tab-spinner">⟳</span>
					{/if}
					<span class="tab-close" onclick={(e) => closeTab(tab.id, e)} role="button" tabindex="0">×</span>
				</div>
			{/each}
			<button class="tab-add" onclick={() => addQueryTab()}>+</button>
		</div>

		{#if activeTab}
			<div class="tab-content" key={activeTab.id}>
				{#if activeTab.type === 'query'}
					<div class="query-panel">
						<div class="split-view" style="--split-pos: {splitPos}%">
							<div class="editor-pane" style="flex: {splitPos}; min-height: 120px;">
								<SQLEditor
									bind:value={activeTab.sql}
									onExecuteCurrent={(sql) => executeQuery(activeTab.id, sql)}
									onExecute={(sql) => executeQuery(activeTab.id, sql)}
									schemaTables={tabSchemaTables[activeTab.id] || []}
									schemaColumns={tabSchemaColumns[activeTab.id] || {}}
								/>
							</div>

							<div class="split-handle" role="separator" tabindex="0"></div>

							<div class="result-pane" style="flex: {100 - splitPos}; min-height: 100px;">
								{#if tabExecuting[activeTab.id]}
									<div class="result-status executing">
										<span class="spinner">⟳</span> Executing query...
									</div>
								{:else if tabErrors[activeTab.id]}
									<div class="result-status error">
										❌ {tabErrors[activeTab.id]}
									</div>
								{:else if tabResults[activeTab.id]}
									<div class="result-status success">
										✅ {tabResults[activeTab.id].total} rows · {tabResults[activeTab.id].durationMs}ms
									</div>
								{:else}
									<div class="result-status empty">
										Execute a query to see results
									</div>
								{/if}

								{#if tabResults[activeTab.id]}
									<ResultGrid
										columns={tabResults[activeTab.id].columns}
										rows={tabResults[activeTab.id].rows}
										onCellEdit={handleCellEdit}
									/>
								{:else}
									<div class="result-placeholder">
										<div class="placeholder-icon">📊</div>
										<p>Query results will appear here</p>
									</div>
								{/if}
							</div>
						</div>

						{#if showHistory === activeTab.id}
							<div class="history-drawer">
								<QueryHistory
									connectionId={activeTab.connectionId || $activeConnectionId || ''}
									onSelect={(sql) => handleHistorySelect(activeTab.id, sql)}
									onClose={() => (showHistory = null)}
								/>
							</div>
						{/if}

						<div class="query-footer">
							<div class="footer-left">
								{#if activeTab.connectionId}
									<span class="footer-badge">🔗 {activeTab.connectionId.slice(0, 8)}</span>
								{:else if $activeConnectionId}
									<span class="footer-badge">🔗 {$activeConnectionId.slice(0, 8)}</span>
								{/if}
								{#if activeTab.database || $activeDatabase}
									<span class="footer-badge">🗄️ {activeTab.database || $activeDatabase}</span>
								{/if}
							</div>
							<div class="footer-right">
								<button class="btn-icon footer-btn" title="Query History"
									onclick={() => (showHistory = showHistory === activeTab.id ? null : activeTab.id)}>
									🕐
								</button>
							</div>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.workarea {
		flex: 1; display: flex; flex-direction: column; overflow: hidden;
		background: var(--bg-primary);
	}

	.welcome {
		flex: 1; display: flex; align-items: center; justify-content: center;
	}
	.welcome-content { text-align: center; }
	.welcome-icon { font-size: 64px; color: var(--accent); margin-bottom: 16px; }
	.welcome h1 { font-size: 32px; font-weight: 700; color: var(--text-primary); margin-bottom: 8px; }
	.welcome p { color: var(--text-muted); margin-bottom: 24px; }
	.welcome-actions { margin-bottom: 32px; }
	.shortcuts { display: flex; flex-direction: column; gap: 8px; align-items: center; }
	.shortcut { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-muted); }
	kbd {
		background: var(--bg-tertiary); border: 1px solid var(--border);
		border-radius: 4px; padding: 2px 6px; font-family: var(--font-mono); font-size: 11px;
	}

	.tab-bar {
		display: flex; align-items: center; background: var(--bg-secondary);
		border-bottom: 1px solid var(--border); overflow-x: auto; flex-shrink: 0;
	}
	.tab {
		display: flex; align-items: center; gap: 6px;
		padding: 7px 12px; font-size: 12px; color: var(--text-secondary);
		background: transparent; border-right: 1px solid var(--border);
		border-radius: 0; white-space: nowrap;
	}
	.tab:hover { background: var(--bg-tertiary); }
	.tab.active {
		background: var(--bg-primary); color: var(--text-primary);
		border-bottom: 2px solid var(--accent);
	}
	.tab-icon { font-size: 12px; }
	.tab-title { max-width: 150px; overflow: hidden; text-overflow: ellipsis; }
	.tab-close {
		margin-left: 4px; padding: 0 4px; border-radius: 3px;
		font-size: 14px; line-height: 1; opacity: 0.5;
		background: transparent; color: inherit; border: none; cursor: pointer;
	}
	.tab-close:hover { opacity: 1; background: var(--bg-hover); }
	.tab-spinner {
		animation: spin 1s linear infinite; font-size: 10px; color: var(--accent);
	}
	.tab-add {
		padding: 7px 10px; font-size: 14px; color: var(--text-muted);
		background: transparent; border-radius: 0;
	}
	.tab-add:hover { background: var(--bg-hover); color: var(--text-primary); }

	.tab-content { flex: 1; overflow: hidden; }

	.query-panel { height: 100%; display: flex; flex-direction: column; }

	.split-view {
		flex: 1; display: flex; flex-direction: column; overflow: hidden;
	}
	.editor-pane { overflow: hidden; display: flex; flex-direction: column; }
	.split-handle {
		height: 4px; background: var(--border); cursor: row-resize;
		flex-shrink: 0; transition: background var(--transition);
	}
	.split-handle:hover { background: var(--accent); }
	.result-pane { overflow: hidden; display: flex; flex-direction: column; }

	.result-status {
		padding: 6px 12px; font-size: 12px; flex-shrink: 0;
		border-bottom: 1px solid var(--border); background: var(--bg-secondary);
		display: flex; align-items: center; gap: 6px;
	}
	.result-status.executing { color: var(--accent); }
	.result-status.error { color: var(--error); background: rgba(243,139,168,0.06); }
	.result-status.success { color: var(--success); }
	.result-status.empty { color: var(--text-muted); font-style: italic; }

	.result-placeholder {
		flex: 1; display: flex; flex-direction: column;
		align-items: center; justify-content: center; color: var(--text-muted);
	}
	.placeholder-icon { font-size: 48px; margin-bottom: 12px; opacity: 0.4; }
	.result-placeholder p { font-size: 14px; }

	.spinner {
		animation: spin 1s linear infinite; display: inline-block;
	}

	.query-footer {
		display: flex; align-items: center; justify-content: space-between;
		padding: 4px 12px; border-top: 1px solid var(--border);
		background: var(--bg-secondary); flex-shrink: 0;
	}
	.footer-left, .footer-right { display: flex; align-items: center; gap: 6px; }
	.footer-badge {
		font-size: 11px; padding: 2px 8px; background: var(--bg-tertiary);
		border-radius: 10px; color: var(--text-secondary);
	}
	.footer-btn { width: 24px; height: 24px; font-size: 14px; }

	.history-drawer {
		height: 200px; border-top: 1px solid var(--border); flex-shrink: 0;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
