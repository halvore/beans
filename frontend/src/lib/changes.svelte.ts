import { gql } from 'urql';
import { client } from './graphqlClient';

export interface FileChange {
	path: string;
	status: string;
	additions: number;
	deletions: number;
	staged: boolean;
}

const FILE_CHANGES_QUERY = gql`
	query FileChanges($path: String) {
		fileChanges(path: $path) {
			path
			status
			additions
			deletions
			staged
		}
	}
`;

class ChangesStore {
	changes = $state<FileChange[]>([]);
	loading = $state(false);
	#intervalId: ReturnType<typeof setInterval> | null = null;
	#currentPath: string | null = null;

	async fetch(path?: string): Promise<void> {
		const result = await client
			.query(FILE_CHANGES_QUERY, { path: path ?? null })
			.toPromise();

		if (result.error) {
			console.error('Failed to fetch file changes:', result.error);
			return;
		}

		this.changes = result.data?.fileChanges ?? [];
	}

	startPolling(path?: string, intervalMs = 3000): void {
		this.stopPolling();
		this.#currentPath = path ?? null;
		this.fetch(path);
		this.#intervalId = setInterval(() => this.fetch(this.#currentPath ?? undefined), intervalMs);
	}

	stopPolling(): void {
		if (this.#intervalId) {
			clearInterval(this.#intervalId);
			this.#intervalId = null;
		}
	}
}

export const changesStore = new ChangesStore();
