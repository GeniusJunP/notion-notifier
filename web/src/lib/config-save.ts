import { api, type Config } from "./api";
import { addToast, configStore } from "./store";

interface SaveConfigOptions {
  successMessage?: string;
  errorMessage: string;
  onSaved?: (saved: Config) => Promise<void> | void;
}

export async function saveConfigWithStore(
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
      addToast(options.successMessage, "success");
    }
    return saved;
  } catch (e: any) {
    const detail = e?.error ? `: ${e.error}` : "";
    addToast(`${options.errorMessage}${detail}`, "error");
    return null;
  }
}
