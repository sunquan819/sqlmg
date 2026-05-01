<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import TopBar from '$lib/components/layout/TopBar.svelte';
	import StatusBar from '$lib/components/layout/StatusBar.svelte';
	import WorkArea from '$lib/components/layout/WorkArea.svelte';
	import ConnectionModal from '$lib/components/modals/ConnectionModal.svelte';
	import TableDesigner from '$lib/components/modals/TableDesigner.svelte';
	import DDLViewer from '$lib/components/modals/DDLViewer.svelte';
	import ImportExportModal from '$lib/components/modals/ImportExportModal.svelte';
	import ContextMenu from '$lib/components/layout/ContextMenu.svelte';

	let { children } = $props();

	let showConnectionModal = $state(false);
	let editingConnectionId = $state<string | null>(null);

	let showTableDesigner = $state(false);
	let designerConnectionId = $state('');
	let designerSchema = $state('');
	let designerTableName = $state('');
	let designerMode = $state<'create' | 'alter'>('create');

	let showDDLViewer = $state(false);
	let ddlConnectionId = $state('');
	let ddlSchema = $state('');
	let ddlTableName = $state('');

	let showImportExport = $state(false);
	let ieMode = $state<'export' | 'import'>('export');
	let ieConnectionId = $state('');
	let ieSchema = $state('');
	let ieTableName = $state('');

	let contextMenu = $state<{ x: number; y: number; items: any[] } | null>(null);

	function openNewConnection() {
		editingConnectionId = null;
		showConnectionModal = true;
	}

	function openEditConnection(id: string) {
		editingConnectionId = id;
		showConnectionModal = true;
	}

	function closeConnectionModal() {
		showConnectionModal = false;
		editingConnectionId = null;
	}

	function openCreateTable(connId: string, schema: string) {
		designerConnectionId = connId;
		designerSchema = schema;
		designerTableName = '';
		designerMode = 'create';
		showTableDesigner = true;
	}

	function openDesignTable(connId: string, schema: string, tableName: string) {
		designerConnectionId = connId;
		designerSchema = schema;
		designerTableName = tableName;
		designerMode = 'alter';
		showTableDesigner = true;
	}

	function closeTableDesigner() {
		showTableDesigner = false;
	}

	function openDDLViewer(connId: string, schema: string, tableName: string) {
		ddlConnectionId = connId;
		ddlSchema = schema;
		ddlTableName = tableName;
		showDDLViewer = true;
	}

	function closeDDLViewer() {
		showDDLViewer = false;
	}

	function openExport(connId: string, schema: string, tableName: string) {
		ieConnectionId = connId;
		ieSchema = schema;
		ieTableName = tableName;
		ieMode = 'export';
		showImportExport = true;
	}

	function openImport(connId: string, schema: string, tableName: string) {
		ieConnectionId = connId;
		ieSchema = schema;
		ieTableName = tableName;
		ieMode = 'import';
		showImportExport = true;
	}

	function closeImportExport() {
		showImportExport = false;
	}

	function showContextMenu(x: number, y: number, items: any[]) {
		contextMenu = { x, y, items };
	}

	function closeContextMenu() {
		contextMenu = null;
	}
</script>

<div id="app">
	<TopBar onNewConnection={openNewConnection} />
	<div class="main-content">
		<Sidebar
			onNewConnection={openNewConnection}
			onEditConnection={openEditConnection}
			onCreateTable={openCreateTable}
			onDesignTable={openDesignTable}
			onViewDDL={openDDLViewer}
			onExport={openExport}
			onImport={openImport}
			onContextMenu={showContextMenu}
		/>
		<WorkArea />
	</div>
	<StatusBar />
</div>

{#if showConnectionModal}
	<ConnectionModal
		connectionId={editingConnectionId}
		onClose={closeConnectionModal}
	/>
{/if}

{#if showTableDesigner}
	<TableDesigner
		connectionId={designerConnectionId}
		schema={designerSchema}
		tableName={designerTableName}
		mode={designerMode}
		onClose={closeTableDesigner}
		onSaved={closeTableDesigner}
	/>
{/if}

{#if showDDLViewer}
	<DDLViewer
		connectionId={ddlConnectionId}
		schema={ddlSchema}
		tableName={ddlTableName}
		onClose={closeDDLViewer}
	/>
{/if}

{#if showImportExport}
	<ImportExportModal
		mode={ieMode}
		connectionId={ieConnectionId}
		schema={ieSchema}
		tableName={ieTableName}
		onClose={closeImportExport}
	/>
{/if}

{#if contextMenu}
	<ContextMenu
		x={contextMenu.x}
		y={contextMenu.y}
		items={contextMenu.items}
		onClose={closeContextMenu}
	/>
{/if}

<style>
	.main-content {
		display: flex;
		flex: 1;
		overflow: hidden;
	}
</style>
