import { browser } from '$app/environment';

export const prerender = true;
export const ssr = false;

export function load() {
  let planningView: 'backlog' | 'board' = 'backlog';
  let selectedBeanId: string | null = null;
  let showPlanningChat = false;
  let showChanges = false;
  let filterText = '';
  let activeView: 'planning' | string = 'planning';

  if (browser) {
    const saved = localStorage.getItem('beans-planning-view');
    if (saved === 'backlog' || saved === 'board') {
      planningView = saved;
    }

    const params = new URLSearchParams(window.location.search);
    selectedBeanId = params.get('bean');

    showPlanningChat = localStorage.getItem('beans-planning-chat') === 'true';

    showChanges = localStorage.getItem('beans-changes-pane') === 'true';

    filterText = localStorage.getItem('beans-filter-text') ?? '';

    activeView = localStorage.getItem('beans-active-view') || 'planning';
  }

  return { planningView, selectedBeanId, showPlanningChat, showChanges, filterText, activeView };
}
