<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title: string;
    onClose?: () => void;
    extra?: Snippet;
    actions?: Snippet;
  }

  let { title, onClose, extra, actions }: Props = $props();
</script>

<div class="toolbar">
  <span class="text-sm font-medium text-text">{title}</span>
  {#if extra}
    {@render extra()}
  {/if}
  <div class="flex-1"></div>
  {#if actions}
    {@render actions()}
  {/if}
  {#if onClose}
    <button onclick={onClose} class="btn-icon" title="Close">&#x2715;</button>
  {/if}
</div>
