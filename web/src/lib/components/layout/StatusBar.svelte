<script lang="ts">
	import { tabs, activeConnectionId, activeDatabase } from '$lib/stores';
	import { t } from '$lib/i18n';

	let activeConn = $derived($activeConnectionId);
	let activeDb = $derived($activeDatabase);
	let tabCount = $derived($tabs.length);
</script>

<footer class="statusbar">
	<div class="statusbar-left">
		{#if activeConn}
			<span class="status-item status-connected">● {t('status.connected')}</span>
		{:else}
			<span class="status-item status-disconnected">○ {t('status.disconnected')}</span>
		{/if}
		{#if activeDb}
			<span class="status-item">DB: {activeDb}</span>
		{/if}
	</div>
	<div class="statusbar-right">
		<span class="status-item">Tabs: {tabCount}</span>
		<span class="status-item">v0.1.0</span>
	</div>
</footer>

<style>
	.statusbar {
		display: flex; align-items: center; justify-content: space-between;
		height: 24px; padding: 0 12px; background: var(--bg-secondary);
		border-top: 1px solid var(--border); font-size: 11px;
		color: var(--text-muted); user-select: none; flex-shrink: 0;
	}
	.statusbar-left, .statusbar-right {
		display: flex; align-items: center; gap: 12px;
	}
	.status-item { display: flex; align-items: center; gap: 4px; }
	.status-connected { color: var(--success); }
	.status-disconnected { color: var(--text-muted); }
</style>
