<script lang="ts">
	import { onMount } from 'svelte';
	import { EditorView, keymap, placeholder as cmPlaceholder, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from '@codemirror/view';
	import { EditorState, Compartment } from '@codemirror/state';
	import { sql } from '@codemirror/lang-sql';
	import { oneDark } from '@codemirror/theme-one-dark';
	import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
	import { autocompletion, type CompletionContext, type CompletionResult } from '@codemirror/autocomplete';
	import { searchKeymap, highlightSelectionMatches } from '@codemirror/search';
	import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';

	let {
		value = $bindable(''),
		onExecute,
		onExecuteCurrent,
		schemaTables = [] as string[],
		schemaColumns = {} as Record<string, string[]>,
		placeholder = 'Enter SQL query (Ctrl+Enter to execute)...',
		readOnly = false
	}: {
		value?: string;
		onExecute?: (sql: string) => void;
		onExecuteCurrent?: (sql: string) => void;
		schemaTables?: string[];
		schemaColumns?: Record<string, string[]>;
		placeholder?: string;
		readOnly?: boolean;
	} = $props();

	let container: HTMLElement;
	let view: EditorView;
	let readOnlyCompartment = new Compartment();
	let themeCompartment = new Compartment();

	function sqlCompletion(context: CompletionContext): CompletionResult | null {
		const word = context.matchBefore(/[\w.]+/);
		if (!word || (word.from === word.to && !context.explicit)) return null;

		const text = word.text;
		const options: { label: string; type: string; detail?: string; apply?: string }[] = [];

		if (text.includes('.')) {
			const parts = text.split('.');
			if (parts.length === 2) {
				const tableName = parts[0];
				const colPrefix = parts[1].toLowerCase();
				const cols = schemaColumns[tableName] || [];
				for (const col of cols) {
					if (col.toLowerCase().startsWith(colPrefix)) {
						options.push({ label: `${tableName}.${col}`, type: 'property', detail: 'column' });
					}
				}
			}
		} else {
			const prefix = text.toLowerCase();
			for (const table of schemaTables) {
				if (table.toLowerCase().startsWith(prefix)) {
					options.push({ label: table, type: 'type', detail: 'table' });
				}
			}
			for (const table of Object.keys(schemaColumns)) {
				for (const col of schemaColumns[table]) {
					if (col.toLowerCase().startsWith(prefix)) {
						options.push({ label: col, type: 'property', detail: 'column' });
					}
				}
			}
		}

		if (options.length === 0) return null;
		return { from: word.from, options, validFor: /^[\w.]*$/ };
	}

	function getCurrentStatement(): string {
		if (!view) return value;
		const sel = view.state.selection.main;
		const selected = view.state.sliceDoc(sel.from, sel.to).trim();
		if (selected) return selected;

		const doc = view.state.doc.toString();
		const cursor = sel.head;
		const lines = doc.split('\n');
		let pos = 0;
		let startLine = 0;
		let endLine = lines.length - 1;

		for (let i = 0; i < lines.length; i++) {
			if (pos + lines[i].length >= cursor) {
				startLine = i;
				break;
			}
			pos += lines[i].length + 1;
		}

		for (let i = startLine; i >= 0; i--) {
			if (lines[i].trim() === '' || lines[i].trim().endsWith(';')) {
				startLine = i + 1;
				break;
			}
			if (i === 0) startLine = 0;
		}

		for (let i = startLine; i < lines.length; i++) {
			if (lines[i].trim().endsWith(';')) {
				endLine = i;
				break;
			}
			if (i === lines.length - 1) endLine = i;
		}

		return lines.slice(startLine, endLine + 1).join('\n').replace(/;$/, '').trim();
	}

	function handleExecute() {
		if (onExecute) onExecute(value);
	}

	function handleExecuteCurrent() {
		const stmt = getCurrentStatement();
		if (onExecuteCurrent) onExecuteCurrent(stmt);
		else if (onExecute) onExecute(stmt);
	}

	function handleFormat() {
		if (!view) return;
		const sql_text = view.state.doc.toString();
		const formatted = formatSQL(sql_text);
		view.dispatch({
			changes: { from: 0, to: view.state.doc.length, insert: formatted }
		});
		value = formatted;
	}

	function formatSQL(sql: string): string {
		const keywords = ['SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN',
			'INNER JOIN', 'OUTER JOIN', 'ON', 'GROUP BY', 'ORDER BY', 'HAVING', 'LIMIT', 'OFFSET',
			'INSERT INTO', 'UPDATE', 'DELETE FROM', 'SET', 'VALUES', 'CREATE TABLE', 'ALTER TABLE',
			'DROP TABLE', 'CREATE INDEX', 'UNION', 'UNION ALL', 'EXCEPT', 'INTERSECT',
			'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'AS', 'IN', 'NOT', 'IS', 'NULL',
			'BETWEEN', 'LIKE', 'EXISTS', 'DISTINCT', 'WITH'];

		let formatted = sql.replace(/\s+/g, ' ').trim();

		for (const kw of keywords) {
			const regex = new RegExp(`\\b${kw.replace(/ /g, '\\s+')}\\b`, 'gi');
			formatted = formatted.replace(regex, '\n' + kw);
		}

		return formatted.replace(/^\n+/, '').replace(/\n{3,}/g, '\n\n');
	}

	onMount(() => {
		const state = EditorState.create({
			doc: value,
			extensions: [
				lineNumbers(),
				highlightActiveLineGutter(),
				highlightActiveLine(),
				history(),
				highlightSelectionMatches(),
				closeBrackets(),
				sql(),
				oneDark,
				autocompletion({
					override: [sqlCompletion],
					activateOnTyping: true,
					maxRenderedOptions: 50
				}),
				cmPlaceholder(placeholder),
				readOnlyCompartment.of(EditorState.readOnly.of(readOnly)),
				keymap.of([
					...closeBracketsKeymap,
					...defaultKeymap,
					...searchKeymap,
					...historyKeymap,
					indentWithTab,
					{ key: 'Ctrl-Enter', run: () => { handleExecuteCurrent(); return true; } },
					{ key: 'Cmd-Enter', run: () => { handleExecuteCurrent(); return true; } },
					{ key: 'Ctrl-Shift-Enter', run: () => { handleExecute(); return true; } },
					{ key: 'Cmd-Shift-Enter', run: () => { handleExecute(); return true; } },
				]),
				EditorView.updateListener.of((update) => {
					if (update.docChanged) {
						value = update.state.doc.toString();
					}
				}),
				EditorView.theme({
					'&': { height: '100%', fontSize: '14px' },
					'.cm-content': { fontFamily: 'var(--font-mono)', padding: '8px 0' },
					'.cm-gutters': { background: 'var(--bg-secondary)', border: 'none', color: 'var(--text-muted)' },
					'.cm-activeLineGutter': { background: 'var(--bg-hover)' },
					'.cm-activeLine': { background: 'rgba(137, 180, 250, 0.06)' },
					'.cm-cursor': { borderLeftColor: 'var(--accent)' },
					'.cm-selectionBackground': { background: 'rgba(137, 180, 250, 0.2) !important' },
					'.cm-tooltip': { background: 'var(--bg-secondary)', border: '1px solid var(--border)', borderRadius: '6px' },
					'.cm-tooltip-autocomplete': { '& > ul > li': { padding: '4px 8px' } },
					'.cm-completion-label': { fontFamily: 'var(--font-mono)', fontSize: '13px' },
				})
			]
		});

		view = new EditorView({ state, parent: container });

		return () => {
			view.destroy();
		};
	});
</script>

<div class="editor-wrapper">
	<div class="editor-toolbar">
		<button class="btn-primary btn-sm" onclick={handleExecuteCurrent} title="Execute current statement (Ctrl+Enter)">
			▶ Run
		</button>
		<button class="btn-secondary btn-sm" onclick={handleExecute} title="Execute all (Ctrl+Shift+Enter)">
			▶▶ Run All
		</button>
		<button class="btn-secondary btn-sm" onclick={handleFormat} title="Format SQL">
			⟳ Format
		</button>
	</div>
	<div class="editor-container" bind:this={container}></div>
</div>

<style>
	.editor-wrapper {
		display: flex; flex-direction: column; height: 100%;
	}
	.editor-toolbar {
		display: flex; gap: 6px; padding: 6px 12px;
		border-bottom: 1px solid var(--border); flex-shrink: 0;
		background: var(--bg-secondary);
	}
	.btn-sm { font-size: 12px; padding: 3px 10px; }
	.editor-container {
		flex: 1; overflow: hidden;
	}
	.editor-container :global(.cm-editor) {
		height: 100%;
	}
</style>
