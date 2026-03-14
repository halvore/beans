import { pipe, subscribe } from 'wonka';
import { client } from './graphqlClient';
import { generateWorkspaceName } from '$lib/nameGenerator';
import {
  WorktreesChangedDocument,
  CreateWorktreeDocument,
  RemoveWorktreeDocument,
  WorktreesDocument,
  type WorktreeFieldsFragment,
} from './graphql/generated';

export const MAIN_WORKSPACE_ID = '__central__';

/** Public Worktree type: codegen fragment with beanIds derived from nested beans. */
export interface Worktree extends Omit<WorktreeFieldsFragment, 'beans'> {
  beanIds: string[];
}

function mapWorktree(raw: WorktreeFieldsFragment): Worktree {
  const { beans, ...rest } = raw;
  return { ...rest, beanIds: beans.map((b) => b.id) };
}

export interface WorktreeStatus {
  hasChanges: boolean;
  hasUnmergedCommits: boolean;
}

class WorktreeStore {
  worktrees = $state<Worktree[]>([]);
  initialized = $state(false);
  loading = $state(false);
  error = $state<string | null>(null);

  #unsubscribe: (() => void) | null = null;

  subscribe(): void {
    if (this.#unsubscribe) return;

    const { unsubscribe } = pipe(
      client.subscription(WorktreesChangedDocument, {}),
      subscribe((result) => {
        if (result.error) {
          console.error('Worktree subscription error:', result.error);
          this.error = result.error.message;
          this.initialized = true;
          return;
        }

        const wts = result.data?.worktreesChanged;
        if (wts) {
          this.worktrees = wts.map(mapWorktree);
          this.initialized = true;
        }
      })
    );

    this.#unsubscribe = unsubscribe;
  }

  unsubscribe(): void {
    if (this.#unsubscribe) {
      this.#unsubscribe();
      this.#unsubscribe = null;
    }
  }

  async createWorktree(): Promise<Worktree | null> {
    this.loading = true;
    this.error = null;

    const name = generateWorkspaceName();
    const result = await client.mutation(CreateWorktreeDocument, { name }).toPromise();

    this.loading = false;

    if (result.error) {
      this.error = result.error.message;
      return null;
    }

    const raw = result.data?.createWorktree ?? null;
    const wt = raw ? mapWorktree(raw) : null;

    // Eagerly add to local state so the layout guard doesn't redirect
    // back to planning before the subscription delivers the update.
    if (wt && !this.worktrees.some((w) => w.id === wt.id)) {
      this.worktrees = [...this.worktrees, wt];
    }

    return wt;
  }

  async removeWorktree(id: string): Promise<boolean> {
    this.loading = true;
    this.error = null;

    // Eagerly remove from local state so the sidebar updates immediately
    // without waiting for the subscription to deliver the new list.
    const previous = this.worktrees;
    this.worktrees = this.worktrees.filter((wt) => wt.id !== id);

    const result = await client.mutation(RemoveWorktreeDocument, { id }).toPromise();

    this.loading = false;

    if (result.error) {
      // Restore on failure so the item reappears
      this.worktrees = previous;
      this.error = result.error.message;
      return false;
    }

    return true;
  }

  hasWorktree(id: string): boolean {
    return this.worktrees.some((wt) => wt.id === id);
  }

  /** Return the worktree ID that contains the given bean, or null. */
  worktreeForBean(beanId: string): string | null {
    return this.worktrees.find((wt) => wt.beanIds.includes(beanId))?.id ?? null;
  }

  /** Fetch fresh git status for a specific worktree (on-demand, not cached). */
  async getWorktreeStatus(id: string): Promise<WorktreeStatus | null> {
    const result = await client.query(WorktreesDocument, {}, { requestPolicy: 'network-only' }).toPromise();
    if (result.error) return null;
    const wt = result.data?.worktrees?.find((w) => w.id === id);
    return wt ? { hasChanges: wt.hasChanges, hasUnmergedCommits: wt.hasUnmergedCommits } : null;
  }
}

export const worktreeStore = new WorktreeStore();
