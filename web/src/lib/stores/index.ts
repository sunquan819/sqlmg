import { writable } from 'svelte/store';
import type { ConnectionInfo } from '$lib/api/client';

export const connections = writable<ConnectionInfo[]>([]);
export const activeConnectionId = writable<string | null>(null);
export const activeDatabase = writable<string>('');
export const activeSchema = writable<string>('');

export interface Tab {
	id: string;
	title: string;
	type: 'query' | 'table' | 'designer';
	connectionId: string;
	database: string;
	schema: string;
	sql?: string;
	tableName?: string;
	resultSet?: unknown;
	executing?: boolean;
	error?: string;
	durationMs?: number;
}

export const tabs = writable<Tab[]>([]);
export const activeTabId = writable<string | null>(null);

export interface TreeNode {
	id: string;
	label: string;
	icon: string;
	type: 'connection' | 'database' | 'schema' | 'table' | 'view' | 'column' | 'index' | 'folder';
	children?: TreeNode[];
	loaded?: boolean;
	loading?: boolean;
	connectionId?: string;
	database?: string;
	schema?: string;
	tableName?: string;
}

export const explorerTree = writable<TreeNode[]>([]);
export const expandedNodes = writable<Set<string>>(new Set());

export const sidebarWidth = writable(280);
export const sidebarCollapsed = writable(false);

export interface Settings {
	theme: 'dark' | 'light' | 'system';
	fontSize: number;
	queryLimit: number;
	confirmDestructive: boolean;
}

export const settings = writable<Settings>({
	theme: 'dark',
	fontSize: 14,
	queryLimit: 1000,
	confirmDestructive: true
});
