export type Lang = 'zh' | 'en';

const translations: Record<Lang, Record<string, string>> = {
	zh: {
		'app.title': 'SQLMG 数据库管理系统',
		'explorer.title': '对象浏览器',
		'explorer.search': '搜索对象...',
		'explorer.empty': '暂无连接',
		'explorer.newConnection': '新建连接',

		'connection.new': '新建连接',
		'connection.edit': '编辑连接',
		'connection.name': '连接名称',
		'connection.driver': '驱动',
		'connection.host': '主机',
		'connection.port': '端口',
		'connection.username': '用户名',
		'connection.password': '密码',
		'connection.database': '数据库',
		'connection.databasePath': '数据库文件路径',
		'connection.test': '测试连接',
		'connection.testing': '测试中...',
		'connection.success': '连接成功！',
		'connection.failed': '连接失败',
		'connection.save': '保存',
		'connection.saving': '保存中...',
		'connection.cancel': '取消',
		'connection.delete': '删除连接',
		'connection.deleteConfirm': '确定删除此连接？',

		'query.new': '新建查询',
		'query.run': '运行',
		'query.runAll': '全部运行',
		'query.format': '格式化',
		'query.placeholder': '输入 SQL 查询 (Ctrl+Enter 执行)...',
		'query.executing': '正在执行查询...',
		'query.noConnection': '未选择连接，请先连接到数据库。',
		'query.result.rows': '行',
		'query.result.noResult': '执行查询查看结果',
		'query.history.title': '查询历史',
		'query.history.empty': '暂无查询历史',
		'query.history.search': '搜索历史...',
		'query.history.clear': '清空历史',

		'designer.create': '创建表',
		'designer.alter': '设计表',
		'designer.columns': '列',
		'designer.indexes': '索引',
		'designer.foreignKeys': '外键',
		'designer.tableName': '表名',
		'designer.comment': '注释',
		'designer.addColumn': '添加列',
		'designer.addIndex': '添加索引',
		'designer.addFK': '添加外键',
		'designer.previewSQL': '预览SQL',
		'designer.createTable': '创建表',
		'designer.saveChanges': '保存修改',
		'designer.saving': '保存中...',
		'designer.generatedSQL': '生成的SQL',
		'designer.execute': '执行',
		'designer.type': '类型',
		'designer.length': '长度',
		'designer.nullable': '可空',
		'designer.default': '默认值',
		'designer.primaryKey': '主键',
		'designer.autoIncrement': '自增',
		'designer.unique': '唯一',
		'designer.references': '引用',
		'designer.onDelete': '删除时',
		'designer.onUpdate': '更新时',
		'designer.noIndexes': '未定义索引',
		'designer.noFKs': '未定义外键',
		'designer.loading': '加载表结构...',
		'designer.nameRequired': '表名不能为空',
		'designer.colNameRequired': '所有列必须有名称',

		'ddl.title': 'DDL',
		'ddl.copy': '复制',
		'ddl.copied': '已复制！',
		'ddl.loading': '加载DDL...',
		'ddl.close': '关闭',

		'er.title': 'ER 图',
		'er.exportImage': '导出图片',
		'er.noTables': '没有表',
		'er.reset': '重置',

		'context.designTable': '设计表',
		'context.viewDDL': '查看DDL',
		'context.newTable': '新建表',
		'context.editConnection': '编辑连接',
		'context.deleteConnection': '删除连接',
		'context.dropTable': '删除表',
		'context.dropTableConfirm': '确定删除表 "{name}"？此操作不可撤销。',
		'context.refresh': '刷新',
		'context.truncateTable': '清空表',
		'context.truncateConfirm': '确定清空表 "{name}" 的所有数据？此操作不可撤销。',

		'export.title': '导出数据',
		'export.format': '格式',
		'export.delimiter': '分隔符',
		'export.includeHeader': '包含表头',
		'export.export': '导出',
		'export.exporting': '导出中...',
		'export.success': '导出成功',
		'export.failed': '导出失败',

		'import.title': '导入数据',
		'import.format': '格式',
		'import.file': '选择文件',
		'import.delimiter': '分隔符',
		'import.hasHeader': '包含表头',
		'import.batchSize': '批量大小',
		'import.preview': '预览',
		'import.import': '导入',
		'import.importing': '导入中...',
		'import.success': '导入成功',
		'import.failed': '导入失败',
		'import.totalRows': '总行数',
		'import.inserted': '已插入',
		'import.failedRows': '失败行数',
		'import.columnMap': '列映射',
		'import.onError': '出错时',
		'import.onErrorAbort': '中止',
		'import.onErrorSkip': '跳过',

		'status.connected': '已连接',
		'status.disconnected': '未连接',
		'status.version': '版本',

		'welcome.title': 'SQLMG',
		'welcome.subtitle': '数据库管理系统',
		'welcome.newQuery': '新建查询',
		'welcome.execute': '执行当前语句',
		'welcome.executeAll': '执行全部',
		'welcome.save': '保存收藏',

		'common.confirm': '确认',
		'common.cancel': '取消',
		'common.close': '关闭',
		'common.save': '保存',
		'common.delete': '删除',
		'common.refresh': '刷新',
		'common.search': '搜索',
		'common.loading': '加载中...',
		'common.error': '错误',
		'common.success': '成功',
		'common.rows': '行',
		'common.columns': '列',
	},

	en: {
		'app.title': 'SQLMG Database Manager',
		'explorer.title': 'Explorer',
		'explorer.search': 'Search objects...',
		'explorer.empty': 'No connections yet',
		'explorer.newConnection': 'New Connection',

		'connection.new': 'New Connection',
		'connection.edit': 'Edit Connection',
		'connection.name': 'Connection Name',
		'connection.driver': 'Driver',
		'connection.host': 'Host',
		'connection.port': 'Port',
		'connection.username': 'Username',
		'connection.password': 'Password',
		'connection.database': 'Database',
		'connection.databasePath': 'Database File Path',
		'connection.test': 'Test Connection',
		'connection.testing': 'Testing...',
		'connection.success': 'Connection successful!',
		'connection.failed': 'Connection failed',
		'connection.save': 'Save',
		'connection.saving': 'Saving...',
		'connection.cancel': 'Cancel',
		'connection.delete': 'Delete Connection',
		'connection.deleteConfirm': 'Delete this connection?',

		'query.new': 'New Query',
		'query.run': 'Run',
		'query.runAll': 'Run All',
		'query.format': 'Format',
		'query.placeholder': 'Enter SQL query (Ctrl+Enter to execute)...',
		'query.executing': 'Executing query...',
		'query.noConnection': 'No connection selected. Please connect to a database first.',
		'query.result.rows': 'rows',
		'query.result.noResult': 'Execute a query to see results',
		'query.history.title': 'Query History',
		'query.history.empty': 'No query history',
		'query.history.search': 'Search history...',
		'query.history.clear': 'Clear History',

		'designer.create': 'Create Table',
		'designer.alter': 'Design Table',
		'designer.columns': 'Columns',
		'designer.indexes': 'Indexes',
		'designer.foreignKeys': 'Foreign Keys',
		'designer.tableName': 'Table Name',
		'designer.comment': 'Comment',
		'designer.addColumn': 'Add Column',
		'designer.addIndex': 'Add Index',
		'designer.addFK': 'Add Foreign Key',
		'designer.previewSQL': 'Preview SQL',
		'designer.createTable': 'Create Table',
		'designer.saveChanges': 'Save Changes',
		'designer.saving': 'Saving...',
		'designer.generatedSQL': 'Generated SQL',
		'designer.execute': 'Execute',
		'designer.type': 'Type',
		'designer.length': 'Length',
		'designer.nullable': 'Nullable',
		'designer.default': 'Default',
		'designer.primaryKey': 'PK',
		'designer.autoIncrement': 'AI',
		'designer.unique': 'Unique',
		'designer.references': 'References',
		'designer.onDelete': 'ON DELETE',
		'designer.onUpdate': 'ON UPDATE',
		'designer.noIndexes': 'No indexes defined',
		'designer.noFKs': 'No foreign keys defined',
		'designer.loading': 'Loading table structure...',
		'designer.nameRequired': 'Table name is required',
		'designer.colNameRequired': 'All columns must have a name',

		'ddl.title': 'DDL',
		'ddl.copy': 'Copy',
		'ddl.copied': 'Copied!',
		'ddl.loading': 'Loading DDL...',
		'ddl.close': 'Close',

		'er.title': 'ER Diagram',
		'er.exportImage': 'Export Image',
		'er.noTables': 'No tables found',
		'er.reset': 'Reset',

		'context.designTable': 'Design Table',
		'context.viewDDL': 'View DDL',
		'context.newTable': 'New Table',
		'context.editConnection': 'Edit Connection',
		'context.deleteConnection': 'Delete Connection',
		'context.dropTable': 'Drop Table',
		'context.dropTableConfirm': 'Drop table "{name}"? This cannot be undone.',
		'context.refresh': 'Refresh',
		'context.truncateTable': 'Truncate Table',
		'context.truncateConfirm': 'Truncate all data in "{name}"? This cannot be undone.',

		'export.title': 'Export Data',
		'export.format': 'Format',
		'export.delimiter': 'Delimiter',
		'export.includeHeader': 'Include Header',
		'export.export': 'Export',
		'export.exporting': 'Exporting...',
		'export.success': 'Export successful',
		'export.failed': 'Export failed',

		'import.title': 'Import Data',
		'import.format': 'Format',
		'import.file': 'Choose File',
		'import.delimiter': 'Delimiter',
		'import.hasHeader': 'Has Header',
		'import.batchSize': 'Batch Size',
		'import.preview': 'Preview',
		'import.import': 'Import',
		'import.importing': 'Importing...',
		'import.success': 'Import successful',
		'import.failed': 'Import failed',
		'import.totalRows': 'Total Rows',
		'import.inserted': 'Inserted',
		'import.failedRows': 'Failed Rows',
		'import.columnMap': 'Column Mapping',
		'import.onError': 'On Error',
		'import.onErrorAbort': 'Abort',
		'import.onErrorSkip': 'Skip',

		'status.connected': 'Connected',
		'status.disconnected': 'Disconnected',
		'status.version': 'Version',

		'welcome.title': 'SQLMG',
		'welcome.subtitle': 'Database Management System',
		'welcome.newQuery': 'New Query',
		'welcome.execute': 'Execute current statement',
		'welcome.executeAll': 'Execute all',
		'welcome.save': 'Save to favorites',

		'common.confirm': 'Confirm',
		'common.cancel': 'Cancel',
		'common.close': 'Close',
		'common.save': 'Save',
		'common.delete': 'Delete',
		'common.refresh': 'Refresh',
		'common.search': 'Search',
		'common.loading': 'Loading...',
		'common.error': 'Error',
		'common.success': 'Success',
		'common.rows': 'rows',
		'common.columns': 'columns',
	}
};

let currentLang: Lang = 'zh';

export function setLang(lang: Lang) {
	currentLang = lang;
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem('sqlmg:lang', lang);
	}
}

export function getLang(): Lang {
	if (typeof localStorage !== 'undefined') {
		const saved = localStorage.getItem('sqlmg:lang') as Lang;
		if (saved && (saved === 'zh' || saved === 'en')) {
			currentLang = saved;
		}
	}
	return currentLang;
}

export function t(key: string, params?: Record<string, string>): string {
	const dict = translations[currentLang] || translations.zh;
	let text = dict[key] || translations.zh[key] || key;

	if (params) {
		for (const [k, v] of Object.entries(params)) {
			text = text.replace(`{${k}}`, v);
		}
	}

	return text;
}

export function switchLang(): Lang {
	const newLang = currentLang === 'zh' ? 'en' : 'zh';
	setLang(newLang);
	return newLang;
}

getLang();
