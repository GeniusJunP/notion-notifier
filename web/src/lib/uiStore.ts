import { writable } from 'svelte/store';

// UI Layout State
function createSidebarStore() {
  const isLg = typeof window !== 'undefined' ? window.innerWidth >= 1024 : true;
  const { subscribe, set, update } = writable<boolean>(isLg);

  return {
    subscribe,
    open: () => set(true),
    close: () => set(false),
    toggle: () => update(n => !n),
    set
  };
}

export const sidebarOpen = createSidebarStore();

// UI Modal State
interface GuideModalState {
  isOpen: boolean;
  title: string;
  content: string;
}

function createGuideModalStore() {
  const { subscribe, set, update } = writable<GuideModalState>({
    isOpen: false,
    title: '',
    content: ''
  });

  return {
    subscribe,
    open: (title: string, content: string) => set({ isOpen: true, title, content }),
    close: () => update(s => ({ ...s, isOpen: false })),
  };
}

export const guideModal = createGuideModalStore();
