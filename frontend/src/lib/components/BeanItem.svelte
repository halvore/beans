<script lang="ts">
	import type { Bean } from '$lib/beans.svelte';
	import { beansStore } from '$lib/beans.svelte';
	import { worktreeStore } from '$lib/worktrees.svelte';
	import BeanItem from './BeanItem.svelte';

	interface Props {
		bean: Bean;
		depth?: number;
		selectedId?: string | null;
		onSelect?: (bean: Bean) => void;
	}

	let { bean, depth = 0, selectedId = null, onSelect }: Props = $props();

	const children = $derived(beansStore.children(bean.id));
	const isSelected = $derived(selectedId === bean.id);
	const hasWorktree = $derived(worktreeStore.hasWorktree(bean.id));

	const statusColors: Record<string, string> = {
		draft: 'bg-status-draft-bg text-status-draft-text',
		todo: 'bg-status-todo-bg text-status-todo-text',
		'in-progress': 'bg-status-in-progress-bg text-status-in-progress-text',
		completed: 'bg-status-completed-bg text-status-completed-text',
		scrapped: 'bg-status-scrapped-bg text-status-scrapped-text'
	};

	const typeBorders: Record<string, string> = {
		milestone: 'border-l-type-milestone-border',
		epic: 'border-l-type-epic-border',
		feature: 'border-l-type-feature-border',
		bug: 'border-l-type-bug-border',
		task: 'border-l-type-task-border'
	};

	function handleClick(e: MouseEvent) {
		e.stopPropagation();
		onSelect?.(bean);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onSelect?.(bean);
		}
	}
</script>

<div class="bean-item">
	<button
		onclick={handleClick}
		onkeydown={handleKeydown}
		class={[
			"relative overflow-hidden w-full rounded-lg border-l-3 p-2 text-left transition-all",
			hasWorktree ? "border-l-success" : (typeBorders[bean.type] ?? "border-l-type-task-border"),
			isSelected ? "bg-accent/10 ring-1 ring-accent" : "bg-surface hover:bg-surface-alt"
		]}
	>
		{#if hasWorktree}
			<div class="absolute top-0 right-0 size-4 bg-success" style="clip-path: polygon(0 0, 100% 0, 100% 100%)"></div>
		{/if}
		<div class="flex min-w-0 items-center gap-2">
			<code class="shrink-0 text-[10px] text-text-faint">{bean.id.slice(-4)}</code>
			<span class="flex-1 truncate text-sm text-text">{bean.title}</span>
			<span
				class={[
					"shrink-0 rounded-full px-1.5 py-0.5 text-[10px] font-medium",
					statusColors[bean.status] ?? "bg-status-todo-bg text-status-todo-text"
				]}
			>
				{bean.status}
			</span>
			{#if children.length > 0}
				<span class="shrink-0 text-[10px] text-text-faint">+{children.length}</span>
			{/if}
		</div>
	</button>

	{#if children.length > 0}
		<div class="mt-1 ml-4 space-y-1 border-l border-border pl-2">
			{#each children as child (child.id)}
				<BeanItem bean={child} depth={depth + 1} {selectedId} {onSelect} />
			{/each}
		</div>
	{/if}
</div>
