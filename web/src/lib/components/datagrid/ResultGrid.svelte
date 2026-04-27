<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { createGrid, type GridApi, type GridOptions, type ColDef, type ICellEditorParams } from 'ag-grid-community';
	import 'ag-grid-community/styles/ag-grid.css';
	import 'ag-grid-community/styles/ag-theme-alpine.css';

	let {
		columns = [] as { name: string; type: string }[],
		rows = [] as Record<string, any>[],
		editable = true,
		onCellEdit
	}: {
		columns: { name: string; type: string }[];
		rows: Record<string, any>[];
		editable?: boolean;
		onCellEdit?: (row: Record<string, any>, field: string, newValue: any) => void;
	} = $props();

	let gridContainer: HTMLElement;
	let gridApi: GridApi | null = null;
	let selectedCount = $state(0);
	let currentRow = $state(0);
	let totalRows = $state(0);

	const columnDefs = $derived<ColDef[]>(columns.map(col => ({
		headerName: `${col.name}\n${col.type}`,
		field: col.name,
		editable: editable,
		sortable: true,
		filter: true,
		resizable: true,
		minWidth: 100,
		headerTooltip: `${col.name} (${col.type})`,
		autoHeaderHeight: true,
		wrapHeaderText: true,
		cellStyle: { fontFamily: 'var(--font-mono)', fontSize: '13px' },
		valueFormatter: (params: any) => {
			if (params.value === null) return 'NULL';
			if (params.value === undefined) return '';
			if (typeof params.value === 'object') return JSON.stringify(params.value);
			return String(params.value);
		}
	})));

	const defaultColDef: ColDef = {
		flex: 1,
		minWidth: 80,
		cellDataType: false
	};

	function onGridReady(params: any) {
		gridApi = params.api;
		updateRowInfo();
	}

	function updateRowInfo() {
		if (!gridApi) return;
		totalRows = gridApi.getDisplayedRowCount();
	}

	function onSelectionChanged() {
		if (!gridApi) return;
		const selected = gridApi.getSelectedRows();
		selectedCount = selected.length;
	}

	function onCellEditingStopped(event: any) {
		if (event.newValue !== event.oldValue && onCellEdit) {
			onCellEdit(event.data, event.colDef.field, event.newValue);
		}
	}

	function exportCSV() {
		if (!gridApi) return;
		gridApi.exportDataAsCsv({
			fileName: 'query_result.csv',
			allColumns: true
		});
	}

	function copySelection() {
		if (!gridApi) return;
		gridApi.copySelectedRowsToClipboard({ includeHeaders: true });
	}

	$effect(() => {
		if (gridApi && columns.length > 0) {
			gridApi.setGridOption('columnDefs', columnDefs);
			gridApi.setGridOption('rowData', rows);
			updateRowInfo();
		}
	});

	onMount(() => {
		if (!gridContainer) return;

		const gridOptions: GridOptions = {
			columnDefs: columnDefs,
			rowData: rows,
			defaultColDef,
			rowSelection: 'multiple',
			suppressRowClickSelection: true,
			enableCellTextSelection: true,
			ensureDomOrder: true,
			onGridReady,
			onSelectionChanged,
			onCellEditingStopped,
			onFirstDataRendered: () => updateRowInfo(),
			navigateToNextCell: (params: any) => {
				const suggested = params.nextCellPosition;
				const key = params.key;
				if (suggested) {
					if (key === 'Tab') {
						return suggested;
					}
				}
				return suggested;
			}
		};

		const api = createGrid(gridContainer, gridOptions, {});
	});

	onDestroy(() => {
		if (gridApi) {
			gridApi.destroy();
		}
	});
</script>

<div class="result-grid-wrapper">
	<div class="grid-toolbar">
		<div class="grid-toolbar-left">
			<button class="btn-icon" title="Copy selected" onclick={copySelection}>📋</button>
			<button class="btn-icon" title="Export CSV" onclick={exportCSV}>💾</button>
		</div>
		<div class="grid-toolbar-right">
			{#if selectedCount > 0}
				<span class="grid-info">{selectedCount} selected</span>
			{/if}
			<span class="grid-info">{totalRows} rows</span>
		</div>
	</div>
	<div class="grid-container ag-theme-alpine-dark" bind:this={gridContainer}></div>
</div>

<style>
	.result-grid-wrapper {
		display: flex; flex-direction: column; height: 100%;
	}
	.grid-toolbar {
		display: flex; align-items: center; justify-content: space-between;
		padding: 4px 12px; border-bottom: 1px solid var(--border);
		flex-shrink: 0; background: var(--bg-secondary);
	}
	.grid-toolbar-left, .grid-toolbar-right {
		display: flex; align-items: center; gap: 4px;
	}
	.grid-info {
		font-size: 11px; color: var(--text-muted);
	}
	.grid-container {
		flex: 1; overflow: hidden;
	}

	:global(.ag-theme-alpine-dark) {
		--ag-background-color: var(--bg-primary);
		--ag-header-background-color: var(--bg-secondary);
		--ag-odd-row-background-color: rgba(49, 50, 68, 0.3);
		--ag-row-hover-color: var(--bg-hover);
		--ag-selected-row-background-color: rgba(137, 180, 250, 0.15);
		--ag-range-selection-border-color: var(--accent);
		--ag-border-color: var(--border);
		--ag-header-foreground-color: var(--text-secondary);
		--ag-foreground-color: var(--text-primary);
		--ag-row-border-color: var(--border);
		--ag-font-size: 13px;
		--ag-font-family: var(--font-mono);
		--ag-grid-size: 4px;
	}
	:global(.ag-theme-alpine-dark .ag-header-cell-label) {
		white-space: pre-line;
	}
</style>
