<script lang="ts">
	import { connections } from '$lib/stores';
	import { api, type ConnectionCreate } from '$lib/api/client';

	let { connectionId = null, onClose } = $props<{
		connectionId: string | null;
		onClose: () => void;
	}>();

	let form = $state<ConnectionCreate>({
		name: '',
		driver: 'mysql',
		host: '127.0.0.1',
		port: 3306,
		username: 'root',
		password: '',
		database: '',
		options: '{}'
	});

	let testing = $state(false);
	let saving = $state(false);
	let testResult = $state<{ ok: boolean; message: string } | null>(null);

	const driverDefaults: Record<string, { port: number; host: string }> = {
		mysql: { port: 3306, host: '127.0.0.1' },
		postgres: { port: 5432, host: '127.0.0.1' },
		sqlite: { port: 0, host: '' }
	};

	function onDriverChange() {
		const defaults = driverDefaults[form.driver];
		if (defaults) {
			form.port = defaults.port;
			form.host = defaults.host;
		}
	}

	async function testConnection() {
		testing = true;
		testResult = null;
		try {
			const result = await api.connections.create(form);
			const testResp = await api.connections.test(result.id);
			testResult = { ok: true, message: 'Connection successful!' };
			await api.connections.delete(result.id);
		} catch (e: any) {
			testResult = { ok: false, message: e.message || 'Connection failed' };
		} finally {
			testing = false;
		}
	}

	async function save() {
		saving = true;
		try {
			if (connectionId) {
				await api.connections.update(connectionId, form);
			} else {
				await api.connections.create(form);
			}
			$connections = await api.connections.list();
			onClose();
		} catch (e: any) {
			testResult = { ok: false, message: e.message || 'Save failed' };
		} finally {
			saving = false;
		}
	}
</script>

<div class="overlay" onclick={onClose}>
	<div class="modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>{connectionId ? 'Edit Connection' : 'New Connection'}</h2>
			<button class="btn-icon" onclick={onClose}>×</button>
		</div>

		<div class="modal-body">
			<div class="form-group">
				<label>Connection Name</label>
				<input type="text" bind:value={form.name} placeholder="My Database" />
			</div>

			<div class="form-group">
				<label>Driver</label>
				<select bind:value={form.driver} onchange={onDriverChange}>
					<option value="mysql">MySQL</option>
					<option value="postgres">PostgreSQL</option>
					<option value="sqlite">SQLite</option>
				</select>
			</div>

			{#if form.driver !== 'sqlite'}
				<div class="form-row">
					<div class="form-group">
						<label>Host</label>
						<input type="text" bind:value={form.host} />
					</div>
					<div class="form-group">
						<label>Port</label>
						<input type="number" bind:value={form.port} />
					</div>
				</div>

				<div class="form-row">
					<div class="form-group">
						<label>Username</label>
						<input type="text" bind:value={form.username} />
					</div>
					<div class="form-group">
						<label>Password</label>
						<input type="password" bind:value={form.password} />
					</div>
				</div>
			{:else}
				<div class="form-group">
					<label>Database File Path</label>
					<input type="text" bind:value={form.database} placeholder="/path/to/database.db" />
				</div>
			{/if}

			{#if form.driver !== 'sqlite'}
				<div class="form-group">
					<label>Database</label>
					<input type="text" bind:value={form.database} placeholder="(optional)" />
				</div>
			{/if}

			{#if testResult}
				<div class="test-result" class:ok={testResult.ok} class:fail={!testResult.ok}>
					{testResult.message}
				</div>
			{/if}
		</div>

		<div class="modal-footer">
			<button class="btn-secondary" onclick={testConnection} disabled={testing}>
				{testing ? 'Testing...' : 'Test Connection'}
			</button>
			<button class="btn-primary" onclick={save} disabled={saving || !form.name}>
				{saving ? 'Saving...' : 'Save'}
			</button>
			<button class="btn-secondary" onclick={onClose}>Cancel</button>
		</div>
	</div>
</div>

<style>
	.test-result {
		padding: 8px 12px; border-radius: var(--radius);
		font-size: 13px; margin-top: 8px;
	}
	.test-result.ok { background: rgba(166,227,161,0.15); color: var(--success); }
	.test-result.fail { background: rgba(243,139,168,0.15); color: var(--error); }
</style>
