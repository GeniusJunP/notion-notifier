import { writable } from 'svelte/store';
import { api, getErrorMessage, type Config, type DashboardData } from './api';

export const configStore = writable<Config | null>(null);

export const dashboardStore = writable<DashboardData | null>(null);
export const serviceActiveStore = writable<boolean>(true);

export function createHealthPoller() {
  let interval: ReturnType<typeof setInterval>;
  
  const check = async () => {
    try {
      const data = await api.getDashboard();
      dashboardStore.set(data);
      serviceActiveStore.set(true);
    } catch {
      serviceActiveStore.set(false);
      dashboardStore.set(null);
    }
  };
  
  return {
    start: () => {
      check();
      interval = setInterval(check, 30000);
    },
    stop: () => {
      if (interval) clearInterval(interval);
    },
    forceCheck: check
  };
}

export const healthPoller = createHealthPoller();

export interface Toast {
  id: string;
  message: string;
  type: 'success' | 'error' | 'info';
}

export const toastStore = writable<Toast[]>([]);

export function addToast(message: string, type: Toast['type'] = 'info') {
  const id = Math.random().toString(36).substring(2, 9);
  toastStore.update((toasts) => [...toasts, { id, message, type }]);
  setTimeout(() => {
    toastStore.update((toasts) => toasts.filter((t) => t.id !== id));
  }, 5000);
}

// Hook api errors to the toast system
api.onError = (msg: string) => addToast(msg, 'error');

export const activeRoute = writable<string>(window.location.pathname);

// Sync browser back/forward buttons
window.addEventListener('popstate', () => {
  activeRoute.set(window.location.pathname);
});

export function navigate(path: string) {
  window.history.pushState({}, '', path);
  activeRoute.set(path);
}

function createDarkModeStore() {
  const { subscribe, set, update } = writable<boolean>(false);
  return {
    subscribe,
    set: (value: boolean) => {
      localStorage.setItem('darkMode', value.toString());
      document.documentElement.classList.toggle('dark', value);
      set(value);
    },
    update: (updater: (v: boolean) => boolean) => {
      update((current) => {
        const value = updater(current);
        localStorage.setItem('darkMode', value.toString());
        document.documentElement.classList.toggle('dark', value);
        return value;
      });
    },
    init: () => {
      const saved = localStorage.getItem('darkMode');
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      const initialDark = saved !== null ? saved === 'true' : prefersDark;
      
      const applyTheme = (v: boolean) => {
        document.documentElement.classList.toggle('dark', v);
        set(v);
      };
      
      applyTheme(initialDark);
      
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      const listener = (e: MediaQueryListEvent) => {
        if (localStorage.getItem('darkMode') === null) {
          applyTheme(e.matches);
        }
      };
      mediaQuery.addEventListener('change', listener);
      return () => mediaQuery.removeEventListener('change', listener);
    }
  };
}

export const darkMode = createDarkModeStore();

interface SaveConfigOptions {
  successMessage?: string;
  errorMessage: string;
  onSaved?: (saved: Config) => Promise<void> | void;
}

export async function saveConfig(
  cfg: Config | null,
  options: SaveConfigOptions,
): Promise<Config | null> {
  if (!cfg) {
    return null;
  }

  try {
    const saved = await api.updateConfig(cfg);
    configStore.set(saved);
    if (options.onSaved) {
      await options.onSaved(saved);
    }
    if (options.successMessage) {
      addToast(options.successMessage, 'success');
    }
    return saved;
  } catch (e: unknown) {
    const detail = getErrorMessage(e);
    addToast(`${options.errorMessage}: ${detail}`, 'error');
    return null;
  }
}

interface SyncNotionOptions {
  successMessage?: (count: number) => string;
  errorMessage?: string;
  onSynced?: (count: number) => Promise<void> | void;
}

export async function syncNotion(options: SyncNotionOptions = {}): Promise<number | null> {
  try {
    const res = await api.syncNotion();
    const successMessage = options.successMessage
      ? options.successMessage(res.count)
      : `${res.count}件のイベントを同期しました`;
    addToast(successMessage, 'success');
    if (options.onSynced) {
      await options.onSynced(res.count);
    }
    return res.count;
  } catch (e: unknown) {
    const detail = getErrorMessage(e);
    addToast(`${options.errorMessage ?? '同期に失敗しました'}: ${detail}`, 'error');
    return null;
  }
}
