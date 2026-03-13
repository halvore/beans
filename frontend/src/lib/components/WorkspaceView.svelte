<script lang="ts">
  import { AgentChatStore } from '$lib/agentChat.svelte';
  import { changesStore } from '$lib/changes.svelte';
  import { ui } from '$lib/uiState.svelte';
  import { worktreeStore, MAIN_WORKSPACE_ID } from '$lib/worktrees.svelte';
  import { onDestroy } from 'svelte';
  import SplitPane from './SplitPane.svelte';
  import AgentChat from './AgentChat.svelte';
  import ChangesPane from './ChangesPane.svelte';

  import TerminalPane from './TerminalPane.svelte';
  import ViewToolbar from './ViewToolbar.svelte';
  import AgentActions from './AgentActions.svelte';

  interface Props {
    worktreeId: string;
  }

  let { worktreeId }: Props = $props();

  const agentStore = new AgentChatStore();

  $effect(() => {
    agentStore.subscribe(worktreeId);
  });

  $effect(() => {
    changesStore.startPolling(worktreePath);
    return () => changesStore.stopPolling();
  });

  onDestroy(() => {
    agentStore.unsubscribe();
  });

  const agentBusy = $derived(agentStore.session?.status === 'RUNNING');

  const worktreePath = $derived(
    worktreeId === MAIN_WORKSPACE_ID
      ? undefined
      : worktreeStore.worktrees.find((wt) => wt.id === worktreeId)?.path
  );
</script>

{#snippet changesPanel()}
  <ChangesPane path={worktreePath} />
{/snippet}

{#snippet agentChatPanel()}
  <AgentChat beanId={worktreeId} store={agentStore} />
{/snippet}

{#snippet terminalPanel()}
  {#if ui.terminalInitialized}
    <TerminalPane sessionId={worktreeId} />
  {/if}
{/snippet}

{#snippet mainContent()}
  <SplitPane
    direction="horizontal"
    panels={[
      { content: agentChatPanel },
      { content: changesPanel, size: 420, collapsed: !ui.showChanges, persistKey: 'workspace-changes' }
    ]}
  />
{/snippet}

<div class="flex h-full flex-col">
  <ViewToolbar>
    {#snippet right()}
      <AgentActions beanId={worktreeId} {agentBusy} />
    {/snippet}
  </ViewToolbar>

  <div class="flex min-h-0 flex-1 flex-col">
    <SplitPane
      direction="vertical"
      panels={[
        { content: mainContent },
        { content: terminalPanel, size: 300, collapsed: !ui.showTerminal, persistKey: 'workspace-terminal' }
      ]}
    />
  </div>
</div>
