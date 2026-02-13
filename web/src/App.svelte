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
    Clock,
    Settings,
    History,
    Menu,
    X,
    AlertCircle,
    CheckCircle2,
    Info,
    Sun,
    Moon,
    Activity,
    RefreshCcw,
    Database,
    BellOff,
    VolumeX,
    Loader2,
  } from "lucide-svelte";

  // Routes
  import Dashboard from "./routes/Dashboard.svelte";
  import Notifications from "./routes/Notifications.svelte";
  import CalendarSync from "./routes/Calendar.svelte";
  import SystemSettings from "./routes/Settings.svelte";
  import NotificationHistory from "./routes/History.svelte";
  import type { DashboardData } from "./lib/api";

  let isSidebarOpen = true;
  let isServiceActive = true;
  let isSyncing = false;
  let dashboardData: DashboardData | null = null;
  let healthInterval: ReturnType<typeof setInterval>;
  let clockInterval: ReturnType<typeof setInterval>;
  let now = new Date();
  let mediaQuery: MediaQueryList | null = null;
  let mediaQueryListener: ((e: MediaQueryListEvent) => void) | null = null;
  const mainNavId = "main-navigation";

  const dateFormatter = new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  const weekdayFormatter = new Intl.DateTimeFormat("ja-JP", {
    weekday: "short",
  });
  const timeFormatter = new Intl.DateTimeFormat("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });

  function updateClock() {
    now = new Date();
  }

  function closeSidebar() {
    isSidebarOpen = false;
  }

  function openSidebar() {
    isSidebarOpen = true;
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key === "Escape" && window.innerWidth < 1024 && isSidebarOpen) {
      closeSidebar();
    }
  }

  async function checkHealth() {
    try {
      dashboardData = await api.getDashboard();
      isServiceActive = true;
    } catch {
      isServiceActive = false;
      dashboardData = null;
    }
  }

  async function handleSync() {
    if (isSyncing) return;
    isSyncing = true;
    try {
      const res = await api.syncNotion();
      addToast(`${res.count}件のイベントを同期しました`, "success");
      await checkHealth();
    } catch (e) {
      addToast("同期に失敗しました", "error");
    } finally {
      isSyncing = false;
    }
  }

  async function handleToggleSnooze() {
    configStore.update(cfg => {
      if (!cfg) return cfg;
      const isSnoozed = dashboardData?.snooze_active;
      
      const updatedCfg = { ...cfg };
      if (isSnoozed) {
        updatedCfg.snooze_until = "";
      } else {
        const snoozeDate = new Date(Date.now() + 60 * 60 * 1000);
        updatedCfg.snooze_until = snoozeDate.toISOString();
      }
      
      api.updateConfig(updatedCfg).then(saved => {
        configStore.set(saved);
        addToast(isSnoozed ? "スヌーズを解除しました" : "1時間スヌーズします", "success");
        checkHealth();
      }).catch(() => {
        addToast("設定の更新に失敗しました", "error");
      });
      
      return updatedCfg;
    });
  }

  function toggleDarkMode() {
    darkMode.update((current) => {
      const newValue = !current;
      setDarkMode(newValue);
      return newValue;
    });
  }

  onMount(async () => {
    if (window.innerWidth < 1024) {
      isSidebarOpen = false;
    }

    try {
      const cfg = await api.getConfig();
      configStore.set(cfg);
      dashboardData = await api.getDashboard();
      isServiceActive = true;
    } catch (e) {
      addToast("設定の読み込みに失敗しました", "error");
      isServiceActive = false;
    }
    healthInterval = setInterval(checkHealth, 30000);
    clockInterval = setInterval(updateClock, 1000);
    updateClock();

    // Initialize dark mode
    const saved = localStorage.getItem('darkMode');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initialDark = saved !== null ? saved === 'true' : prefersDark;
    setDarkMode(initialDark);

    // Listen for system dark mode changes
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQueryListener = (e: MediaQueryListEvent) => {
      // Only auto-switch if no manual preference is saved
      if (localStorage.getItem('darkMode') === null) {
        setDarkMode(e.matches);
      }
    };
    mediaQuery.addEventListener('change', mediaQueryListener);
    window.addEventListener("keydown", handleGlobalKeydown);
  });

  onDestroy(() => {
    if (healthInterval) clearInterval(healthInterval);
    if (clockInterval) clearInterval(clockInterval);
    window.removeEventListener("keydown", handleGlobalKeydown);
    if (mediaQuery && mediaQueryListener) {
      mediaQuery.removeEventListener("change", mediaQueryListener);
    }
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
  $: currentDate = dateFormatter.format(now);
  $: currentWeekday = weekdayFormatter.format(now);
  $: currentTime = timeFormatter.format(now);
</script>

<a
  href="#main-content"
  class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[120] focus:px-3 focus:py-2 focus:rounded-lg focus:bg-brand-600 focus:text-white"
>
  メインコンテンツへスキップ
</a>

<div class="flex h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 overflow-hidden font-sans">
  <!-- Sidebar -->
  <aside
    class="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 transform transition-transform duration-300 lg:relative lg:translate-x-0 {isSidebarOpen
      ? 'translate-x-0'
      : '-translate-x-full'}"
    aria-label="サイドバー"
  >
    <div
      class="flex items-center justify-between h-14 px-6 border-b border-gray-100 dark:border-gray-700"
    >
      <div class="flex items-center gap-2">
        <div
          class="w-7 h-7 bg-brand-600 rounded flex items-center justify-center text-white"
        >
          <Bell size={16} />
        </div>
        <span class="font-bold text-base tracking-tight">Notion Notifier</span>
      </div>
      <button
        class="lg:hidden text-gray-500"
        on:click={closeSidebar}
        aria-label="サイドバーを閉じる"
      >
        <X size={20} />
      </button>
    </div>

    <nav id={mainNavId} class="p-4 space-y-1 overflow-y-auto h-[calc(100%-3.5rem)]" aria-label="メインナビゲーション">
      {#each navItems as item}
        <button
          on:click={() => {
            navigate(item.path);
            if (window.innerWidth < 1024) closeSidebar();
          }}
          class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group {$activeRoute ===
          item.path
            ? 'bg-brand-50 dark:bg-brand-900/20 text-brand-700 dark:text-brand-300 font-semibold'
            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100'}"
          aria-current={$activeRoute === item.path ? "page" : undefined}
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

      <div class="px-4 py-6 mt-auto border-t border-gray-100 dark:border-gray-700 space-y-4">
        <div class="px-3 flex items-center justify-between">
          <span class="text-[10px] font-bold text-gray-400 dark:text-gray-500 tracking-widest uppercase">System Status</span>
          <div
            class="flex items-center gap-1.5"
            title={isServiceActive ? "稼働中" : "オフライン"}
          >
            <div
              class="w-2 h-2 rounded-full {isServiceActive
                ? 'bg-green-500 animate-pulse shadow-[0_0_8px_rgba(34,197,94,0.4)]'
                : 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.4)]'}"
            ></div>
            <span class="text-[10px] font-bold {isServiceActive ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
              {isServiceActive ? "ONLINE" : "OFFLINE"}
            </span>
          </div>
        </div>

        <div class="space-y-1">
          <button
            on:click={handleSync}
            disabled={isSyncing}
            class="w-full flex items-center justify-between px-3 py-2 rounded-xl transition-all duration-200 text-gray-600 dark:text-gray-400 hover:bg-brand-50 dark:hover:bg-brand-900/10 hover:text-brand-700 dark:hover:text-brand-300 disabled:opacity-50 group"
          >
            <div class="flex items-center gap-3">
              <div class={isSyncing ? "animate-spin" : "transition-transform duration-200 group-hover:rotate-180"}>
                <RefreshCcw size={16} />
              </div>
              <span class="text-sm font-medium">Notion Sync</span>
            </div>
            {#if dashboardData}
              <span class="text-[10px] tabular-nums font-medium opacity-60">
                {dashboardData.last_sync ? new Date(dashboardData.last_sync).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' }) : '--:--'}
              </span>
            {/if}
          </button>

          <button
            on:click={handleToggleSnooze}
            class="w-full flex items-center justify-between px-3 py-2 rounded-xl transition-all duration-200 {dashboardData?.snooze_active 
              ? 'bg-amber-50 dark:bg-amber-900/10 text-amber-700 dark:text-amber-300' 
              : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700'}"
          >
            <div class="flex items-center gap-3">
              {#if dashboardData?.snooze_active}
                <BellOff size={16} />
              {:else}
                <Bell size={16} />
              {/if}
              <span class="text-sm font-medium">Snooze (1h)</span>
            </div>
            {#if dashboardData?.snooze_active}
              <span class="text-[10px] font-bold">ON</span>
            {/if}
          </button>
        </div>

        {#if dashboardData?.snooze_active}
          <div class="mx-3 px-3 py-2 bg-gray-100 dark:bg-gray-700/50 rounded-lg flex items-center gap-2 text-gray-500 dark:text-gray-400">
            <BellOff size={14} />
            <span class="text-[10px] font-bold uppercase tracking-wider">SNOOZE ACTIVE</span>
          </div>
        {/if}

      </div>
    </nav>
  </aside>

  {#if isSidebarOpen}
    <button
      class="fixed inset-0 z-40 bg-black/30 backdrop-blur-[1px] lg:hidden"
      on:click={closeSidebar}
      aria-label="サイドバーを閉じる"
    ></button>
  {/if}

  <!-- Main Content -->
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <header
      class="h-14 bg-white/80 dark:bg-gray-800/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700 flex items-center px-4 md:px-8 justify-between sticky top-0 z-40"
    >
      <div class="flex items-center gap-4">
        <button
          class="lg:hidden p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          on:click={openSidebar}
          aria-expanded={isSidebarOpen}
          aria-controls={mainNavId}
          aria-label="サイドバーを開く"
        >
          <Menu size={20} />
        </button>
        <h1 class="text-xl font-bold text-gray-800 dark:text-gray-200">
          {navItems.find((n) => n.path === $activeRoute)?.label || "Dashboard"}
        </h1>
      </div>

      <div class="flex items-center gap-2 md:gap-4">
        <div
          class="hidden sm:flex items-center gap-4 text-sm font-medium text-gray-500 dark:text-gray-400 tabular-nums"
          aria-label="現在日時"
        >
          <div class="flex items-center gap-1.5">
            <Calendar size={14} />
            <span>{currentDate} <span class="text-xs opacity-70">({currentWeekday})</span></span>
          </div>
          <div class="flex items-center gap-1.5 border-l border-gray-200 dark:border-gray-700 pl-4">
            <Clock size={14} />
            <span class="text-gray-700 dark:text-gray-200 font-semibold">{currentTime}</span>
          </div>
        </div>

        <div class="h-4 w-px bg-gray-200 dark:bg-gray-700"></div>

        <button
          on:click={toggleDarkMode}
          class="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          aria-label={$darkMode ? "ライトモードに切り替え" : "ダークモードに切り替え"}
        >
          {#if $darkMode}
            <Sun size={18} />
          {:else}
            <Moon size={18} />
          {/if}
        </button>
      </div>
    </header>

    <main id="main-content" class="flex-1 overflow-y-auto p-4 md:p-8 custom-scrollbar">
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
          aria-label="通知を閉じる"
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
