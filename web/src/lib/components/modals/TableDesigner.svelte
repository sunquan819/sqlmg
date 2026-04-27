<script lang="ts">
	import { api, type ColumnInfo, type IndexInfo, type ForeignKeyInfo } from '$lib/api/client';

	let {
		connectionId,
		schema,
		tableName,
		mode,
		onClose,
		onSaved
	} = $props<{
		connectionId: string;
		schema: string;
		tableName?: string;
		mode: 'create' | 'alter';
		onClose: () => void;
		onSaved: () => void;
	}>();

	interface ColumnDef {
		name: string;
		type: string;
		length: string;
		nullable: boolean;
		defaultValue: string;
		isPrimary: boolean;
		autoIncrement: boolean;
		comment: string;
	}

	interface IndexDef {
		name: string;
		unique: boolean;
		columns: string[];
	}

	interface ForeignKeyDef {
		name: string;
		columns: string[];
		refTable: string;
		refColumns: string;
		onDelete: string;
		onUpdate: string;
	}

	const SQL_TYPES = [
		'INTEGER', 'BIGINT', 'SMALLINT', 'TINYINT', 'INT',
		'VARCHAR', 'CHAR', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT',
		'DECIMAL', 'FLOAT', 'DOUBLE', 'NUMERIC',
		'DATE', 'DATETIME', 'TIMESTAMP', 'TIME',
		'BOOLEAN', 'BIT',
		'JSON', 'JSONB',
		'UUID', 'BLOB', 'BINARY', 'VARBINARY',
		'ENUM', 'SERIAL', 'BIGSERIAL'
	];

	const FK_ACTIONS = ['RESTRICT', 'CASCADE', 'SET NULL', 'NO ACTION'];

	let activeTab = $state<'columns' | 'indexes' | 'foreignKeys'>('columns');
	let localTableName = $state(tableName ?? '');
	let tableComment = $state('');
	let columns = $state<ColumnDef[]>([]);
	let indexes = $state<IndexDef[]>([]);
	let foreignKeys = $state<ForeignKeyDef[]>([]);
	let loading = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);
	let previewSql = $state<string | null>(null);
	let previewing = $state(false);

	let validColumns = $derived(columns.filter((c) => c.name.trim() !== ''));
	let canSave = $derived(validColumns.length > 0 && localTableName.trim() !== '');

	$effect(() => {
		if (mode === 'alter' && tableName) {
			loadExisting();
		}
	});

	async function loadExisting() {
		if (!tableName) return;
		loading = true;
		error = null;
		try {
			const [existingColumns, existingIndexes, existingFks] = await Promise.all([
				api.metadata.columns(connectionId, schema, tableName),
				api.metadata.indexes(connectionId, schema, tableName),
				api.metadata.foreignKeys(connectionId, schema, tableName)
			]);

			columns = existingColumns.map((c: ColumnInfo) => {
				const match = c.type.match(/^(\w+)\s*\(([^)]*)\)/i);
				return {
					name: c.name,
					type: match ? match[1].toUpperCase() : c.type.toUpperCase(),
					length: match ? match[2].trim() : '',
					nullable: c.nullable,
					defaultValue: c.defaultValue ?? '',
					isPrimary: c.isPrimary,
					autoIncrement: c.autoIncrement,
					comment: c.comment ?? ''
				};
			});

			indexes = existingIndexes.map((i: IndexInfo) => ({
				name: i.name,
				unique: i.unique,
				columns: [...i.columns]
			}));

			foreignKeys = existingFks.map((fk: ForeignKeyInfo) => ({
				name: fk.name,
				columns: [...fk.columns],
				refTable: fk.refTable,
				refColumns: fk.refColumns.join(', '),
				onDelete: fk.onDelete || 'RESTRICT',
				onUpdate: fk.onUpdate || 'RESTRICT'
			}));
		} catch (e: any) {
			error = e.message || 'Failed to load table structure';
		} finally {
			loading = false;
		}
	}

	function addColumn() {
		columns = [
			...columns,
			{ name: '', type: 'VARCHAR', length: '255', nullable: true, defaultValue: '', isPrimary: false, autoIncrement: false, comment: '' }
		];
	}

	function removeColumn(index: number) {
		columns = columns.filter((_, i) => i !== index);
	}

	function moveColumn(index: number, direction: -1 | 1) {
		const target = index + direction;
		if (target < 0 || target >= columns.length) return;
		const arr = [...columns];
		const temp = arr[index];
		arr[index] = arr[target];
		arr[target] = temp;
		columns = arr;
	}

	function onPrimaryKeyChange(col: ColumnDef) {
		if (!col.isPrimary) {
			col.autoIncrement = false;
		}
	}

	function onAutoIncrementChange(col: ColumnDef) {
		if (col.autoIncrement) {
			col.isPrimary = true;
			columns = columns.map((c) =>
				c === col ? c : { ...c, autoIncrement: false }
			);
		}
	}

	function addIndex() {
		indexes = [...indexes, { name: '', unique: false, columns: [] }];
	}

	function removeIndex(index: number) {
		indexes = indexes.filter((_, i) => i !== index);
	}

	function toggleIndexColumn(idx: number, colName: string) {
		const ix = indexes[idx];
		if (!ix) return;
		const pos = ix.columns.indexOf(colName);
		const newCols = pos >= 0 ? ix.columns.filter((c) => c !== colName) : [...ix.columns, colName];
		indexes = indexes.map((item, i) => i === idx ? { ...item, columns: newCols } : item);
	}

	function addForeignKey() {
		foreignKeys = [...foreignKeys, { name: '', columns: [], refTable: '', refColumns: '', onDelete: 'RESTRICT', onUpdate: 'RESTRICT' }];
	}

	function removeForeignKey(index: number) {
		foreignKeys = foreignKeys.filter((_, i) => i !== index);
	}

	function toggleFkColumn(idx: number, colName: string) {
		const fk = foreignKeys[idx];
		if (!fk) return;
		const pos = fk.columns.indexOf(colName);
		const newCols = pos >= 0 ? fk.columns.filter((c) => c !== colName) : [...fk.columns, colName];
		foreignKeys = foreignKeys.map((item, i) => i === idx ? { ...item, columns: newCols } : item);
	}

	function buildPayload() {
		return {
			database: '',
			schema,
			table: localTableName.trim(),
			columns: validColumns.map((c) => ({
				name: c.name,
				type: c.length ? c.type + '(' + c.length + ')' : c.type,
				nullable: c.nullable,
				defaultValue: c.defaultValue || null,
				isPrimary: c.isPrimary,
				autoIncrement: c.autoIncrement,
				comment: c.comment
			})),
			indexes: indexes
				.filter((i) => i.name.trim() && i.columns.length > 0)
				.map((i) => ({ name: i.name, unique: i.unique, columns: i.columns })),
			foreignKeys: foreignKeys
				.filter((fk) => fk.name.trim() && fk.columns.length > 0 && fk.refTable.trim())
				.map((fk) => ({
					name: fk.name,
					columns: fk.columns,
					refTable: fk.refTable,
					refColumns: fk.refColumns.split(',').map((s) => s.trim()).filter(Boolean),
					onDelete: fk.onDelete,
					onUpdate: fk.onUpdate
				})),
			comment: tableComment
		};
	}

	async function preview() {
		previewing = true;
		error = null;
		try {
			const resp = await fetch('/api/connections/' + connectionId + '/ddl/create/preview', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(buildPayload())
			});
			if (!resp.ok) {
				const err = await resp.json().catch(() => ({ error: resp.statusText }));
				throw new Error(err.error || 'HTTP ' + resp.status);
			}
			const data = await resp.json();
			previewSql = data.sql || data.preview || JSON.stringify(data, null, 2);
		} catch (e: any) {
			error = e.message || 'Preview failed';
		} finally {
			previewing = false;
		}
	}

	async function save() {
		if (!canSave) return;
		saving = true;
		error = null;
		try {
			const payload = buildPayload();
			const endpoint = mode === 'create'
				? '/api/connections/' + connectionId + '/ddl/create'
				: '/api/connections/' + connectionId + '/ddl/alter';

			const resp = await fetch(endpoint, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});

			if (!resp.ok) {
				const err = await resp.json().catch(() => ({ error: resp.statusText }));
				throw new Error(err.error || 'HTTP ' + resp.status);
			}

			onSaved();
			onClose();
		} catch (e: any) {
			error = e.message || 'Save failed';
		} finally {
			saving = false;
		}
	}
</script>

<div class="overlay" onclick={onClose}>
	<div class="modal wide" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>{mode === 'create' ? 'Create Table' : 'Alter Table: ' + tableName}</h2>
			<button class="btn-icon" onclick={onClose}>&times;</button>
		</div>

		{#if loading}
			<div class="modal-body">
				<div class="loading">Loading table structure&hellip;</div>
			</div>
		{:else}
			<div class="modal-body">
				{#if mode === 'create'}
					<div class="form-group">
						<label>Table Name</label>
						<input type="text" bind:value={localTableName} placeholder="table_name" />
					</div>
				{/if}

				<div class="form-group">
					<label>Table Comment</label>
					<input type="text" bind:value={tableComment} placeholder="(optional)" />
				</div>

				<div class="tabs">
					<button class="tab" class:active={activeTab === 'columns'} onclick={() => (activeTab = 'columns')}>Columns</button>
					<button class="tab" class:active={activeTab === 'indexes'} onclick={() => (activeTab = 'indexes')}>Indexes</button>
					<button class="tab" class:active={activeTab === 'foreignKeys'} onclick={() => (activeTab = 'foreignKeys')}>Foreign Keys</button>
				</div>

				{#if error}
					<div class="error-banner">{error}</div>
				{/if}

				{#if activeTab === 'columns'}
					<div class="columns-toolbar">
						<button class="btn-secondary btn-sm" onclick={addColumn}>+ Add Column</button>
					</div>
					<div class="columns-table-wrap">
						<table class="columns-table">
							<thead>
								<tr>
									<th class="th-narrow"></th>
									<th>Name</th>
									<th>Type</th>
									<th class="th-length">Length</th>
									<th class="th-check">Nullable</th>
									<th class="th-check">PK</th>
									<th class="th-check">AI</th>
									<th>Default</th>
									<th>Comment</th>
									<th class="th-narrow"></th>
								</tr>
							</thead>
							<tbody>
								{#each columns as col, i (i)}
									<tr>
										<td class="td-reorder">
											<button class="btn-icon btn-tiny" onclick={() => moveColumn(i, -1)} disabled={i === 0} title="Move up">&uarr;</button>
											<button class="btn-icon btn-tiny" onclick={() => moveColumn(i, 1)} disabled={i === columns.length - 1} title="Move down">&darr;</button>
										</td>
										<td><input type="text" bind:value={col.name} placeholder="column_name" /></td>
										<td>
											<select bind:value={col.type}>
												{#each SQL_TYPES as t}<option value={t}>{t}</option>{/each}
											</select>
										</td>
										<td><input type="text" bind:value={col.length} placeholder="&mdash;" /></td>
										<td class="td-center"><input type="checkbox" bind:checked={col.nullable} /></td>
										<td class="td-center"><input type="checkbox" bind:checked={col.isPrimary} onchange={() => onPrimaryKeyChange(col)} /></td>
										<td class="td-center"><input type="checkbox" bind:checked={col.autoIncrement} onchange={() => onAutoIncrementChange(col)} /></td>
										<td><input type="text" bind:value={col.defaultValue} placeholder="&mdash;" /></td>
										<td><input type="text" bind:value={col.comment} placeholder="&mdash;" /></td>
										<td class="td-center"><button class="btn-icon btn-tiny btn-danger-text" onclick={() => removeColumn(i)} title="Remove">&times;</button></td>
									</tr>
								{/each}
								{#if columns.length === 0}
									<tr><td colspan="10" class="empty-row">No columns defined. Click "Add Column" to begin.</td></tr>
								{/if}
							</tbody>
						</table>
					</div>
				{:else if activeTab === 'indexes'}
					<div class="cards-toolbar">
						<button class="btn-secondary btn-sm" onclick={addIndex}>+ Add Index</button>
					</div>
					{#if indexes.length === 0}
						<div class="empty-state">No indexes defined.</div>
					{/if}
					{#each indexes as idx, i (i)}
						<div class="card">
							<div class="card-header">
								<span class="card-title">Index #{i + 1}</span>
								<button class="btn-icon btn-tiny btn-danger-text" onclick={() => removeIndex(i)}>&times;</button>
							</div>
							<div class="card-body">
								<div class="form-row">
									<div class="form-group">
										<label>Index Name</label>
										<input type="text" bind:value={idx.name} placeholder="idx_name" />
									</div>
									<div class="form-group checkbox-group">
										<label class="inline-label">
											<input type="checkbox" bind:checked={idx.unique} />
											Unique
										</label>
									</div>
								</div>
								<div class="form-group">
									<label>Columns</label>
									{#if validColumns.length === 0}
										<span class="text-muted">Define columns first.</span>
									{:else}
										<div class="checkbox-row">
											{#each validColumns as vc}
												<label class="checkbox-label">
													<input
														type="checkbox"
														checked={idx.columns.includes(vc.name)}
														onchange={() => toggleIndexColumn(i, vc.name)}
													/>
													{vc.name}
												</label>
											{/each}
										</div>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				{:else if activeTab === 'foreignKeys'}
					<div class="cards-toolbar">
						<button class="btn-secondary btn-sm" onclick={addForeignKey}>+ Add Foreign Key</button>
					</div>
					{#if foreignKeys.length === 0}
						<div class="empty-state">No foreign keys defined.</div>
					{/if}
					{#each foreignKeys as fk, i (i)}
						<div class="card">
							<div class="card-header">
								<span class="card-title">Foreign Key #{i + 1}</span>
								<button class="btn-icon btn-tiny btn-danger-text" onclick={() => removeForeignKey(i)}>&times;</button>
							</div>
							<div class="card-body">
								<div class="form-group">
									<label>Constraint Name</label>
									<input type="text" bind:value={fk.name} placeholder="fk_name" />
								</div>
								<div class="form-group">
									<label>Columns</label>
									{#if validColumns.length === 0}
										<span class="text-muted">Define columns first.</span>
									{:else}
										<div class="checkbox-row">
											{#each validColumns as vc}
												<label class="checkbox-label">
													<input
														type="checkbox"
														checked={fk.columns.includes(vc.name)}
														onchange={() => toggleFkColumn(i, vc.name)}
													/>
													{vc.name}
												</label>
											{/each}
										</div>
									{/if}
								</div>
								<div class="form-row">
									<div class="form-group">
										<label>Reference Table</label>
										<input type="text" bind:value={fk.refTable} placeholder="other_table" />
									</div>
									<div class="form-group">
										<label>Reference Columns</label>
										<input type="text" bind:value={fk.refColumns} placeholder="col1, col2" />
									</div>
								</div>
								<div class="form-row">
									<div class="form-group">
										<label>ON DELETE</label>
										<select bind:value={fk.onDelete}>
											{#each FK_ACTIONS as action}<option value={action}>{action}</option>{/each}
										</select>
									</div>
									<div class="form-group">
										<label>ON UPDATE</label>
										<select bind:value={fk.onUpdate}>
											{#each FK_ACTIONS as action}<option value={action}>{action}</option>{/each}
										</select>
									</div>
								</div>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn-secondary" onclick={preview} disabled={previewing || !canSave}>
					{previewing ? 'Loading...' : 'Preview SQL'}
				</button>
				<button class="btn-primary" onclick={save} disabled={saving || !canSave}>
					{saving ? 'Saving...' : (mode === 'create' ? 'Create Table' : 'Save Changes')}
				</button>
				<button class="btn-secondary" onclick={onClose}>Cancel</button>
			</div>
		{/if}

		{#if previewSql}
			<div class="preview-overlay" onclick={() => (previewSql = null)}>
				<div class="preview-modal" onclick={(e) => e.stopPropagation()}>
					<div class="preview-header">
						<h3>Generated SQL</h3>
						<button class="btn-icon" onclick={() => (previewSql = null)}>&times;</button>
					</div>
					<pre class="preview-sql">{previewSql}</pre>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.modal.wide {
		min-width: 720px;
		width: 85vw;
	}

	.loading {
		padding: 40px;
		text-align: center;
		color: var(--text-muted);
	}

	.tabs {
		display: flex;
		gap: 2px;
		margin-bottom: 16px;
		border-bottom: 1px solid var(--border);
	}

	.tab {
		padding: 8px 16px;
		background: transparent;
		color: var(--text-secondary);
		border: none;
		border-bottom: 2px solid transparent;
		border-radius: 0;
		font-weight: 500;
	}

	.tab:hover {
		color: var(--text-primary);
		background: transparent;
	}

	.tab.active {
		color: var(--accent);
		border-bottom-color: var(--accent);
	}

	.error-banner {
		padding: 8px 12px;
		margin-bottom: 12px;
		border-radius: var(--radius);
		background: rgba(243, 139, 168, 0.15);
		color: var(--error);
		font-size: 13px;
	}

	.btn-sm {
		padding: 4px 10px;
		font-size: 12px;
	}

	.btn-tiny {
		width: 22px;
		height: 22px;
		font-size: 11px;
		padding: 0;
	}

	.btn-danger-text {
		background: transparent;
		color: var(--error);
	}

	.btn-danger-text:hover {
		background: rgba(243, 139, 168, 0.15);
	}

	.columns-toolbar,
	.cards-toolbar {
		margin-bottom: 8px;
	}

	.columns-table-wrap {
		overflow-x: auto;
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}

	.columns-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 12px;
	}

	.columns-table th,
	.columns-table td {
		padding: 4px 6px;
		border: 1px solid var(--border);
		text-align: left;
		white-space: nowrap;
	}

	.columns-table th {
		background: var(--bg-tertiary);
		color: var(--text-secondary);
		font-weight: 600;
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.03em;
	}

	.th-narrow {
		width: 36px;
		min-width: 36px;
	}

	.th-length {
		width: 64px;
		min-width: 64px;
	}

	.th-check {
		width: 50px;
		min-width: 50px;
		text-align: center;
	}

	.td-reorder {
		display: flex;
		gap: 2px;
		align-items: center;
		justify-content: center;
	}

	.td-center {
		text-align: center;
	}

	.empty-row {
		text-align: center;
		padding: 24px 12px !important;
		color: var(--text-muted);
	}

	.columns-table input[type="text"],
	.columns-table select {
		width: 100%;
		min-width: 60px;
		background: transparent;
		border: 1px solid transparent;
		padding: 2px 4px;
		font-size: 12px;
		border-radius: 3px;
	}

	.columns-table input[type="text"]:focus,
	.columns-table select:focus {
		background: var(--bg-secondary);
		border-color: var(--accent);
	}

	.columns-table input[type="checkbox"] {
		margin: 0;
	}

	.card {
		background: var(--bg-tertiary);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		margin-bottom: 8px;
	}

	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		border-bottom: 1px solid var(--border);
	}

	.card-title {
		font-weight: 600;
		font-size: 13px;
	}

	.card-body {
		padding: 12px;
	}

	.checkbox-group {
		display: flex;
		align-items: flex-end;
	}

	.inline-label {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		cursor: pointer;
		padding: 6px 0;
	}

	.checkbox-row {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		cursor: pointer;
	}

	.text-muted {
		color: var(--text-muted);
		font-size: 12px;
	}

	.empty-state {
		text-align: center;
		padding: 24px;
		color: var(--text-muted);
		font-size: 13px;
	}

	.preview-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1001;
	}

	.preview-modal {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: 12px;
		min-width: 560px;
		max-width: 80vw;
		max-height: 70vh;
		overflow-y: auto;
	}

	.preview-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid var(--border);
	}

	.preview-header h3 {
		font-size: 14px;
		font-weight: 600;
	}

	.preview-sql {
		padding: 16px;
		font-family: var(--font-mono);
		font-size: 13px;
		line-height: 1.5;
		color: var(--text-primary);
		white-space: pre-wrap;
		word-break: break-word;
		margin: 0;
	}
</style>
