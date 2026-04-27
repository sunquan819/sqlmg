<script lang="ts">
	let { node, depth, toggleNode, expanded, onContextMenu }: {
		node: any;
		depth: number;
		toggleNode: (node: any) => void;
		expanded: Set<string>;
		onContextMenu?: (node: any, event: MouseEvent) => void;
	} = $props();

	let isExpanded = $derived(expanded.has(node.id));
	let hasChildren = $derived(node.children && node.children.length > 0);
</script>

<div
	class="tree-node"
	style:padding-left="{depth * 16 + 8}px"
	onclick={() => toggleNode(node)}
	oncontextmenu={(e) => onContextMenu?.(node, e)}
	tabindex="0"
>
	<span class="tree-expand">
		{#if hasChildren}
			{isExpanded ? '▾' : '▸'}
		{:else}
			·
		{/if}
	</span>
	<span class="tree-icon">{node.icon}</span>
	<span class="tree-label">{node.label}</span>
	{#if node.loading}
		<span class="tree-spinner">⟳</span>
	{/if}
</div>

{#if isExpanded && node.children}
	{#each node.children as child (child.id)}
		<TreeNode node={child} depth={depth + 1} {toggleNode} {expanded} {onContextMenu} />
	{/each}
{/if}

<style>
	.tree-node {
		display: flex; align-items: center; gap: 4px;
		padding: 3px 8px 3px 0; cursor: pointer;
		user-select: none; white-space: nowrap; overflow: hidden;
	}
	.tree-node:hover { background: var(--bg-hover); }
	.tree-node:focus { background: var(--bg-hover); outline: none; }
	.tree-expand {
		width: 14px; text-align: center; font-size: 10px; color: var(--text-muted);
	}
	.tree-icon { font-size: 14px; }
	.tree-label {
		font-size: 13px; overflow: hidden; text-overflow: ellipsis;
	}
	.tree-spinner {
		animation: spin 1s linear infinite; font-size: 12px; color: var(--accent);
	}
	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
