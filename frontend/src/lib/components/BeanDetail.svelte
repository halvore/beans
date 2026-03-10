<script lang="ts">
	import type { Bean } from '$lib/beans.svelte';
	import { beansStore } from '$lib/beans.svelte';
	import { worktreeStore } from '$lib/worktrees.svelte';
	import { ui } from '$lib/uiState.svelte';
	import { renderMarkdown } from '$lib/markdown';
	import { statusColors, typeColors, priorityColors } from '$lib/styles';
	import { client } from '$lib/graphqlClient';
	import { gql } from 'urql';
	import BeanCard from './BeanCard.svelte';
	import ConfirmModal from './ConfirmModal.svelte';

	interface Props {
		bean: Bean;
		onSelect?: (bean: Bean) => void;
		onEdit?: (bean: Bean) => void;
	}

	let { bean, onSelect, onEdit }: Props = $props();

	const parent = $derived(bean.parentId ? beansStore.get(bean.parentId) : null);
	const children = $derived(beansStore.children(bean.id));
	const blocking = $derived(
		bean.blockingIds.map((id) => beansStore.get(id)).filter((b): b is Bean => b !== undefined)
	);
	const blockedBy = $derived(beansStore.blockedBy(bean.id));

	let renderedBody = $state('');

	$effect(() => {
		const body = bean.body;
		if (body) {
			renderMarkdown(body).then((html) => {
				renderedBody = html;
			});
		} else {
			renderedBody = '';
		}
	});

	let copied = $state(false);

	function copyId() {
		navigator.clipboard.writeText(bean.id);
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}

	const worktree = $derived(worktreeStore.worktrees.find((wt) => wt.beanId === bean.id));
	const canStartWork = $derived(!worktree);

	let startingWork = $state(false);
	let removingWorktree = $state(false);
	let confirmingDestroy = $state(false);

	let worktreeError = $state<string | null>(null);

	const isArchivable = $derived(bean.status === 'completed' || bean.status === 'scrapped');
	let archiving = $state(false);

	const ARCHIVE_BEAN = gql`
		mutation ArchiveBean($id: ID!) {
			archiveBean(id: $id)
		}
	`;

	async function archiveBean() {
		archiving = true;
		const result = await client.mutation(ARCHIVE_BEAN, { id: bean.id }).toPromise();
		if (result.error) {
			worktreeError = result.error.message;
		}
		archiving = false;
	}

	async function startWork() {
		startingWork = true;
		worktreeError = null;
		const ok = await worktreeStore.startWork(bean.id);
		if (!ok) {
			worktreeError = worktreeStore.error;
		}
		startingWork = false;
	}

	async function destroyWorktree() {
		confirmingDestroy = false;
		removingWorktree = true;
		worktreeError = null;
		const ok = await worktreeStore.stopWork(bean.id);
		if (!ok) {
			worktreeError = worktreeStore.error;
		}
		removingWorktree = false;
	}

	function handleBeanLinkClick(e: MouseEvent) {
		const target = (e.target as HTMLElement).closest<HTMLElement>('[data-bean-id]');
		if (!target) return;
		e.preventDefault();
		const linkedBean = beansStore.get(target.dataset.beanId!);
		if (linkedBean) ui.selectBean(linkedBean);
	}
</script>


<div class="h-full overflow-auto p-6">
	<!-- Header -->
	<div class="mb-6">
		<div class="flex items-center gap-2 mb-2 flex-wrap">
			<button
				onclick={copyId}
				class="px-2 py-1 text-xs font-mono rounded hover:bg-surface-alt transition-colors flex items-center gap-1"
				title="Copy ID to clipboard"
			>
				{bean.id}
				{#if copied}
					<span class="text-success">&#10003;</span>
				{:else}
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
						/>
					</svg>
				{/if}
			</button>
			<span class={["text-[11px] px-2 py-0.5 rounded-full font-medium", typeColors[bean.type] ?? "bg-type-task-bg text-type-task-text"]}>{bean.type}</span>
			<span class={["text-[11px] px-2 py-0.5 rounded-full font-medium", statusColors[bean.status] ?? "bg-status-todo-bg text-status-todo-text"]}>{bean.status}</span>
			{#if bean.priority && bean.priority !== 'normal'}
				<span class={["text-[11px] px-2 py-0.5 rounded-full font-medium border", priorityColors[bean.priority]]}>
					{bean.priority}
				</span>
			{/if}
		</div>
		<div class="flex items-center gap-2">
			<h1 class="text-2xl font-bold text-text flex-1">{bean.title}</h1>
			{#if canStartWork}
				<button
					class="px-3 py-1.5 text-sm font-medium rounded-md bg-success text-white hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center gap-2"
					onclick={startWork}
					disabled={startingWork}
				>
					{#if startingWork}
						<span class="inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
					{/if}
					Start Work
				</button>
			{/if}
			{#if isArchivable}
				<button
					class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md border border-border text-text-muted hover:bg-surface-alt transition-colors disabled:opacity-50"
					onclick={archiveBean}
					disabled={archiving}
					title="Archive this bean"
				>
					<span class="icon-[uil--archive] size-4"></span>
					{archiving ? 'Archiving…' : 'Archive'}
				</button>
			{/if}
			{#if onEdit}
				<button class="px-3 py-1.5 text-sm font-medium rounded-md border border-border text-text-muted hover:bg-surface-alt transition-colors" onclick={() => onEdit(bean)}>Edit</button>
			{/if}
		</div>
	</div>

	<!-- Worktree error -->
	{#if worktreeError}
		<div class="mb-6 rounded-lg border border-danger/30 bg-danger/5 p-3">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2 min-w-0">
					<span class="text-danger text-xs font-semibold uppercase shrink-0">Worktree Error</span>
					<span class="text-xs text-danger/80 truncate">{worktreeError}</span>
				</div>
				<button
					class="text-danger/60 hover:text-danger text-xs px-1 cursor-pointer"
					onclick={() => (worktreeError = null)}
				>
					✕
				</button>
			</div>
		</div>
	{/if}

	<!-- Worktree -->
	{#if worktree}
		<div class="mb-6 rounded-lg border border-success/30 bg-success/5 p-3">
			<div class="flex items-center justify-between mb-2">
				<h2 class="text-xs font-semibold text-success uppercase">Active Worktree</h2>
				<button
					class="px-2 py-1 text-xs font-medium rounded-md border border-danger/30 text-danger hover:bg-danger/10 transition-colors disabled:opacity-50"
					onclick={() => (confirmingDestroy = true)}
					disabled={removingWorktree}
				>
					{#if removingWorktree}
						Removing…
					{:else}
						Destroy Worktree
					{/if}
				</button>
			</div>
			<div class="text-xs text-text-muted space-y-1">
				<div class="flex gap-2">
					<span class="text-text-faint w-12 shrink-0">Branch</span>
					<code class="text-text truncate">{worktree.branch}</code>
				</div>
				<div class="flex gap-2">
					<span class="text-text-faint w-12 shrink-0">Path</span>
					<code class="text-text truncate">{worktree.path}</code>
				</div>
			</div>
		</div>
	{/if}

	<!-- Tags -->
	{#if bean.tags.length > 0}
		<div class="mb-6">
			<h2 class="text-xs font-semibold text-text-muted uppercase mb-2">Tags</h2>
			<div class="flex gap-1 flex-wrap">
				{#each bean.tags as tag}
					<span class="text-[11px] px-2 py-0.5 rounded-full border border-border text-text-muted">{tag}</span>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Relationships -->
	{#if parent || children.length > 0 || blocking.length > 0 || blockedBy.length > 0}
		<div class="mb-6 space-y-3">
			{#if parent}
				<div>
					<h2 class="text-xs font-semibold text-text-muted uppercase mb-1">Parent</h2>
					<BeanCard bean={parent} variant="compact" onclick={() => onSelect?.(parent)} />
				</div>
			{/if}

			{#if children.length > 0}
				<div>
					<h2 class="text-xs font-semibold text-text-muted uppercase mb-1">
						Children ({children.length})
					</h2>
					<div class="space-y-0.5">
						{#each children as child}
							<BeanCard bean={child} variant="compact" onclick={() => onSelect?.(child)} />
						{/each}
					</div>
				</div>
			{/if}

			{#if blocking.length > 0}
				<div>
					<h2 class="text-xs font-semibold text-text-muted uppercase mb-1">
						Blocking ({blocking.length})
					</h2>
					<div class="space-y-0.5">
						{#each blocking as b}
							<BeanCard bean={b} variant="compact" onclick={() => onSelect?.(b)} />
						{/each}
					</div>
				</div>
			{/if}

			{#if blockedBy.length > 0}
				<div>
					<h2 class="text-xs font-semibold text-text-muted uppercase mb-1">
						Blocked By ({blockedBy.length})
					</h2>
					<div class="space-y-0.5">
						{#each blockedBy as b}
							<BeanCard bean={b} variant="compact" onclick={() => onSelect?.(b)} />
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Body -->
	{#if bean.body}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="mb-6" onclick={handleBeanLinkClick}>
			<h2 class="text-xs font-semibold text-text-muted uppercase mb-2">Description</h2>
			<div class="bean-body prose max-w-none">
				{@html renderedBody}
			</div>
		</div>
	{/if}

	<!-- Metadata -->
	<div class="border-t border-border my-4"></div>
	<div class="text-xs text-text-faint space-y-1">
		<div>Created: {new Date(bean.createdAt).toLocaleString()}</div>
		<div>Updated: {new Date(bean.updatedAt).toLocaleString()}</div>
		<div>Path: {bean.path}</div>
	</div>
</div>

{#if confirmingDestroy}
	<ConfirmModal
		title="Destroy Worktree"
		message="This will delete the worktree branch and working directory. Any uncommitted changes will be lost."
		confirmLabel="Destroy"
		danger
		onConfirm={destroyWorktree}
		onCancel={() => (confirmingDestroy = false)}
	/>
{/if}

<style>
	.bean-body :global(h1) {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--th-md-h1);
		border-bottom: 1px solid var(--th-md-h1-border);
		padding-bottom: 0.25rem;
		margin-top: 1.5rem;
	}

	.bean-body :global(h2) {
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--th-md-h2);
		margin-top: 1.25rem;
	}

	.bean-body :global(h3) {
		font-size: 1rem;
		font-weight: 600;
		color: var(--th-md-h3);
		margin-top: 1rem;
	}

	.bean-body :global(h4),
	.bean-body :global(h5),
	.bean-body :global(h6) {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--th-md-h456);
		margin-top: 0.75rem;
	}

	.bean-body :global(ul:has(input[type='checkbox'])) {
		list-style: none;
		padding-left: 0;
	}

	.bean-body :global(li:has(> input[type='checkbox'])) {
		display: flex;
		align-items: flex-start;
		gap: 0.5rem;
		padding-left: 0;
	}

	.bean-body :global(li:has(> input[type='checkbox'])::before) {
		content: none;
	}

	.bean-body :global(input[type='checkbox']) {
		margin-top: 0.25rem;
		accent-color: #22c55e;
	}

	.bean-body :global(pre.shiki) {
		padding: 1rem;
		border-radius: 0.5rem;
		overflow-x: auto;
		font-size: 0.875rem;
		line-height: 1.5;
		margin: 1rem 0;
	}

	.bean-body :global(pre.shiki code) {
		font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, 'Cascadia Code', Consolas,
			'Liberation Mono', 'Courier New', monospace;
	}

	.bean-body :global(code:not(pre code)) {
		color: var(--th-text);
		background-color: var(--th-md-code-bg);
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		font-size: 0.875em;
		font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, 'Cascadia Code', Consolas,
			'Liberation Mono', 'Courier New', monospace;
	}
</style>
