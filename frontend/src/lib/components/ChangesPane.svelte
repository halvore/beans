<script lang="ts">
  import { gql } from 'urql';
  import { onMount, onDestroy } from 'svelte';
  import { changesStore, type FileChange } from '$lib/changes.svelte';
  import { ui } from '$lib/uiState.svelte';
  import { client } from '$lib/graphqlClient';
  import PaneHeader from '$lib/components/PaneHeader.svelte';
  import SplitPane from '$lib/components/SplitPane.svelte';

  interface AgentAction {
    id: string;
    label: string;
    description: string | null;
  }

  interface Props {
    path?: string;
    beanId?: string;
    agentBusy?: boolean;
  }

  let { path, beanId, agentBusy = false }: Props = $props();

  const AGENT_ACTIONS_QUERY = gql`
    query AgentActions($beanId: ID!) {
      agentActions(beanId: $beanId) {
        id
        label
        description
      }
    }
  `;

  const EXECUTE_AGENT_ACTION = gql`
    mutation ExecuteAgentAction($beanId: ID!, $actionId: ID!) {
      executeAgentAction(beanId: $beanId, actionId: $actionId)
    }
  `;

  const FILE_DIFF_QUERY = gql`
    query FileDiff($filePath: String!, $staged: Boolean!, $path: String) {
      fileDiff(filePath: $filePath, staged: $staged, path: $path)
    }
  `;

  let actions = $state<AgentAction[]>([]);
  let executingAction = $state<string | null>(null);

  // Diff view state
  let selectedFile = $state<{ path: string; staged: boolean } | null>(null);
  let diffContent = $state<string>('');
  let diffLoading = $state(false);

  async function fetchActions() {
    if (!beanId) return;
    const result = await client.query(AGENT_ACTIONS_QUERY, { beanId }).toPromise();
    if (result.error) {
      console.error('Failed to fetch agent actions:', result.error);
      return;
    }
    if (result.data?.agentActions) {
      actions = result.data.agentActions;
    }
  }

  async function executeAction(actionId: string) {
    if (!beanId || agentBusy) return;
    executingAction = actionId;
    try {
      await client.mutation(EXECUTE_AGENT_ACTION, { beanId, actionId }).toPromise();
    } finally {
      executingAction = null;
    }
  }

  function selectFile(change: FileChange) {
    const key = { path: change.path, staged: change.staged };
    // Toggle off if clicking the same file
    if (selectedFile?.path === key.path && selectedFile?.staged === key.staged) {
      selectedFile = null;
      diffContent = '';
      return;
    }
    selectedFile = key;
    fetchDiff(key.path, key.staged);
  }

  async function fetchDiff(filePath: string, staged: boolean) {
    diffLoading = true;
    const result = await client
      .query(FILE_DIFF_QUERY, { filePath, staged, path: path ?? null })
      .toPromise();

    // Guard against stale response if user clicked a different file while loading
    if (selectedFile?.path !== filePath || selectedFile?.staged !== staged) return;

    if (result.error) {
      console.error('Failed to fetch diff:', result.error);
      diffContent = '';
    } else {
      diffContent = result.data?.fileDiff ?? '';
    }
    diffLoading = false;
  }

  // Re-fetch actions when beanId changes
  $effect(() => {
    if (beanId) {
      fetchActions();
    }
  });

  // Re-fetch actions when agent transitions to idle
  let wasAgentBusy = $state(false);
  $effect(() => {
    if (wasAgentBusy && !agentBusy) {
      fetchActions();
    }
    wasAgentBusy = agentBusy;
  });

  // Clear selection when the selected file disappears from the changes list
  $effect(() => {
    if (selectedFile) {
      const stillExists = changesStore.changes.some(
        (c) => c.path === selectedFile!.path && c.staged === selectedFile!.staged
      );
      if (!stillExists) {
        selectedFile = null;
        diffContent = '';
      }
    }
  });

  const stagedChanges = $derived(changesStore.changes.filter((c) => c.staged));
  const unstagedChanges = $derived(changesStore.changes.filter((c) => !c.staged));
  const totalCount = $derived(changesStore.changes.length);

  const diffLines = $derived(diffContent ? diffContent.split('\n') : []);

  onMount(() => {
    changesStore.startPolling(path);
  });

  onDestroy(() => {
    changesStore.stopPolling();
  });

  function statusColor(status: string): string {
    switch (status) {
      case 'added':
      case 'untracked':
        return 'text-success';
      case 'deleted':
        return 'text-danger';
      case 'renamed':
        return 'text-accent';
      default:
        return 'text-warning';
    }
  }

  function statusLabel(status: string): string {
    switch (status) {
      case 'modified':
        return 'M';
      case 'added':
        return 'A';
      case 'deleted':
        return 'D';
      case 'untracked':
        return '?';
      case 'renamed':
        return 'R';
      default:
        return '?';
    }
  }

  function fileName(filePath: string): string {
    return filePath.split('/').pop() ?? filePath;
  }

  function dirName(filePath: string): string {
    const parts = filePath.split('/');
    if (parts.length <= 1) return '';
    return parts.slice(0, -1).join('/') + '/';
  }

  function isFileSelected(change: FileChange): boolean {
    return selectedFile?.path === change.path && selectedFile?.staged === change.staged;
  }

  function diffLineClass(line: string): string {
    if (line.startsWith('+') && !line.startsWith('+++')) return 'diff-add';
    if (line.startsWith('-') && !line.startsWith('---')) return 'diff-del';
    if (line.startsWith('@@')) return 'diff-hunk';
    return '';
  }
</script>

{#snippet fileRow(change: FileChange)}
  <button
    class={[
      'flex w-full cursor-pointer items-center gap-1.5 px-3 py-0.5 text-left hover:bg-surface-alt',
      isFileSelected(change) && 'bg-surface-alt'
    ]}
    onclick={() => selectFile(change)}
  >
    <span class={['w-3 shrink-0 text-center font-mono font-bold', statusColor(change.status)]}>
      {statusLabel(change.status)}
    </span>
    <span class="min-w-0 flex-1 truncate" title={change.path}>
      <span class="text-text-muted">{dirName(change.path)}</span><span class="text-text">{fileName(change.path)}</span>
    </span>
    {#if change.additions > 0 || change.deletions > 0}
      <span class="shrink-0 font-mono">
        {#if change.additions > 0}<span class="text-success">+{change.additions}</span>{/if}
        {#if change.deletions > 0}<span class={[change.additions > 0 && 'ml-1', 'text-danger']}>-{change.deletions}</span>{/if}
      </span>
    {/if}
  </button>
{/snippet}

{#snippet fileList()}
  <div class="flex-1 overflow-auto">
    {#if totalCount === 0}
      <p class="px-3 py-4 text-center text-text-muted">No changes</p>
    {:else}
      {#if stagedChanges.length > 0}
        <div class="px-3 pt-2 pb-1 font-medium text-text-muted">Staged</div>
        {#each stagedChanges as change (change.path + ':staged')}
          {@render fileRow(change)}
        {/each}
      {/if}

      {#if unstagedChanges.length > 0}
        {#if stagedChanges.length > 0}
          <div class="px-3 pt-2 pb-1 font-medium text-text-muted">Unstaged</div>
        {/if}
        {#each unstagedChanges as change (change.path + ':unstaged')}
          {@render fileRow(change)}
        {/each}
      {/if}
    {/if}
  </div>
{/snippet}

{#snippet diffView()}
  <div class="flex h-full flex-col border-t border-border">
    <div class="flex items-center justify-between px-3 py-1.5">
      <span class="truncate font-mono text-xs text-text-muted">
        {selectedFile?.path}
        {#if selectedFile?.staged}
          <span class="text-text-faint">(staged)</span>
        {/if}
      </span>
      <button
        class="btn-icon shrink-0 cursor-pointer"
        onclick={() => { selectedFile = null; diffContent = ''; }}
        aria-label="Close diff"
      >
        <span class="iconify lucide--x size-3.5"></span>
      </button>
    </div>
    <div class="flex-1 overflow-auto bg-surface-alt">
      {#if diffLoading}
        <p class="px-3 py-4 text-center text-text-muted">Loading...</p>
      {:else if diffContent === ''}
        <p class="px-3 py-4 text-center text-text-muted">No diff available</p>
      {:else}
        <pre class="font-mono text-xs leading-relaxed">{#each diffLines as line}<span class={diffLineClass(line)}>{line}
</span>{/each}</pre>
      {/if}
    </div>
  </div>
{/snippet}

<div class="flex h-full flex-col border-l border-border bg-surface">
  <PaneHeader title="Status" onClose={() => ui.toggleChanges()}>
    {#snippet extra()}
      {#if totalCount > 0}
        <span class="ml-1 text-sm text-text-muted">({totalCount})</span>
      {/if}
    {/snippet}
  </PaneHeader>

  {#if selectedFile}
    <SplitPane direction="vertical" side="start" initialSize={200} minSize={60} persistKey="changes-diff">
      {#snippet children()}
        {@render diffView()}
      {/snippet}
      {#snippet aside()}
        {@render fileList()}
      {/snippet}
    </SplitPane>
  {:else}
    {@render fileList()}
  {/if}

  {#if beanId && actions.length > 0}
    <div class="flex gap-2 border-t border-border px-3 py-2">
      {#each actions as action (action.id)}
        <button
          class={[
            'flex-1 rounded border border-border px-3 py-1.5 text-sm font-medium transition-colors',
            agentBusy || executingAction
              ? 'cursor-not-allowed text-text-faint'
              : 'cursor-pointer text-text-muted hover:bg-surface-alt hover:text-text'
          ]}
          disabled={agentBusy || !!executingAction}
          title={action.description ?? undefined}
          onclick={() => executeAction(action.id)}
        >
          {action.label}
        </button>
      {/each}
    </div>
  {/if}
</div>
