<script lang="ts">
	import { connections, explorerTree, expandedNodes, sidebarCollapsed } from '$lib/stores';
	import { api } from '$lib/api/client';
	import { t } from '$lib/i18n';
	import TreeNode from './TreeNode.svelte';

	let {
		onNewConnection,
		onEditConnection,
		onCreateTable,
		onDesignTable,
		onViewDDL,
		onExport,
		onImport,
		onContextMenu
	}: {
		onNewConnection: () => void;
		onEditConnection: (id: string) => void;
		onCreateTable: (connId: string, schema: string) => void;
		onDesignTable: (connId: string, schema: string, tableName: string) => void;
		onViewDDL: (connId: string, schema: string, tableName: string) => void;
		onExport: (connId: string, schema: string, tableName: string) => void;
		onImport: (connId: string, schema: string, tableName: string) => void;
		onContextMenu: (x: number, y: number, items: any[]) => void;
	} = $props();

	let searchQuery = $state('');

	async function loadConnections() {
		try {
			const conns = await api.connections.list();
			$connections = conns;
			$explorerTree = conns.map(c => ({
				id: `conn-${c.id}`,
				label: c.name,
				icon: c.driver === 'mysql' ? '🐬' : c.driver === 'postgres' ? '🐘' : '📁',
				type: 'connection' as const,
				children: [] as any[],
				loaded: false,
				loading: false,
				connectionId: c.id
			}));
		} catch (e) {
			console.error('Failed to load connections:', e);
		}
	}

	async function toggleNode(node: any) {
		const expanded = new Set($expandedNodes);
		if (expanded.has(node.id)) {
			expanded.delete(node.id);
			$expandedNodes = expanded;
			return;
		}
		expanded.add(node.id);
		$expandedNodes = expanded;

		if (!node.loaded && !node.loading) {
			await loadChildren(node);
		}
	}

	async function loadChildren(node: any) {
		if (node.type === 'connection' && node.connectionId) {
			node.loading = true;
			try {
				const dbs = await api.metadata.databases(node.connectionId);
				node.children = dbs.map((db: any) => ({
					id: `${node.id}-db-${db.name}`,
					label: db.name,
					icon: '🗄️',
					type: 'database',
					children: [],
					loaded: false,
					loading: false,
					connectionId: node.connectionId,
					database: db.name
				}));
				node.loaded = true;
			} catch (e) { console.error(e); }
			finally { node.loading = false; }
			$explorerTree = [...$explorerTree];
		}

		if (node.type === 'database' && node.connectionId) {
			node.loading = true;
			try {
				const tables = await api.metadata.tables(node.connectionId, node.database || '');
				node.children = tables.map((t: any) => ({
					id: `${node.id}-tbl-${t.name}`,
					label: t.name,
					icon: t.type === 'VIEW' ? '👁️' : '📋',
					type: 'table',
					children: [],
					loaded: false,
					loading: false,
					connectionId: node.connectionId,
					database: node.database,
					schema: node.database,
					tableName: t.name
				}));
				node.loaded = true;
			} catch (e) { console.error(e); }
			finally { node.loading = false; }
			$explorerTree = [...$explorerTree];
		}

		if (node.type === 'table' && node.connectionId) {
			node.loading = true;
			try {
				const [columns, indexes] = await Promise.all([
					api.metadata.columns(node.connectionId, node.schema || '', node.tableName || ''),
					api.metadata.indexes(node.connectionId, node.schema || '', node.tableName || '')
				]);
				node.children = [
					{
						id: `${node.id}-cols`,
						label: `Columns (${columns.length})`,
						icon: '📂',
						type: 'folder',
						children: columns.map((c: any) => ({
							id: `${node.id}-col-${c.name}`,
							label: `${c.name}  ${c.type}${c.isPrimary ? ' 🔑' : ''}`,
							icon: c.isPrimary ? '🔑' : '📏',
							type: 'column',
							connectionId: node.connectionId
						})),
						loaded: true
					},
					{
						id: `${node.id}-idxs`,
						label: `Indexes (${indexes.length})`,
						icon: '📂',
						type: 'folder',
						children: indexes.map((idx: any) => ({
							id: `${node.id}-idx-${idx.name}`,
							label: `${idx.name} (${idx.columns.join(', ')})`,
							icon: idx.unique ? '💎' : '📇',
							type: 'index',
							connectionId: node.connectionId
						})),
						loaded: true
					}
				];
				node.loaded = true;
			} catch (e) { console.error(e); }
			finally { node.loading = false; }
			$explorerTree = [...$explorerTree];
		}
	}

	function handleNodeContextMenu(node: any, event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();

		const items: any[] = [];

		if (node.type === 'connection' && node.connectionId) {
			items.push({ label: t('context.editConnection'), icon: '✏️', action: () => onEditConnection(node.connectionId) });
			items.push({ label: t('context.deleteConnection'), icon: '🗑️', action: async () => {
				if (confirm(t('connection.deleteConfirm'))) {
					await api.connections.delete(node.connectionId);
					loadConnections();
				}
			}, danger: true });
		}

		if (node.type === 'database' && node.connectionId) {
			items.push({ label: t('context.newTable'), icon: '➕', action: () => onCreateTable(node.connectionId, node.database) });
			items.push({ separator: true, label: '' });
			items.push({ label: t('common.refresh'), icon: '↻', action: () => { node.loaded = false; loadChildren(node); } });
		}

		if (node.type === 'table' && node.connectionId) {
			items.push({ label: t('context.designTable'), icon: '🔧', action: () => onDesignTable(node.connectionId, node.schema, node.tableName) });
			items.push({ label: t('context.viewDDL'), icon: '📜', action: () => onViewDDL(node.connectionId, node.schema, node.tableName) });
			items.push({ separator: true, label: '' });
			items.push({ label: t('export.title'), icon: '📤', action: () => onExport(node.connectionId, node.schema, node.tableName) });
			items.push({ label: t('import.title'), icon: '📥', action: () => onImport(node.connectionId, node.schema, node.tableName) });
			items.push({ separator: true, label: '' });
			items.push({ label: t('common.refresh'), icon: '↻', action: () => { node.loaded = false; loadChildren(node); } });
			items.push({ separator: true, label: '' });
			items.push({ label: t('context.dropTable'), icon: '🗑️', action: async () => {
				if (confirm(t('context.dropTableConfirm').replace('{name}', node.tableName))) {
					try {
						await fetch(`/api/connections/${node.connectionId}/schemas/${node.schema}/tables/${node.tableName}`, { method: 'DELETE' });
						loadConnections();
					} catch (e) { alert('Failed to drop table: ' + (e as any).message); }
				}
			}, danger: true });
		}

		if (items.length > 0) {
			onContextMenu(event.clientX, event.clientY, items);
		}
	}

	$effect(() => { loadConnections(); });
</script>

<div class="sidebar" class:collapsed={$sidebarCollapsed}>
	<div class="sidebar-header">
		<span class="sidebar-title">{t('explorer.title')}</span>
		<div class="sidebar-actions">
			<button class="btn-icon" title={t('explorer.newConnection')} onclick={onNewConnection}>⊕</button>
			<button class="btn-icon" title={t('common.refresh')} onclick={loadConnections}>↻</button>
			<button class="btn-icon" onclick={() => ($sidebarCollapsed = !$sidebarCollapsed)}>
				{$sidebarCollapsed ? '▸' : '◂'}
			</button>
		</div>
	</div>

	{#if !$sidebarCollapsed}
		<div class="sidebar-search">
			<input type="text" placeholder={t('explorer.search')} bind:value={searchQuery} />
		</div>
		<div class="sidebar-tree">
			{#if $explorerTree.length === 0}
				<div class="empty-state">
					<p>{t('explorer.empty')}</p>
					<button class="btn-primary" onclick={onNewConnection}>{t('explorer.newConnection')}</button>
				</div>
			{:else}
				{#each $explorerTree as node (node.id)}
					<TreeNode {node} depth={0} {toggleNode} expanded={$expandedNodes} onContextMenu={handleNodeContextMenu} />
				{/each}
			{/if}
		</div>
	{/if}
</div>

<style>
	.sidebar {
		width: 280px; min-width: 200px; max-width: 500px;
		background: var(--bg-secondary); border-right: 1px solid var(--border);
		display: flex; flex-direction: column; overflow: hidden; flex-shrink: 0;
	}
	.sidebar.collapsed { width: 40px; min-width: 40px; }
	.sidebar-header {
		display: flex; align-items: center; justify-content: space-between;
		padding: 8px 8px 8px 12px; border-bottom: 1px solid var(--border);
	}
	.sidebar-title {
		font-size: 11px; font-weight: 600; text-transform: uppercase;
		letter-spacing: 0.5px; color: var(--text-muted);
	}
	.sidebar-actions { display: flex; gap: 2px; }
	.sidebar-search { padding: 8px; }
	.sidebar-search input { width: 100%; font-size: 12px; padding: 5px 8px; }
	.sidebar-tree { flex: 1; overflow-y: auto; padding: 4px 0; }
	.empty-state {
		display: flex; flex-direction: column; align-items: center;
		padding: 32px 16px; gap: 12px; color: var(--text-muted);
	}
</style>
