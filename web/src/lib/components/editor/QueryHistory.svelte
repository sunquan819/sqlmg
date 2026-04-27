<script lang="ts">
	import { api } from '$lib/api/client';

	let { connectionId, onSelect, onClose }: {
		connectionId: string;
		onSelect: (sql: string) => void;
		onClose: () => void;
	} = $props();

	let histories = $state<any[]>([]);
	let searchQuery = $state('');

	async function loadHistory() {
		try {
			const resp = await fetch(`/api/connections/${connectionId}/history`);
			if (resp.ok) {
				histories = await resp.json();
			}
		} catch (e) {
			console.error(e);
		}
	}

	async function clearHistory() {
		try {
			await fetch(`/api/connections/${connectionId}/history`, { method: 'DELETE' });
			histories = [];
		} catch (e) {
			console.error(e);
		}
	}

	let filtered = $derived(
		searchQuery
			? histories.filter(h => h.sql.toLowerCase().includes(searchQuery.toLowerCase()))
			: histories
	);

	$effect(() => { loadHistory(); });

	function formatTime(ts: string): string {
		try {
			return new Date(ts).toLocaleString();
		} catch { return ts; }
	}

	function truncate(s: string, len: number = 100): string {
		return s.length > len ? s.slice(0, len) + '...' : s;
	}
</script>

<div class="history-panel">
	<div class="history-header">
		<span class="history-title">Query History</span>
		<div class="history-actions">
			<button class="btn-icon" title="Refresh" onclick={loadHistory}>↻</button>
			<button class="btn-icon" title="Clear" onclick={clearHistory}>🗑</button>
			<button class="btn-icon" title="Close" onclick={onClose}>×</button>
		</div>
	</div>
	<div class="history-search">
		<input type="text" placeholder="Search history..." bind:value={searchQuery} />
	</div>
	<div class="history-list">
		{#if filtered.length === 0}
			<div class="history-empty">No query history</div>
		{:else}
			{#each filtered as item (item.id)}
				<div class="history-item" onclick={() => onSelect(item.sql)} role="button" tabindex="0">
					<div class="history-sql">{truncate(item.sql)}</div>
					<div class="history-meta">
						<span class="history-status" class:ok={item.status === 'success'} class:fail={item.status === 'error'}>
							{item.status}
						</span>
						{#if item.duration_ms}
							<span>{item.duration_ms}ms</span>
						{/if}
						{#if item.row_count}
							<span>{item.row_count} rows</span>
						{/if}
						<span class="history-time">{formatTime(item.created_at)}</span>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.history-panel {
		display: flex; flex-direction: column; height: 100%;
		background: var(--bg-secondary);
	}
	.history-header {
		display: flex; align-items: center; justify-content: space-between;
		padding: 8px 12px; border-bottom: 1px solid var(--border);
	}
	.history-title {
		font-size: 12px; font-weight: 600; text-transform: uppercase;
		letter-spacing: 0.5px; color: var(--text-muted);
	}
	.history-actions { display: flex; gap: 2px; }
	.history-search { padding: 8px; }
	.history-search input { width: 100%; font-size: 12px; padding: 5px 8px; }
	.history-list { flex: 1; overflow-y: auto; }
	.history-empty {
		padding: 24px; text-align: center; color: var(--text-muted); font-size: 13px;
	}
	.history-item {
		padding: 8px 12px; cursor: pointer; border-bottom: 1px solid var(--border);
		transition: background var(--transition);
	}
	.history-item:hover { background: var(--bg-hover); }
	.history-sql {
		font-family: var(--font-mono); font-size: 12px; color: var(--text-primary);
		white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
		margin-bottom: 4px;
	}
	.history-meta {
		display: flex; gap: 8px; font-size: 11px; color: var(--text-muted);
	}
	.history-status { font-weight: 600; }
	.history-status.ok { color: var(--success); }
	.history-status.fail { color: var(--error); }
	.history-time { margin-left: auto; }
</style>
