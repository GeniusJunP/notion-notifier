import { writable } from 'svelte/store';
import type { Config } from './api';

export const configStore = writable<Config | null>(null);

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

export const activeRoute = writable<string>(window.location.pathname);

// Sync browser back/forward buttons
window.addEventListener('popstate', () => {
  activeRoute.set(window.location.pathname);
});

export function navigate(path: string) {
  window.history.pushState({}, '', path);
  activeRoute.set(path);
}

export const darkMode = writable<boolean>(false);

// Function to update dark mode and apply to DOM
export function setDarkMode(value: boolean) {
  darkMode.set(value);
  localStorage.setItem('darkMode', value.toString());
  document.documentElement.classList.toggle('dark', value);
}
