<script lang="ts">
  import type { Bean } from '$lib/beans.svelte';
  import { ui } from '$lib/uiState.svelte';
  import { worktreeStore } from '$lib/worktrees.svelte';
  import SplitPane from './SplitPane.svelte';
  import AgentChat from './AgentChat.svelte';
  import BeanPane from './BeanPane.svelte';
  import ChangesPane from './ChangesPane.svelte';
  import PaneHeader from './PaneHeader.svelte';

  interface Props {
    bean: Bean;
  }

  let { bean }: Props = $props();

  const worktreePath = $derived(
    worktreeStore.worktrees.find((wt) => wt.beanId === bean.id)?.path
  );
</script>

{#snippet agentToolbar()}
  <PaneHeader title="Agent">
    {#snippet actions()}
      <button
        onclick={() => ui.toggleChanges()}
        class={['btn-toggle-icon', ui.showChanges ? 'btn-toggle-active' : 'btn-toggle-inactive']}
        title={ui.showChanges ? 'Hide changes' : 'Show changes'}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill="currentColor"
          class="h-4 w-4"
        >
          <path
            d="M18 2H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h10c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-1 9h-3v3h-2v-3H9V9h3V6h2v3h3v2zM4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm12 9H10v-2h6v2z"
          />
        </svg>
      </button>
    {/snippet}
  </PaneHeader>
{/snippet}

<SplitPane direction="horizontal" side="end" persistKey="workspace-chat-width" initialSize={480}>
  {#snippet aside()}
    {#if ui.showChanges}
      <SplitPane
        direction="horizontal"
        side="end"
        persistKey="workspace-changes-chat-split"
        initialSize={480}
      >
        {#snippet children()}
          <ChangesPane path={worktreePath} />
        {/snippet}
        {#snippet aside()}
          <div class="flex h-full flex-col border-l border-border bg-surface">
            {@render agentToolbar()}
            <div class="min-h-0 flex-1">
              <AgentChat beanId={bean.id} />
            </div>
          </div>
        {/snippet}
      </SplitPane>
    {:else}
      <div class="flex h-full flex-col border-l border-border bg-surface">
        {@render agentToolbar()}
        <div class="min-h-0 flex-1">
          <AgentChat beanId={bean.id} />
        </div>
      </div>
    {/if}
  {/snippet}

  {#snippet children()}
    <BeanPane {bean} onSelect={(b) => ui.selectBean(b)} onEdit={(b) => ui.openEditForm(b)} />
  {/snippet}
</SplitPane>
