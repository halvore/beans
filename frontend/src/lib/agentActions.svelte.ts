import { client } from '$lib/graphqlClient';
import {
  AgentActionsDocument,
  ExecuteAgentActionDocument,
  type AgentActionFieldsFragment,
} from './graphql/generated';

export type AgentAction = AgentActionFieldsFragment;

export class AgentActionsStore {
  actions = $state<AgentAction[]>([]);
  executingAction = $state<string | null>(null);
  #wasAgentBusy = false;

  async fetch(beanId: string) {
    const result = await client.query(AgentActionsDocument, { beanId }).toPromise();
    if (result.error) {
      console.error('Failed to fetch agent actions:', result.error);
      return;
    }
    if (result.data?.agentActions) {
      this.actions = result.data.agentActions;
    }
  }

  /**
   * Call this reactively with the current agent busy state.
   * Automatically re-fetches actions when the agent transitions from busy to idle.
   */
  notifyAgentStatus(beanId: string, busy: boolean) {
    if (this.#wasAgentBusy && !busy) {
      this.fetch(beanId);
    }
    this.#wasAgentBusy = busy;
  }

  async execute(beanId: string, actionId: string) {
    this.executingAction = actionId;
    try {
      await client.mutation(ExecuteAgentActionDocument, { beanId, actionId }).toPromise();
    } finally {
      this.executingAction = null;
    }
  }
}
