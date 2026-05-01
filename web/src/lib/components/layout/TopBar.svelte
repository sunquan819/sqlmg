<script lang="ts">
	import { settings } from '$lib/stores';
	import { t, switchLang, getLang } from '$lib/i18n';

	let { onNewConnection } = $props();
	let currentLang = $state(getLang());

	function handleSwitchLang() {
		currentLang = switchLang();
	}
</script>

<header class="topbar">
	<div class="topbar-left">
		<div class="app-logo">
			<span class="logo-icon">⬡</span>
			<span class="logo-text">SQLMG</span>
		</div>
	</div>
	<div class="topbar-center">
		<button class="topbar-btn" onclick={onNewConnection}>
			<span class="icon">⊕</span> {t('explorer.newConnection')}
		</button>
	</div>
	<div class="topbar-right">
		<button class="btn-icon" title="语言/Language" onclick={handleSwitchLang}>
			{currentLang === 'zh' ? '中' : 'EN'}
		</button>
		<button class="btn-icon" title={t('common.refresh')}>⚙</button>
	</div>
</header>

<style>
	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 40px;
		padding: 0 12px;
		background: var(--bg-secondary);
		border-bottom: 1px solid var(--border);
		user-select: none;
		flex-shrink: 0;
	}

	.topbar-left {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.app-logo {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.logo-icon {
		font-size: 18px;
		color: var(--accent);
	}

	.logo-text {
		font-size: 14px;
		font-weight: 700;
		color: var(--text-primary);
		letter-spacing: 0.5px;
	}

	.topbar-center {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.topbar-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		background: var(--bg-tertiary);
		color: var(--text-secondary);
		font-size: 12px;
		padding: 4px 10px;
		border-radius: var(--radius);
	}

	.topbar-btn:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.topbar-right {
		display: flex;
		align-items: center;
		gap: 4px;
	}
</style>
