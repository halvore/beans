<script lang="ts">
	import { ui } from '$lib/uiState.svelte';

	let inputEl = $state<HTMLInputElement | null>(null);
	const isMac = navigator.platform.startsWith('Mac');
	const shortcutHint = isMac ? '⌘/' : 'Ctrl+/';

	export function focus() {
		inputEl?.focus();
	}
</script>

<div class="relative">
	<input
		bind:this={inputEl}
		type="text"
		placeholder="Filter beans… ({shortcutHint})"
		value={ui.filterText}
		oninput={(e) => ui.setFilterText(e.currentTarget.value)}
		class="w-full rounded border border-border bg-surface px-3 py-1.5 pr-8 text-sm text-text placeholder:text-text-faint focus:border-accent focus:outline-none"
		data-testid="filter-input"
	/>
	{#if ui.filterText}
		<button
			onclick={() => {
				ui.setFilterText('');
				inputEl?.focus();
			}}
			class="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text cursor-pointer"
			title="Clear filter"
			data-testid="filter-clear"
		>
			&#x2715;
		</button>
	{/if}
</div>
