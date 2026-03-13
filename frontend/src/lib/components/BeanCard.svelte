<script lang="ts">
  import type { Bean } from '$lib/beans.svelte';
  import { beansStore } from '$lib/beans.svelte';
  import { worktreeStore } from '$lib/worktrees.svelte';
  import { agentStatusesStore } from '$lib/agentStatuses.svelte';
  import { statusColors, typeColors, typeBorders, priorityIndicators } from '$lib/styles';
  import { client } from '$lib/graphqlClient';
  import { gql } from 'urql';

  interface Props {
    bean: Bean;
    variant?: 'list' | 'board' | 'compact';
    selected?: boolean;
    onclick?: () => void;
  }

  let { bean, variant = 'list', selected = false, onclick }: Props = $props();

  const hasWorktree = $derived(variant !== 'compact' && worktreeStore.hasWorktree(bean.id));
  const agentRunning = $derived(hasWorktree && agentStatusesStore.isRunning(bean.id));
  const isArchivable = $derived(bean.status === 'completed' || bean.status === 'scrapped');

  const ARCHIVE_BEAN = gql`
    mutation ArchiveBean($id: ID!) {
      archiveBean(id: $id)
    }
  `;

  let archiving = $state(false);

  async function handleArchive(e: MouseEvent) {
    e.stopPropagation();
    archiving = true;
    await client.mutation(ARCHIVE_BEAN, { id: bean.id }).toPromise();
    archiving = false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onclick?.();
    }
  }
</script>

<!-- Using div instead of button so we can nest the archive <button> inside (HTML forbids button-in-button) -->
<div
  {onclick}
  onkeydown={handleKeydown}
  role="button"
  tabindex="0"
  class={[
    'relative w-full cursor-pointer overflow-hidden text-left transition-all',
    variant === 'board'
      ? 'p-3'
      : [
          'rounded-xs p-2',
          variant === 'compact' ? 'border-l-2' : 'border-l-3',
          hasWorktree
            ? 'border-l-success'
            : (typeBorders[bean.type] ?? 'border-l-type-task-border'),
          selected ? 'bg-accent/10 ring-1 ring-accent' : 'bg-surface hover:bg-surface-alt'
        ]
  ]}
>
  {#if hasWorktree}
    <div
      class={['absolute top-0 right-0 size-4 bg-success', agentRunning && 'bean-card-agent-pulse']}
      style="clip-path: polygon(0 0, 100% 0, 100% 100%)"
    ></div>
  {/if}

  {#if variant === 'board'}
    <!-- Board: two-row layout -->
    <div class="flex min-w-0 items-start gap-2">
      <span class="flex-1 text-sm leading-snug text-text">{bean.title}</span>
      {#if bean.priority && bean.priority !== 'normal' && priorityIndicators[bean.priority]}
        <span class={['shrink-0 text-xs', priorityIndicators[bean.priority]]}>
          {bean.priority}
        </span>
      {/if}
    </div>
    <div class="mt-1 flex items-center gap-2">
      <code class="text-[10px] text-text-faint">{bean.id.slice(-4)}</code>
      <span
        class={[
          'badge-sm',
          typeColors[bean.type] ?? 'bg-type-task-bg text-type-task-text'
        ]}
      >
        {bean.type}
      </span>
      {#each bean.tags as tag}
        <span class="badge-sm bg-surface-alt text-text-muted">{tag}</span>
      {/each}
      {#if isArchivable}
        <button
          class="ml-auto cursor-pointer icon-[uil--archive] size-3.5 text-text-faint transition-colors hover:text-text-muted disabled:opacity-50"
          title="Archive"
          onclick={handleArchive}
          disabled={archiving}
        ></button>
      {/if}
    </div>
  {:else}
    <!-- List / Compact: single-row layout -->
    <div class="flex min-w-0 items-center gap-2">
      <code
        class={['shrink-0 text-text-faint', variant === 'compact' ? 'text-[9px]' : 'text-[10px]']}
        >{bean.id.slice(-4)}</code
      >
      <span class={['flex-1 truncate text-text', variant === 'compact' ? 'text-xs' : 'text-sm']}
        >{bean.title}</span
      >
      {#each bean.tags as tag}
        <span class="shrink-0 badge-sm bg-surface-alt text-text-muted">{tag}</span>
      {/each}
      <span
        class={[
          'shrink-0 badge-sm',
          statusColors[bean.status] ?? 'bg-status-todo-bg text-status-todo-text'
        ]}
      >
        {bean.status}
      </span>
      {#if isArchivable}
        <button
          class="cursor-pointer icon-[uil--archive] size-3.5 shrink-0 text-text-faint transition-colors hover:text-text-muted disabled:opacity-50"
          title="Archive"
          onclick={handleArchive}
          disabled={archiving}
        ></button>
      {/if}
    </div>
  {/if}
</div>

<style>
  .bean-card-agent-pulse {
    animation: agent-pulse 2s ease-in-out infinite;
  }

  @keyframes agent-pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.3;
    }
  }
</style>
