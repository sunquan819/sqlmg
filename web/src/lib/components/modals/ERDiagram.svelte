<script lang="ts">
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n';

	let {
		connectionId,
		schema,
		onClose
	}: {
		connectionId: string;
		schema: string;
		onClose: () => void;
	} = $props();

	interface ERNode {
		id: string;
		name: string;
		type: string;
		schema: string;
		columns: { name: string; type: string; isPrimary: boolean }[];
		x?: number;
		y?: number;
		width?: number;
		height?: number;
	}

	interface EREdge {
		id: string;
		source: string;
		sourceCol: string;
		target: string;
		targetCol: string;
		label?: string;
	}

	interface ERGraph {
		nodes: ERNode[];
		edges: EREdge[];
	}

	let graph = $state<ERGraph>({ nodes: [], edges: [] });
	let loading = $state(true);
	let error = $state('');
	let selectedTable = $state<string | null>(null);
	let zoom = $state(1);
	let panX = $state(0);
	let panY = $state(0);

	let selectedNode = $derived(graph.nodes.find(n => n.name === selectedTable));

	const NODE_WIDTH = 180;
	const NODE_HEADER_HEIGHT = 28;
	const NODE_ROW_HEIGHT = 22;
	const NODE_PADDING = 4;

	async function loadERGraph() {
		loading = true;
		error = '';
		try {
			const resp = await fetch(`/api/connections/${connectionId}/schemas/${schema}/er`);
			const data = await resp.json();
			if (data.error) {
				error = data.error;
			} else {
				graph = data;
				await layoutGraph();
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function layoutGraph() {
		if (graph.nodes.length === 0) return;

		try {
			const elk = await import('elkjs');
			const elkInstance = new elk.default();

			const elkNodes = graph.nodes.map(n => {
				const height = NODE_HEADER_HEIGHT + n.columns.length * NODE_ROW_HEIGHT + NODE_PADDING * 2;
				return {
					id: n.id,
					width: NODE_WIDTH,
					height: height
				};
			});

			const elkEdges = graph.edges.map(e => ({
				id: e.id,
				source: e.source,
				target: e.target
			}));

			const layoutOptions = {
				'algorithm': 'org.eclipse.elk.layered',
				'nodePlacement.strategy': 'BRANDES_KOEPF',
				'layering.strategy': 'LONGEST_PATH',
				'spacing.nodeNode': 40,
				'spacing.layer': 60,
				'direction': 'RIGHT'
			};

			const result = await elkInstance.layout({
				id: 'root',
				layoutOptions,
				children: elkNodes,
				edges: elkEdges
			});

			if (result.children) {
				for (let i = 0; i < result.children.length; i++) {
					const elkNode = result.children[i];
					const originalNode = graph.nodes.find(n => n.id === elkNode.id);
					if (originalNode && elkNode.x !== undefined && elkNode.y !== undefined) {
						originalNode.x = elkNode.x;
						originalNode.y = elkNode.y;
						originalNode.width = elkNode.width;
						originalNode.height = elkNode.height;
					}
				}
			}

			graph = { ...graph };
		} catch (e) {
			console.error('ELK layout failed, using fallback:', e);
			fallbackLayout();
		}
	}

	function fallbackLayout() {
		const cols = Math.ceil(Math.sqrt(graph.nodes.length));
		let x = 0, y = 0, col = 0;

		for (const node of graph.nodes) {
			node.x = x;
			node.y = y;
			node.width = NODE_WIDTH;
			node.height = NODE_HEADER_HEIGHT + node.columns.length * NODE_ROW_HEIGHT + NODE_PADDING * 2;

			col++;
			if (col >= cols) {
				col = 0;
				x = 0;
				y += 300;
			} else {
				x += 250;
			}
		}

		graph = { ...graph };
	}

	function getNodeHeight(node: ERNode): number {
		return NODE_HEADER_HEIGHT + node.columns.length * NODE_ROW_HEIGHT + NODE_PADDING * 2;
	}

	function getColumnY(node: ERNode, colIndex: number): number {
		return (node.y || 0) + NODE_HEADER_HEIGHT + NODE_PADDING + colIndex * NODE_ROW_HEIGHT + NODE_ROW_HEIGHT / 2;
	}

	function handleWheel(e: WheelEvent) {
		e.preventDefault();
		const delta = e.deltaY > 0 ? -0.1 : 0.1;
		zoom = Math.max(0.3, Math.min(2, zoom + delta));
	}

	function handleMouseDown(e: MouseEvent) {
		if (e.button !== 0) return;
		const startX = e.clientX - panX;
		const startY = e.clientY - panY;

		function onMouseMove(e: MouseEvent) {
			panX = e.clientX - startX;
			panY = e.clientY - startY;
		}

		function onMouseUp() {
			document.removeEventListener('mousemove', onMouseMove);
			document.removeEventListener('mouseup', onMouseUp);
		}

		document.addEventListener('mousemove', onMouseMove);
		document.addEventListener('mouseup', onMouseUp);
	}

	function selectTable(tableName: string) {
		selectedTable = tableName;
	}

	function exportImage() {
		const svg = document.querySelector('.er-svg');
		if (!svg) return;

		const svgData = new XMLSerializer().serializeToString(svg);
		const canvas = document.createElement('canvas');
		const ctx = canvas.getContext('2d');
		const img = new Image();

		img.onload = () => {
			canvas.width = img.width * 2;
			canvas.height = img.height * 2;
			ctx?.scale(2, 2);
			ctx?.drawImage(img, 0, 0);

			const link = document.createElement('a');
			link.download = `er_diagram_${schema}.png`;
			link.href = canvas.toDataURL('image/png');
			link.click();
		};

		img.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svgData)));
	}

	$effect(() => {
		loadERGraph();
	});
</script>

<div class="overlay" onclick={onClose}>
	<div class="modal er-modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>{t('er.title')} - {schema}</h2>
			<div class="header-actions">
				<button class="btn-icon" title={t('er.exportImage')} onclick={exportImage}>📷</button>
				<button class="btn-icon" title={t('common.refresh')} onclick={loadERGraph}>↻</button>
				<button class="btn-icon" onclick={onClose}>×</button>
			</div>
		</div>

		<div class="modal-body">
			{#if loading}
				<div class="loading-state">{t('common.loading')}</div>
			{:else if error}
				<div class="error-state">{error}</div>
			{:else if graph.nodes.length === 0}
				<div class="empty-state">{t('er.noTables')}</div>
			{:else}
				<div class="er-controls">
					<button class="btn-secondary btn-sm" onclick={() => (zoom = Math.max(0.3, zoom - 0.1))}>−</button>
					<span class="zoom-label">{Math.round(zoom * 100)}%</span>
					<button class="btn-secondary btn-sm" onclick={() => (zoom = Math.min(2, zoom + 0.1))}>+</button>
					<button class="btn-secondary btn-sm" onclick={() => { zoom = 1; panX = 0; panY = 0; }}>{t('er.reset')}</button>
				</div>

				<div class="er-container" onwheel={handleWheel} onmousedown={handleMouseDown}>
					<svg
						class="er-svg"
						viewBox="0 0 2000 1500"
						style:transform="translate({panX}px, {panY}px) scale({zoom})"
					>
						<g class="edges">
							{#each graph.edges as edge (edge.id)}
								{@const sourceNode = graph.nodes.find(n => n.id === edge.source)}
								{@const targetNode = graph.nodes.find(n => n.id === edge.target)}
								{@const sourceY = sourceNode ? getColumnY(sourceNode, sourceNode.columns.findIndex(c => c.name === edge.sourceCol)) : 0}
								{@const targetY = targetNode ? getColumnY(targetNode, targetNode.columns.findIndex(c => c.name === edge.targetCol)) : 0}
								<path
									class="er-edge"
									d="M {(sourceNode?.x || 0) + NODE_WIDTH} {sourceY}
									   C {(sourceNode?.x || 0) + NODE_WIDTH + 30} {sourceY},
									     {(targetNode?.x || 0) - 30} {targetY},
									     {targetNode?.x || 0} {targetY}"
								/>
								<circle cx={sourceNode?.x || 0 + NODE_WIDTH} cy={sourceY} r="3" fill="var(--accent)" />
								<circle cx={targetNode?.x || 0} cy={targetY} r="3" fill="var(--accent)" />
							{/each}
						</g>

						<g class="nodes">
							{#each graph.nodes as node (node.id)}
								<g
									class="er-node"
									class:selected={selectedTable === node.name}
									transform="translate({node.x || 0}, {node.y || 0})"
									onclick={() => selectTable(node.name)}
								>
									<rect
										width={NODE_WIDTH}
										height={getNodeHeight(node)}
										rx="6"
										class="node-bg"
									/>
									<rect
										width={NODE_WIDTH}
										height={NODE_HEADER_HEIGHT}
										rx="6"
										class="node-header"
									/>
									<clipPath id="clip-{node.id}">
										<rect width={NODE_WIDTH - 8} height={getNodeHeight(node) - NODE_HEADER_HEIGHT - 4} x="4" y={NODE_HEADER_HEIGHT} />
									</clipPath>
									<text x="8" y="18" class="node-title">{node.name}</text>

									<g clip-path="url(clip-{node.id})">
										{#each node.columns as col, i (col.name)}
											<g transform="translate(0, {NODE_HEADER_HEIGHT + i * NODE_ROW_HEIGHT})">
												<rect width={NODE_WIDTH} height={NODE_ROW_HEIGHT} class="column-row" />
												<text x="8" y="14" class="column-name {col.isPrimary ? 'pk' : ''}">{col.name}</text>
												<text x={NODE_WIDTH - 8} y="14" class="column-type">{col.type}</text>
												{#if col.isPrimary}
													<text x="8" y="14" class="pk-icon">🔑</text>
												{/if}
											</g>
										{/each}
									</g>
								</g>
							{/each}
						</g>
					</svg>
				</div>

				{#if selectedTable}
					<div class="table-detail">
						<h3>{selectedTable}</h3>
						{#if selectedNode}
							<table class="detail-table">
								<thead>
									<tr>
										<th>{t('designer.type')}</th>
										<th>{t('designer.nullable')}</th>
										<th>{t('designer.primaryKey')}</th>
									</tr>
								</thead>
								<tbody>
									{#each selectedNode.columns as col (col.name)}
										<tr>
											<td>{col.type}</td>
											<td>{col.isPrimary ? '—' : '✓'}</td>
											<td>{col.isPrimary ? '🔑' : ''}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<div class="modal-footer">
			<button class="btn-secondary" onclick={onClose}>{t('common.close')}</button>
		</div>
	</div>
</div>

<style>
	.er-modal {
		min-width: 900px;
		max-width: 95vw;
		height: 85vh;
	}

	.loading-state, .error-state, .empty-state {
		padding: 32px;
		text-align: center;
		color: var(--text-muted);
	}
	.error-state { color: var(--error); }

	.header-actions {
		display: flex;
		gap: 4px;
	}

	.er-controls {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		border-bottom: 1px solid var(--border);
		flex-shrink: 0;
	}
	.zoom-label {
		font-size: 12px;
		color: var(--text-secondary);
		min-width: 40px;
		text-align: center;
	}

	.er-container {
		flex: 1;
		overflow: hidden;
		cursor: grab;
		background: var(--bg-primary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		margin: 8px;
	}

	.er-svg {
		width: 100%;
		height: 100%;
		overflow: visible;
	}

	.er-edge {
		fill: none;
		stroke: var(--accent);
		stroke-width: 1.5;
		opacity: 0.7;
	}

	.er-node {
		cursor: pointer;
	}
	.er-node:hover .node-bg {
		stroke: var(--accent);
	}
	.er-node.selected .node-bg {
		stroke: var(--success);
		stroke-width: 2;
	}

	.node-bg {
		fill: var(--bg-secondary);
		stroke: var(--border);
		stroke-width: 1;
	}
	.node-header {
		fill: var(--bg-tertiary);
	}

	.node-title {
		fill: var(--text-primary);
		font-size: 13px;
		font-weight: 600;
		font-family: var(--font-sans);
	}

	.column-row {
		fill: transparent;
	}
	.column-row:hover {
		fill: var(--bg-hover);
	}

	.column-name {
		fill: var(--text-primary);
		font-size: 12px;
		font-family: var(--font-mono);
	}
	.column-name.pk {
		fill: var(--success);
		font-weight: 600;
	}

	.column-type {
		fill: var(--text-muted);
		font-size: 11px;
		text-anchor: end;
		font-family: var(--font-mono);
	}

	.pk-icon {
		font-size: 10px;
	}

	.table-detail {
		position: absolute;
		right: 24px;
		bottom: 80px;
		width: 280px;
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 12px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.3);
	}
	.table-detail h3 {
		font-size: 14px;
		font-weight: 600;
		margin-bottom: 8px;
		color: var(--text-primary);
	}
	.detail-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 12px;
	}
	.detail-table th {
		text-align: left;
		color: var(--text-muted);
		padding: 4px 8px;
		border-bottom: 1px solid var(--border);
	}
	.detail-table td {
		padding: 4px 8px;
		color: var(--text-secondary);
	}
</style>