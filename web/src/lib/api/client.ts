const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
	const resp = await fetch(`${API_BASE}${path}`, {
		headers: {
			'Content-Type': 'application/json',
			...options?.headers
		},
		...options
	});

	if (!resp.ok) {
		const err = await resp.json().catch(() => ({ error: resp.statusText }));
		throw new Error(err.error || `HTTP ${resp.status}`);
	}

	return resp.json();
}

export interface ConnectionInfo {
	id: string;
	name: string;
	driver: string;
	host: string;
	port: number;
	username: string;
	database: string;
	options: string;
	createdAt: string;
	updatedAt: string;
}

export interface ConnectionCreate {
	name: string;
	driver: string;
	host: string;
	port: number;
	username: string;
	password: string;
	database: string;
	options?: string;
}

export interface DatabaseInfo {
	name: string;
}

export interface SchemaInfo {
	name: string;
}

export interface TableInfo {
	name: string;
	schema: string;
	type: string;
	comment: string;
	rowCount: number;
}

export interface ColumnInfo {
	name: string;
	type: string;
	nullable: boolean;
	defaultValue: string;
	isPrimary: boolean;
	autoIncrement: boolean;
	comment: string;
}

export interface IndexInfo {
	name: string;
	columns: string[];
	unique: boolean;
	type: string;
}

export interface ForeignKeyInfo {
	name: string;
	columns: string[];
	refTable: string;
	refColumns: string[];
	onDelete: string;
	onUpdate: string;
}

export interface ResultSet {
	columns: { name: string; type: string }[];
	rows: Record<string, unknown>[];
	total: number;
	durationMs: number;
}

export interface QueryHistoryEntry {
	id: number;
	connectionId: string;
	database: string;
	sql: string;
	duration_ms: number;
	row_count: number;
	status: string;
	error_message: string;
	created_at: string;
}

export interface ExecResult {
	rowsAffected: number;
	lastInsertId?: number;
}

export const api = {
	connections: {
		list: () => request<ConnectionInfo[]>('/connections'),
		get: (id: string) => request<ConnectionInfo>(`/connections/${id}`),
		create: (data: ConnectionCreate) =>
			request<{ id: string }>('/connections', { method: 'POST', body: JSON.stringify(data) }),
		update: (id: string, data: Partial<ConnectionCreate>) =>
			request<{ message: string }>(`/connections/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
		delete: (id: string) =>
			request<{ message: string }>(`/connections/${id}`, { method: 'DELETE' }),
		test: (id: string) =>
			request<{ status: string }>(`/connections/${id}/test`, { method: 'POST' })
	},

	metadata: {
		databases: (connId: string) =>
			request<DatabaseInfo[]>(`/connections/${connId}/databases`),
		schemas: (connId: string, db?: string) =>
			request<SchemaInfo[]>(`/connections/${connId}/schemas${db ? `?database=${db}` : ''}`),
		tables: (connId: string, schema: string) =>
			request<TableInfo[]>(`/connections/${connId}/schemas/${schema}/tables`),
		columns: (connId: string, schema: string, table: string) =>
			request<ColumnInfo[]>(`/connections/${connId}/schemas/${schema}/tables/${table}/columns`),
		indexes: (connId: string, schema: string, table: string) =>
			request<IndexInfo[]>(`/connections/${connId}/schemas/${schema}/tables/${table}/indexes`),
		foreignKeys: (connId: string, schema: string, table: string) =>
			request<ForeignKeyInfo[]>(`/connections/${connId}/schemas/${schema}/tables/${table}/fks`)
	},

	query: {
		execute: (connId: string, sql: string, database?: string) =>
			request<ResultSet>(`/connections/${connId}/query`, {
				method: 'POST',
				body: JSON.stringify({ sql, database })
			}),
		explain: (connId: string, sql: string, database?: string) =>
			request<ResultSet>(`/connections/${connId}/explain`, {
				method: 'POST',
				body: JSON.stringify({ sql, database })
			})
	},

	history: {
		list: (connId: string, limit?: number) =>
			request<QueryHistoryEntry[]>(`/connections/${connId}/history${limit ? `?limit=${limit}` : ''}`),
		clear: (connId: string) =>
			request<{ message: string }>(`/connections/${connId}/history`, { method: 'DELETE' }),
		saveFavorite: (connId: string, title: string, sql: string, database?: string) =>
			request<{ message: string }>(`/connections/${connId}/favorites`, {
				method: 'POST',
				body: JSON.stringify({ title, sql, database })
			})
	}
};
