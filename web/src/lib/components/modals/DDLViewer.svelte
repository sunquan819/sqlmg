<script lang="ts">
	import { onMount } from 'svelte';

	let {
		connectionId,
		schema,
		tableName,
		onClose
	}: {
		connectionId: string;
		schema: string;
		tableName: string;
		onClose: () => void;
	} = $props();

	let ddl = $state('');
	let loading = $state(true);
	let error = $state('');
	let copied = $state(false);

	async function loadDDL() {
		loading = true;
		error = '';
		try {
			const resp = await fetch(`/api/connections/${connectionId}/schemas/${schema}/tables/${tableName}/ddl`);
			const data = await resp.json();
			if (data.error) {
				error = data.error;
			} else {
				ddl = data.ddl || '';
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function copyDDL() {
		navigator.clipboard.writeText(ddl);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	$effect(() => { loadDDL(); });
</script>

<div class="overlay" onclick={onClose}>
	<div class="modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>DDL: {schema}.{tableName}</h2>
			<button class="btn-icon" onclick={onClose}>×</button>
		</div>

		<div class="modal-body">
			{#if loading}
				<div class="loading-state">Loading DDL...</div>
			{:else if error}
				<div class="error-state">{error}</div>
			{:else}
				<pre class="ddl-viewer">{ddl}</pre>
			{/if}
		</div>

		<div class="modal-footer">
			<button class="btn-secondary" onclick={copyDDL} disabled={!ddl}>
				{copied ? 'Copied!' : 'Copy'}
			</button>
			<button class="btn-secondary" onclick={onClose}>Close</button>
		</div>
	</div>
</div>

<style>
	.loading-state, .error-state {
		padding: 24px; text-align: center; color: var(--text-muted); font-size: 13px;
	}
	.error-state { color: var(--error); }
	.ddl-viewer {
		background: var(--bg-primary); border: 1px solid var(--border);
		border-radius: var(--radius); padding: 16px; overflow: auto;
		max-height: 60vh; font-family: var(--font-mono); font-size: 13px;
		line-height: 1.6; color: var(--text-primary); white-space: pre-wrap;
		word-break: break-all; margin: 0;
	}
</style>
