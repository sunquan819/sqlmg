<script lang="ts">
	import { t } from '$lib/i18n';

	let {
		mode,
		connectionId,
		schema,
		tableName,
		onClose
	}: {
		mode: 'export' | 'import';
		connectionId: string;
		schema: string;
		tableName?: string;
		onClose: () => void;
	} = $props();

	let format = $state<'csv' | 'json' | 'sql'>('csv');
	let delimiter = $state(',');
	let includeHeader = $state(true);
	let batchSize = $state(1000);

	let file = $state<File | null>(null);
	let importTable = $state(tableName ?? '');
	let hasHeader = $state(true);
	let onError = $state<'abort' | 'skip'>('abort');
	let importBatchSize = $state(1000);

	let loading = $state(false);
	let message = $state<{ type: 'success' | 'error'; text: string } | null>(null);

	let previewData = $state<{ columns: string[]; rows: string[][] } | null>(null);
	let importResult = $state<{
		totalRows: number;
		inserted: number;
		failed: number;
		errors: string[];
	} | null>(null);

	let isDragging = $state(false);
	let fileInput: HTMLInputElement;

	function onDragOver(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function onDragLeave() {
		isDragging = false;
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		const dropped = e.dataTransfer?.files[0];
		if (dropped) {
			file = dropped;
			autoDetectFormat(dropped.name);
		}
	}

	function onFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.files?.[0]) {
			file = input.files[0];
			autoDetectFormat(input.files[0].name);
		}
	}

	function autoDetectFormat(filename: string) {
		const ext = filename.split('.').pop()?.toLowerCase();
		if (ext === 'json') format = 'json';
		else if (ext === 'sql') format = 'sql';
		else format = 'csv';
	}

	function buildExportQueryParams(): string {
		const params = new URLSearchParams();
		params.set('format', format);
		if (format === 'csv') {
			params.set('delimiter', delimiter);
			params.set('header', String(includeHeader));
		}
		if (format === 'sql') {
			params.set('batchSize', String(batchSize));
		}
		return params.toString();
	}

	async function doExport() {
		loading = true;
		message = null;
		try {
			const qs = buildExportQueryParams();
			const url = tableName
				? `/api/connections/${connectionId}/schemas/${schema}/tables/${tableName}/export?${qs}`
				: `/api/connections/${connectionId}/schemas/${schema}/export?${qs}`;
			const link = document.createElement('a');
			link.href = url;
			link.download = `${tableName || schema}_export.${format}`;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
			message = { type: 'success', text: t('export.success') };
		} catch (e: any) {
			message = { type: 'error', text: e.message || t('export.failed') };
		} finally {
			loading = false;
		}
	}

	async function doPreview() {
		if (!file) return;
		loading = true;
		message = null;
		previewData = null;
		importResult = null;
		try {
			const fd = new FormData();
			fd.append('file', file);
			fd.append('format', format);
			fd.append('schema', schema);
			fd.append('table', importTable);
			if (format === 'csv') {
				fd.append('delimiter', delimiter);
				fd.append('header', String(hasHeader));
			}
			fd.append('batchSize', String(importBatchSize));
			fd.append('onError', onError);
			const resp = await fetch(`/api/connections/${connectionId}/import/preview`, {
				method: 'POST',
				body: fd
			});
			const data = await resp.json();
			if (data.error) {
				message = { type: 'error', text: data.error };
			} else {
				previewData = {
					columns: data.columns || [],
					rows: (data.rows || []).slice(0, 10)
				};
			}
		} catch (e: any) {
			message = { type: 'error', text: e.message || t('common.error') };
		} finally {
			loading = false;
		}
	}

	async function doImport() {
		if (!file) return;
		loading = true;
		message = null;
		importResult = null;
		try {
			const fd = new FormData();
			fd.append('file', file);
			fd.append('format', format);
			fd.append('schema', schema);
			fd.append('table', importTable);
			if (format === 'csv') {
				fd.append('delimiter', delimiter);
				fd.append('header', String(hasHeader));
			}
			fd.append('batchSize', String(importBatchSize));
			fd.append('onError', onError);
			const resp = await fetch(`/api/connections/${connectionId}/import`, {
				method: 'POST',
				body: fd
			});
			const data = await resp.json();
			if (data.error) {
				message = { type: 'error', text: data.error };
			} else {
				importResult = {
					totalRows: data.totalRows ?? 0,
					inserted: data.inserted ?? 0,
					failed: data.failed ?? 0,
					errors: data.errors ?? []
				};
				message = { type: 'success', text: t('import.success') };
			}
		} catch (e: any) {
			message = { type: 'error', text: e.message || t('import.failed') };
		} finally {
			loading = false;
		}
	}

	function clearFile() {
		file = null;
		previewData = null;
		importResult = null;
		if (fileInput) fileInput.value = '';
	}
</script>

<div class="overlay" onclick={onClose}>
	<div class="modal import-export-modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>{mode === 'export' ? t('export.title') : t('import.title')}</h2>
			<button class="btn-icon" onclick={onClose}>&times;</button>
		</div>

		<div class="modal-body">
			{#if message}
				<div class="message" class:message-success={message.type === 'success'} class:message-error={message.type === 'error'}>
					{message.text}
				</div>
			{/if}

			{#if mode === 'export'}
				<div class="form-group">
					<label>{t('export.format')}</label>
					<select bind:value={format}>
						<option value="csv">CSV</option>
						<option value="json">JSON</option>
						<option value="sql">SQL</option>
					</select>
				</div>

				{#if format === 'csv'}
					<div class="form-row">
						<div class="form-group">
							<label>{t('export.delimiter')}</label>
							<input type="text" bind:value={delimiter} maxlength="1" />
						</div>
						<div class="form-group checkbox-group">
							<label>
								<input type="checkbox" bind:checked={includeHeader} />
								{t('export.includeHeader')}
							</label>
						</div>
					</div>
				{/if}

				{#if format === 'sql'}
					<div class="form-group">
						<label>{t('import.batchSize')}</label>
						<input type="number" bind:value={batchSize} min="1" />
					</div>
				{/if}
			{:else}
				<div
					class="drop-zone"
					class:dragging={isDragging}
					class:has-file={!!file}
					ondragover={onDragOver}
					ondragleave={onDragLeave}
					ondrop={onDrop}
					onclick={() => fileInput?.click()}
					role="button"
					tabindex="0"
					onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') fileInput?.click(); }}
				>
					<input
						bind:this={fileInput}
						type="file"
						accept=".csv,.json,.sql"
						onchange={onFileSelect}
						style="display: none;"
					/>
					{#if file}
						<div class="file-info">
							<span class="file-name">{file.name}</span>
							<span class="file-size">({(file.size / 1024).toFixed(1)} KB)</span>
							<button class="btn-icon file-remove" onclick={(e) => { e.stopPropagation(); clearFile(); }}>&times;</button>
						</div>
					{:else}
						<div class="drop-prompt">{t('import.file')}</div>
					{/if}
				</div>

				<div class="form-row">
					<div class="form-group">
						<label>{t('import.format')}</label>
						<select bind:value={format}>
							<option value="csv">CSV</option>
							<option value="json">JSON</option>
							<option value="sql">SQL</option>
						</select>
					</div>
					<div class="form-group">
						<label>{t('designer.tableName')}</label>
						<input type="text" bind:value={importTable} />
					</div>
				</div>

				{#if format === 'csv'}
					<div class="form-row">
						<div class="form-group">
							<label>{t('import.delimiter')}</label>
							<input type="text" bind:value={delimiter} maxlength="1" />
						</div>
						<div class="form-group">
							<label>{t('import.batchSize')}</label>
							<input type="number" bind:value={importBatchSize} min="1" />
						</div>
					</div>
					<div class="form-row">
						<div class="form-group checkbox-group">
							<label>
								<input type="checkbox" bind:checked={hasHeader} />
								{t('import.hasHeader')}
							</label>
						</div>
						<div class="form-group">
							<label>{t('import.onError')}</label>
							<select bind:value={onError}>
								<option value="abort">{t('import.onErrorAbort')}</option>
								<option value="skip">{t('import.onErrorSkip')}</option>
							</select>
						</div>
					</div>
				{:else}
					<div class="form-row">
						<div class="form-group">
							<label>{t('import.batchSize')}</label>
							<input type="number" bind:value={importBatchSize} min="1" />
						</div>
						<div class="form-group">
							<label>{t('import.onError')}</label>
							<select bind:value={onError}>
								<option value="abort">{t('import.onErrorAbort')}</option>
								<option value="skip">{t('import.onErrorSkip')}</option>
							</select>
						</div>
					</div>
				{/if}

				{#if previewData}
					<div class="preview-section">
						<h4>{t('import.preview')}</h4>
						<div class="preview-table-wrap">
							<table class="preview-table">
								<thead>
									<tr>
										{#each previewData.columns as col}
											<th>{col}</th>
										{/each}
									</tr>
								</thead>
								<tbody>
									{#each previewData.rows as row}
										<tr>
											{#each row as cell}
												<td>{cell ?? ''}</td>
											{/each}
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				{/if}

				{#if importResult}
					<div class="import-result">
						<div class="result-stat">
							<span class="stat-label">{t('import.totalRows')}</span>
							<span class="stat-value">{importResult.totalRows}</span>
						</div>
						<div class="result-stat">
							<span class="stat-label">{t('import.inserted')}</span>
							<span class="stat-value stat-success">{importResult.inserted}</span>
						</div>
						<div class="result-stat">
							<span class="stat-label">{t('import.failedRows')}</span>
							<span class="stat-value stat-error">{importResult.failed}</span>
						</div>
						{#if importResult.errors.length > 0}
							<div class="error-list">
								{#each importResult.errors.slice(0, 5) as err}
									<div class="error-item">{err}</div>
								{/each}
								{#if importResult.errors.length > 5}
									<div class="error-item error-more">+{importResult.errors.length - 5}</div>
								{/if}
							</div>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<div class="modal-footer">
			{#if mode === 'export'}
				<button class="btn-primary" onclick={doExport} disabled={loading}>
					{loading ? t('export.exporting') : t('export.export')}
				</button>
			{:else}
				<button class="btn-secondary" onclick={doPreview} disabled={loading || !file}>
					{t('import.preview')}
				</button>
				<button class="btn-primary" onclick={doImport} disabled={loading || !file || !importTable}>
					{loading ? t('import.importing') : t('import.import')}
				</button>
			{/if}
			<button class="btn-secondary" onclick={onClose}>{t('common.cancel')}</button>
		</div>
	</div>
</div>

<style>
	.import-export-modal {
		min-width: 520px;
		max-width: 640px;
	}

	.message {
		padding: 8px 12px;
		border-radius: var(--radius);
		font-size: 13px;
		margin-bottom: 12px;
	}

	.message-success {
		background: rgba(166, 227, 161, 0.15);
		color: var(--success);
	}

	.message-error {
		background: rgba(243, 139, 168, 0.15);
		color: var(--error);
	}

	.drop-zone {
		border: 2px dashed var(--border);
		border-radius: var(--radius);
		padding: 32px 16px;
		text-align: center;
		cursor: pointer;
		transition: all var(--transition);
		margin-bottom: 14px;
		background: var(--bg-primary);
	}

	.drop-zone:hover,
	.drop-zone.dragging {
		border-color: var(--accent);
		background: rgba(137, 180, 250, 0.05);
	}

	.drop-zone.has-file {
		border-style: solid;
		border-color: var(--success);
		padding: 12px 16px;
	}

	.drop-prompt {
		color: var(--text-muted);
		font-size: 13px;
	}

	.file-info {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}

	.file-name {
		color: var(--text-primary);
		font-weight: 500;
	}

	.file-size {
		color: var(--text-muted);
		font-size: 12px;
	}

	.file-remove {
		color: var(--text-muted) !important;
		font-size: 16px;
	}

	.file-remove:hover {
		color: var(--error) !important;
	}

	.checkbox-group {
		display: flex;
		align-items: flex-end;
	}

	.checkbox-group label {
		display: flex;
		align-items: center;
		gap: 6px;
		cursor: pointer;
		padding-bottom: 7px;
		color: var(--text-primary);
		font-size: 13px;
	}

	.checkbox-group input[type="checkbox"] {
		width: auto;
		accent-color: var(--accent);
	}

	.preview-section {
		margin-top: 14px;
	}

	.preview-section h4 {
		font-size: 13px;
		font-weight: 600;
		margin-bottom: 8px;
		color: var(--text-secondary);
	}

	.preview-table-wrap {
		overflow: auto;
		max-height: 240px;
		border: 1px solid var(--border);
		border-radius: var(--radius);
	}

	.preview-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 12px;
		font-family: var(--font-mono);
	}

	.preview-table th {
		background: var(--bg-tertiary);
		padding: 6px 10px;
		text-align: left;
		color: var(--text-secondary);
		font-weight: 600;
		white-space: nowrap;
		position: sticky;
		top: 0;
		z-index: 1;
	}

	.preview-table td {
		padding: 5px 10px;
		border-top: 1px solid var(--border);
		color: var(--text-primary);
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.import-result {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.result-stat {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 6px 12px;
		background: var(--bg-primary);
		border-radius: var(--radius);
	}

	.stat-label {
		color: var(--text-secondary);
		font-size: 13px;
	}

	.stat-value {
		font-weight: 600;
		font-size: 13px;
		color: var(--text-primary);
	}

	.stat-success {
		color: var(--success);
	}

	.stat-error {
		color: var(--error);
	}

	.error-list {
		margin-top: 4px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.error-item {
		padding: 4px 12px;
		font-size: 12px;
		color: var(--error);
		background: var(--bg-primary);
		border-radius: var(--radius);
		font-family: var(--font-mono);
	}

	.error-more {
		color: var(--text-muted);
		font-style: italic;
	}
</style>
