<script lang="ts">
  import type { Bean } from '$lib/beans.svelte';
  import { beansStore } from '$lib/beans.svelte';
  import { worktreeStore } from '$lib/worktrees.svelte';
  import { statusColors, typeColors, priorityColors } from '$lib/styles';
  import { client } from '$lib/graphqlClient';
  import { gql } from 'urql';
  import BeanCard from './BeanCard.svelte';
  import ConfirmModal from './ConfirmModal.svelte';
  import RenderedMarkdown from './RenderedMarkdown.svelte';

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

</script>

<div class="h-full overflow-auto p-6">
  <!-- Header -->
  <div class="mb-6">
    <div class="mb-2 flex flex-wrap items-center gap-2">
      <button
        onclick={copyId}
        class="flex items-center gap-1 rounded px-2 py-1 font-mono text-xs transition-colors hover:bg-surface-alt"
        title="Copy ID to clipboard"
      >
        {bean.id}
        {#if copied}
          <span class="text-success">&#10003;</span>
        {:else}
          <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
            />
          </svg>
        {/if}
      </button>
      <span
        class={[
          'rounded-full px-2 py-0.5 text-[11px] font-medium',
          typeColors[bean.type] ?? 'bg-type-task-bg text-type-task-text'
        ]}>{bean.type}</span
      >
      <span
        class={[
          'rounded-full px-2 py-0.5 text-[11px] font-medium',
          statusColors[bean.status] ?? 'bg-status-todo-bg text-status-todo-text'
        ]}>{bean.status}</span
      >
      {#if bean.priority && bean.priority !== 'normal'}
        <span
          class={[
            'rounded-full border px-2 py-0.5 text-[11px] font-medium',
            priorityColors[bean.priority]
          ]}
        >
          {bean.priority}
        </span>
      {/if}
    </div>
    <div class="flex items-center gap-2">
      <h1 class="flex-1 text-2xl font-bold text-text">{bean.title}</h1>
      {#if canStartWork}
        <button
          class="flex items-center gap-2 rounded-md bg-success px-3 py-1.5 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-50"
          onclick={startWork}
          disabled={startingWork}
        >
          {#if startingWork}
            <span
              class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
            ></span>
          {/if}
          Start Work
        </button>
      {/if}
      {#if isArchivable}
        <button
          class="flex items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-text-muted transition-colors hover:bg-surface-alt disabled:opacity-50"
          onclick={archiveBean}
          disabled={archiving}
          title="Archive this bean"
        >
          <span class="icon-[uil--archive] size-4"></span>
          {archiving ? 'Archiving…' : 'Archive'}
        </button>
      {/if}
      {#if onEdit}
        <button
          class="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-text-muted transition-colors hover:bg-surface-alt"
          onclick={() => onEdit(bean)}>Edit</button
        >
      {/if}
    </div>
  </div>

  <!-- Worktree error -->
  {#if worktreeError}
    <div class="mb-6 rounded-lg border border-danger/30 bg-danger/5 p-3">
      <div class="flex items-center justify-between">
        <div class="flex min-w-0 items-center gap-2">
          <span class="shrink-0 text-xs font-semibold text-danger uppercase">Worktree Error</span>
          <span class="truncate text-xs text-danger/80">{worktreeError}</span>
        </div>
        <button
          class="cursor-pointer px-1 text-xs text-danger/60 hover:text-danger"
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
      <div class="mb-2 flex items-center justify-between">
        <h2 class="text-xs font-semibold text-success uppercase">Active Worktree</h2>
        <button
          class="rounded-md border border-danger/30 px-2 py-1 text-xs font-medium text-danger transition-colors hover:bg-danger/10 disabled:opacity-50"
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
      <div class="space-y-1 text-xs text-text-muted">
        <div class="flex gap-2">
          <span class="w-12 shrink-0 text-text-faint">Branch</span>
          <code class="truncate text-text">{worktree.branch}</code>
        </div>
        <div class="flex gap-2">
          <span class="w-12 shrink-0 text-text-faint">Path</span>
          <code class="truncate text-text">{worktree.path}</code>
        </div>
      </div>
    </div>
  {/if}

  <!-- Tags -->
  {#if bean.tags.length > 0}
    <div class="mb-6">
      <h2 class="mb-2 text-xs font-semibold text-text-muted uppercase">Tags</h2>
      <div class="flex flex-wrap gap-1">
        {#each bean.tags as tag}
          <span class="rounded-full border border-border px-2 py-0.5 text-[11px] text-text-muted"
            >{tag}</span
          >
        {/each}
      </div>
    </div>
  {/if}

  <!-- Relationships -->
  {#if parent || children.length > 0 || blocking.length > 0 || blockedBy.length > 0}
    <div class="mb-6 space-y-3">
      {#if parent}
        <div>
          <h2 class="mb-1 text-xs font-semibold text-text-muted uppercase">Parent</h2>
          <BeanCard bean={parent} variant="compact" onclick={() => onSelect?.(parent)} />
        </div>
      {/if}

      {#if children.length > 0}
        <div>
          <h2 class="mb-1 text-xs font-semibold text-text-muted uppercase">
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
          <h2 class="mb-1 text-xs font-semibold text-text-muted uppercase">
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
          <h2 class="mb-1 text-xs font-semibold text-text-muted uppercase">
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
    <div class="mb-6">
      <h2 class="mb-2 text-xs font-semibold text-text-muted uppercase">Description</h2>
      <RenderedMarkdown content={bean.body} class="bean-body prose max-w-none" />
    </div>
  {/if}

  <!-- Metadata -->
  <div class="my-4 border-t border-border"></div>
  <div class="space-y-1 text-xs text-text-faint">
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
    font-family:
      ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, 'Cascadia Code', Consolas,
      'Liberation Mono', 'Courier New', monospace;
  }

  .bean-body :global(code:not(pre code)) {
    color: var(--th-text);
    background-color: var(--th-md-code-bg);
    padding: 0.125rem 0.375rem;
    border-radius: 0.25rem;
    font-size: 0.875em;
    font-family:
      ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, 'Cascadia Code', Consolas,
      'Liberation Mono', 'Courier New', monospace;
  }
</style>
