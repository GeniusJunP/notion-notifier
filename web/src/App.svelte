<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    navigate,
    activeRoute,
    toastStore,
    configStore,
    addToast,
    darkMode,
    setDarkMode,
  } from "./lib/store";
  import { api } from "./lib/api";
  import {
    LayoutDashboard,
    Bell,
    Calendar,
    Settings,
    History,
    Menu,
    X,
    ChevronRight,
    AlertCircle,
    CheckCircle2,
    Info,
    Sun,
    Moon,
  } from "lucide-svelte";

  // Routes
  import Dashboard from "./routes/Dashboard.svelte";
  import Notifications from "./routes/Notifications.svelte";
  import CalendarSync from "./routes/Calendar.svelte";
  import SystemSettings from "./routes/Settings.svelte";
  import NotificationHistory from "./routes/History.svelte";

  let isSidebarOpen = true;
  let isServiceActive = true;
  let healthInterval: ReturnType<typeof setInterval>;

  async function checkHealth() {
    try {
      await api.getDashboard();
      isServiceActive = true;
    } catch {
      isServiceActive = false;
    }
  }

  function toggleDarkMode() {
    darkMode.update((current) => {
      const newValue = !current;
      setDarkMode(newValue);
      return newValue;
    });
  }

  onMount(async () => {
    try {
      const cfg = await api.getConfig();
      configStore.set(cfg);
      isServiceActive = true;
    } catch (e) {
      addToast("設定の読み込みに失敗しました", "error");
      isServiceActive = false;
    }
    healthInterval = setInterval(checkHealth, 30000);

    // Initialize dark mode
    const saved = localStorage.getItem('darkMode');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initialDark = saved !== null ? saved === 'true' : prefersDark;
    setDarkMode(initialDark);

    // Listen for system dark mode changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
      // Only auto-switch if no manual preference is saved
      if (localStorage.getItem('darkMode') === null) {
        setDarkMode(e.matches);
      }
    };
    mediaQuery.addEventListener('change', handleChange);

    // Cleanup on destroy
    onDestroy(() => {
      mediaQuery.removeEventListener('change', handleChange);
    });
  });

  onDestroy(() => {
    if (healthInterval) clearInterval(healthInterval);
  });

  const navItems = [
    { path: "/", label: "ダッシュボード", icon: LayoutDashboard },
    { path: "/notifications", label: "通知設定", icon: Bell },
    { path: "/calendar", label: "カレンダー連携", icon: Calendar },
    { path: "/settings", label: "システム設定", icon: Settings },
    { path: "/history", label: "履歴", icon: History },
  ];

  $: currentComponent = (() => {
    switch ($activeRoute) {
      case "/":
        return Dashboard;
      case "/notifications":
        return Notifications;
      case "/calendar":
        return CalendarSync;
      case "/settings":
        return SystemSettings;
      case "/history":
        return NotificationHistory;
      default:
        return Dashboard;
    }
  })();
</script>

<div class="flex h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 overflow-hidden font-sans">
  <!-- Sidebar -->
  <aside
    class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transform transition-transform duration-300 lg:relative lg:translate-x-0 {isSidebarOpen
      ? 'translate-x-0'
      : '-translate-x-full'}"
  >
    <div
      class="flex items-center justify-between h-16 px-6 border-b border-gray-100 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <div
          class="w-8 h-8 bg-brand-600 rounded-lg flex items-center justify-center text-white shadow-lg shadow-brand-200 dark:shadow-brand-900"
        >
          <Bell size={18} />
        </div>
        <span class="font-bold text-lg tracking-tight">Notion Notifier</span>
      </div>
      <button
        class="lg:hidden text-gray-500"
        on:click={() => (isSidebarOpen = false)}
      >
        <X size={20} />
      </button>
    </div>

    <nav class="p-4 space-y-1 overflow-y-auto h-[calc(100%-4rem)]">
      {#each navItems as item}
        <button
          on:click={() => {
            navigate(item.path);
            if (window.innerWidth < 1024) isSidebarOpen = false;
          }}
          class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group {$activeRoute ===
          item.path
            ? 'bg-brand-50 dark:bg-brand-900/20 text-brand-700 dark:text-brand-300 font-semibold'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100'}"
        >
          <div class="transition-transform duration-200 group-hover:scale-110">
            <svelte:component
              this={item.icon}
              size={20}
              strokeWidth={$activeRoute === item.path ? 2.5 : 2}
            />
          </div>
          <span>{item.label}</span>
          {#if $activeRoute === item.path}
            <div class="ml-auto w-1.5 h-1.5 rounded-full bg-brand-600"></div>
          {/if}
        </button>
      {/each}
    </nav>
  </aside>

  <!-- Main Content -->
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <header
      class="h-16 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center px-4 md:px-8 justify-between sticky top-0 z-40"
    >
      <div class="flex items-center gap-4">
        <button
          class="lg:hidden p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          on:click={() => (isSidebarOpen = true)}
        >
          <Menu size={20} />
        </button>
        <h1 class="text-xl font-bold text-gray-800 dark:text-gray-200">
          {navItems.find((n) => n.path === $activeRoute)?.label || "Dashboard"}
        </h1>
      </div>

      <div class="flex items-center gap-4">
        <button
          on:click={toggleDarkMode}
          class="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          aria-label="Toggle dark mode"
        >
          {#if $darkMode}
            <Sun size={20} />
          {:else}
            <Moon size={20} />
          {/if}
        </button>
        <div
          class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium border {isServiceActive
            ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 border-green-100 dark:border-green-700'
            : 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 border-red-100 dark:border-red-700'}"
        >
          <div
            class="w-2 h-2 rounded-full {isServiceActive
              ? 'bg-green-500 animate-pulse'
              : 'bg-red-500'}"
          ></div>
          {isServiceActive ? "Service Active" : "Service Offline"}
        </div>
      </div>
    </header>

    <main class="flex-1 overflow-y-auto p-4 md:p-8 custom-scrollbar">
      <div class="max-w-6xl mx-auto">
        <svelte:component this={currentComponent} />
      </div>
    </main>
  </div>

  <!-- Toast Container -->
  <div
    class="fixed bottom-6 right-6 z-[100] flex flex-col gap-3 max-w-sm w-full pointer-events-none"
  >
    {#each $toastStore as toast (toast.id)}
      <div
        class="pointer-events-auto flex items-start gap-3 p-4 bg-white dark:bg-gray-800 rounded-2xl shadow-2xl border border-gray-100 dark:border-gray-700 animate-in slide-in-from-right fade-in duration-300"
      >
        <div class="mt-0.5">
          {#if toast.type === "error"}
            <AlertCircle size={20} class="text-red-500" />
          {:else if toast.type === "success"}
            <CheckCircle2 size={20} class="text-green-500" />
          {:else}
            <Info size={20} class="text-brand-500" />
          {/if}
        </div>
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-gray-100 leading-tight">
            {toast.message}
          </p>
        </div>
        <button
          on:click={() =>
            toastStore.update((toasts) =>
              toasts.filter((t) => t.id !== toast.id),
            )}
          class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
        >
          <X size={16} />
        </button>
      </div>
    {/each}
  </div>
</div>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: #e2e8f0;
    border-radius: 10px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #cbd5e1;
  }
</style>
