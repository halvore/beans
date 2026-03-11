<script lang="ts">
  import type { SubagentActivity } from '$lib/agentChat.svelte';

  interface Props {
    beanId: string;
    isRunning: boolean;
    hasMessages: boolean;
    agentMode: 'plan' | 'act';
    systemStatus: string | null;
    subagentActivities: SubagentActivity[];
    onSend: (message: string) => void;
    onStop: () => void;
    onSetMode: (mode: 'plan' | 'act') => void;
    onCompact: () => void;
    onClear: () => void;
  }

  let {
    beanId,
    isRunning,
    hasMessages,
    agentMode,
    systemStatus,
    subagentActivities,
    onSend,
    onStop,
    onSetMode,
    onCompact,
    onClear
  }: Props = $props();

  const inputStorageKey = $derived(`agent-chat-input:${beanId}`);
  let inputText = $state('');

  // Load persisted composer input when beanId changes
  $effect(() => {
    inputText = localStorage.getItem(inputStorageKey) ?? '';
  });

  // Persist composer input to localStorage so it survives navigation/reloads
  $effect(() => {
    if (inputText) {
      localStorage.setItem(inputStorageKey, inputText);
    } else {
      localStorage.removeItem(inputStorageKey);
    }
  });

  function send() {
    const text = inputText.trim();
    if (!text) return;
    inputText = '';
    onSend(text);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  }
</script>

<div class="border-t border-border bg-surface p-3">
  {#if isRunning}
    <div class="flex items-center gap-2 px-1 pb-2 text-text-muted">
      <span class="agent-spinner"></span>
      <span class="font-mono text-xs">
        {#if subagentActivities.length > 0}
          {subagentActivities.length} subagent{subagentActivities.length > 1 ? 's' : ''} working...
        {:else if systemStatus}
          Agent is {systemStatus}...
        {:else}
          Agent is working...
        {/if}
      </span>
    </div>
  {/if}
  <div class="flex items-end gap-2">
    <textarea
      bind:value={inputText}
      onkeydown={handleKeydown}
      placeholder="Send a message..."
      rows={1}
      class="flex-1 resize-none rounded border border-border bg-surface-alt px-3 py-2 font-mono text-sm
				text-text placeholder:text-text-faint
				focus:border-accent focus:ring-2 focus:ring-accent/40 focus:outline-none"
    ></textarea>

    <button
      onclick={send}
      disabled={!inputText.trim()}
      class="inline-flex shrink-0 items-center gap-1.5 rounded bg-accent px-3 py-2 font-mono
				text-sm text-accent-text transition-colors hover:bg-accent/90
				disabled:cursor-not-allowed disabled:opacity-50"
    >
      <span class="icon-[uil--message] size-4"></span>
      Send
    </button>

    {#if isRunning}
      <button
        onclick={onStop}
        class="inline-flex shrink-0 items-center gap-1.5 rounded bg-danger px-3 py-2 font-mono
					text-sm text-white transition-colors hover:bg-danger/90"
      >
        <span class="icon-[uil--stop-circle] size-4"></span>
        Stop
      </button>
    {/if}
  </div>

  <!-- Mode toggle + Clear -->
  <div class="flex items-center gap-3 pt-2">
    <div class={['flex', isRunning && 'pointer-events-none opacity-50']}>
      <button
        onclick={() => onSetMode('plan')}
        disabled={isRunning}
        class={[
          'btn-tab-sm rounded-l',
          agentMode === 'plan'
            ? 'border-warning/30 bg-warning/10 text-warning'
            : 'btn-tab-sm-inactive'
        ]}
      >
        <span class="icon-[uil--eye] size-3"></span>
        Plan
      </button>
      <button
        onclick={() => onSetMode('act')}
        disabled={isRunning}
        class={[
          'btn-tab-sm rounded-r border-l-0',
          agentMode === 'act'
            ? 'border-success/30 bg-success/10 text-success'
            : 'btn-tab-sm-inactive'
        ]}
      >
        <span class="icon-[uil--play] size-3"></span>
        Act
      </button>
    </div>

    <div
      class={['flex', (isRunning || !hasMessages) && 'pointer-events-none opacity-30']}
    >
      <button
        onclick={onCompact}
        disabled={isRunning || !hasMessages}
        class="btn-tab-sm btn-tab-sm-inactive rounded-l"
      >
        <span class="icon-[uil--compress-arrows] size-3"></span>
        Compact
      </button>
      <button
        onclick={onClear}
        disabled={isRunning || !hasMessages}
        class="btn-tab-sm btn-tab-sm-inactive rounded-r border-l-0"
      >
        <span class="icon-[uil--trash-alt] size-3"></span>
        Clear
      </button>
    </div>
  </div>
</div>

<style>
  .agent-spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
