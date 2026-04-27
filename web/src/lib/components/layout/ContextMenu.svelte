<script lang="ts">
	let {
		x = 0,
		y = 0,
		items = [] as { label: string; icon?: string; action: () => void; danger?: boolean; separator?: boolean }[],
		onClose
	}: {
		x: number;
		y: number;
		items: { label: string; icon?: string; action: () => void; danger?: boolean; separator?: boolean }[];
		onClose: () => void;
	} = $props();

	function handleClick(item: typeof items[0]) {
		item.action();
		onClose();
	}
</script>

<svelte:window onclick={onClose} onkeydown={(e) => e.key === 'Escape' && onClose()} />

<div class="context-menu" style:left="{x}px" style:top="{y}px">
	{#each items as item (item.label)}
		{#if item.separator}
			<div class="menu-separator"></div>
		{:else}
			<button class="menu-item" class:danger={item.danger} onclick={() => handleClick(item)}>
				{#if item.icon}
					<span class="menu-icon">{item.icon}</span>
				{/if}
				<span class="menu-label">{item.label}</span>
			</button>
		{/if}
	{/each}
</div>

<style>
	.context-menu {
		position: fixed; z-index: 2000;
		background: var(--bg-secondary); border: 1px solid var(--border);
		border-radius: 8px; padding: 4px 0; min-width: 180px;
		box-shadow: 0 8px 32px rgba(0,0,0,0.4);
		animation: fadeIn 100ms ease;
	}
	@keyframes fadeIn {
		from { opacity: 0; transform: scale(0.95); }
		to { opacity: 1; transform: scale(1); }
	}
	.menu-item {
		display: flex; align-items: center; gap: 8px;
		width: 100%; padding: 7px 12px; font-size: 13px;
		color: var(--text-primary); background: transparent;
		text-align: left; border: none; cursor: pointer;
	}
	.menu-item:hover { background: var(--bg-hover); }
	.menu-item.danger { color: var(--error); }
	.menu-item.danger:hover { background: rgba(243,139,168,0.12); }
	.menu-icon { font-size: 14px; width: 20px; text-align: center; }
	.menu-separator {
		height: 1px; background: var(--border); margin: 4px 8px;
	}
</style>
