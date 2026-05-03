<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    activeRoute,
    configStore,
    addToast,
    darkMode,
    syncNotion as syncNotionState,
    healthPoller,
    dashboardStore,
    serviceActiveStore,
  } from "./lib/store";
  import { sidebarOpen, guideModal } from "./lib/uiStore";
  import { api } from "./lib/api";
  import {
    LayoutDashboard,
    Bell,
    Calendar,
    Settings,
    History,
  } from "lucide-svelte";

  // Routes
  import Dashboard from "./routes/Dashboard.svelte";
  import Notifications from "./routes/Notifications.svelte";
  import CalendarSync from "./routes/Calendar.svelte";
  import SystemSettings from "./routes/Settings.svelte";
  import NotificationHistory from "./routes/History.svelte";
  import PreviewModal from "./components/PreviewModal.svelte";
  import Sidebar from "./components/layout/Sidebar.svelte";
  import Header from "./components/layout/Header.svelte";
  import ToastContainer from "./components/ToastContainer.svelte";
  $: dashboardData = $dashboardStore;
  $: isServiceActive = $serviceActiveStore;
  let isSyncing = false;
  $: config = $configStore;
  const mainNavId = "main-navigation";
  let unsubscribeDarkMode: () => void;

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key === "Escape" && window.innerWidth < 1024 && $sidebarOpen) {
      sidebarOpen.close();
    }
  }

  async function handleSync() {
    if (isSyncing) return;
    isSyncing = true;
    try {
      await syncNotionState({
        successMessage: (count) => `${count}件のイベントを同期しました`,
        errorMessage: "同期に失敗しました",
        onSynced: () => healthPoller.forceCheck(),
      });
    } finally {
      isSyncing = false;
    }
  }

  async function saveSnooze() {
    if (!config) return;
    try {
      const saved = await api.updateSnooze(config.snooze);
      config.snooze = saved;
      configStore.set(config);
      await healthPoller.forceCheck();
    } catch {
      addToast("スヌーズ設定の保存に失敗しました", "error");
    }
  }

  async function clearSnooze() {
    if (!config) return;
    config.snooze = {
      until: "",
      mute_upcoming: true,
      mute_periodic: true,
    };
    configStore.set(config);
    await saveSnooze();
  }

  onMount(async () => {
    try {
      const cfg = await api.getConfig();
      configStore.set(cfg);
      await healthPoller.forceCheck();
    } catch {
      addToast("設定の読み込みに失敗しました", "error");
    }

    healthPoller.start();
    unsubscribeDarkMode = darkMode.init();

    window.addEventListener("keydown", handleGlobalKeydown);
  });

  onDestroy(() => {
    healthPoller.stop();
    if (unsubscribeDarkMode) unsubscribeDarkMode();
    window.removeEventListener("keydown", handleGlobalKeydown);
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
  $: showTemplateGuide =
    $activeRoute === "/notifications" || $activeRoute === "/settings";
</script>

<a
  href="#main-content"
  class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[120] focus:px-3 focus:py-2 focus:rounded-lg focus:bg-brand-600 focus:text-white"
>
  メインコンテンツへスキップ
</a>

<div
  class="flex h-screen overflow-hidden bg-gray-100 text-gray-900 dark:bg-gray-950 dark:text-gray-100 font-sans"
>
  <!-- Sidebar -->
  <Sidebar
    {navItems}
    activeRouteValue={$activeRoute}
    {isSyncing}
    {dashboardData}
    {config}
    {showTemplateGuide}
    {mainNavId}
    on:sync={handleSync}
    on:saveSnooze={saveSnooze}
    on:clearSnooze={clearSnooze}
  />

  {#if $sidebarOpen}
    <button
      class="fixed inset-0 z-40 bg-black/30 backdrop-blur-[1px] lg:hidden"
      on:click={() => sidebarOpen.close()}
      aria-label="サイドバーを閉じる"
    ></button>
  {/if}

  <!-- Main Content -->
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <Header
      activeRouteLabel={navItems.find((n) => n.path === $activeRoute)?.label ||
        "Dashboard"}
      {isServiceActive}
      {mainNavId}
    />

    <main
      id="main-content"
      class="ui-scrollbar flex-1 overflow-y-auto p-4 md:p-8"
    >
      <div class="max-w-6xl mx-auto">
        <svelte:component this={currentComponent} />
      </div>
    </main>
  </div>

  <PreviewModal
    open={$guideModal.isOpen}
    title={$guideModal.title}
    content={$guideModal.content}
    mode="guide"
    on:close={() => guideModal.close()}
  />

  <ToastContainer />
</div>
