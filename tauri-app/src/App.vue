<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";
import { listen, type UnlistenFn } from "@tauri-apps/api/event";
import { FolderGit2, Inbox, RefreshCw, Settings } from "lucide-vue-next";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import WorktreeItem from "@/components/WorktreeItem.vue";
import SettingsPane from "@/components/SettingsPane.vue";
import type { Opener, Project, StatusCheck } from "@/types";

type View = "list" | "settings";

const view = ref<View>("list");

const projects = ref<Project[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const spinning = ref(false);

const checks = ref<StatusCheck[]>([]);
const checksLoading = ref(false);

// Installed "Open with…" apps — detected once (per-machine, not per-worktree).
const openers = ref<Opener[]>([]);
const availableOpeners = computed(() =>
  openers.value.filter((o) => o.available),
);

const worktreeCount = computed(() =>
  projects.value.reduce((n, p) => n + p.worktrees.length, 0),
);

async function refresh() {
  spinning.value = true;
  try {
    projects.value = await invoke<Project[]>("list_projects");
    error.value = null;
  } catch (e) {
    error.value = String(e);
  } finally {
    loading.value = false;
    // Keep the spin visible briefly so a refresh always reads as feedback.
    setTimeout(() => (spinning.value = false), 250);
  }
}

async function loadChecks() {
  checksLoading.value = true;
  try {
    checks.value = await invoke<StatusCheck[]>("check_dependencies");
  } finally {
    checksLoading.value = false;
  }
}

function toggleSettings() {
  if (view.value === "settings") {
    view.value = "list";
  } else {
    view.value = "settings";
    void loadChecks();
  }
}

// The refresh button re-runs whichever pane is showing.
function onRefresh() {
  if (view.value === "settings") void loadChecks();
  else void refresh();
}

let unlisten: UnlistenFn | undefined;

onMounted(async () => {
  await refresh();
  // Detecting installed apps is instant (filesystem checks); load it once.
  invoke<Opener[]>("list_openers")
    .then((o) => (openers.value = o))
    .catch(() => {});
  // The Rust watcher emits this whenever the registry directory changes.
  unlisten = await listen<Project[]>("registry-changed", (event) => {
    projects.value = event.payload;
    error.value = null;
    loading.value = false;
  });
});

onUnmounted(() => unlisten?.());
</script>

<template>
  <div class="flex h-full flex-col overflow-hidden bg-background">
    <!-- Header -->
    <header class="flex items-center gap-2 px-3.5 py-2.5">
      <template v-if="view === 'settings'">
        <Settings class="size-4 text-primary" />
        <span class="text-sm font-semibold">Settings</span>
      </template>
      <template v-else>
        <FolderGit2 class="size-4 text-primary" />
        <div class="flex flex-col leading-tight">
          <span class="text-sm font-semibold">Worktrees</span>
          <span class="text-[11px] text-muted-foreground">
            {{ worktreeCount }} across {{ projects.length }}
            {{ projects.length === 1 ? "project" : "projects" }}
          </span>
        </div>
      </template>

      <button
        class="ml-auto flex size-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        :title="view === 'settings' ? 'Re-check' : 'Refresh'"
        @click="onRefresh"
      >
        <RefreshCw
          class="size-4"
          :class="(spinning || checksLoading) && 'animate-spin'"
        />
      </button>
      <button
        class="flex size-7 items-center justify-center rounded-md transition-colors hover:bg-accent hover:text-foreground"
        :class="
          view === 'settings'
            ? 'bg-accent text-foreground'
            : 'text-muted-foreground'
        "
        :title="view === 'settings' ? 'Close settings' : 'Settings'"
        @click="toggleSettings"
      >
        <Settings class="size-4" />
      </button>
    </header>

    <Separator />

    <!-- Body -->
    <ScrollArea class="flex-1">
      <SettingsPane
        v-if="view === 'settings'"
        :checks="checks"
        :loading="checksLoading"
      />

      <div v-else class="flex flex-col gap-2.5 p-3">
        <div
          v-if="loading"
          class="py-10 text-center text-sm text-muted-foreground"
        >
          Loading…
        </div>

        <div
          v-else-if="error"
          class="py-10 text-center text-sm text-destructive"
        >
          {{ error }}
        </div>

        <div
          v-else-if="!projects.length"
          class="flex flex-col items-center gap-2 py-12 text-center text-muted-foreground"
        >
          <Inbox class="size-8 opacity-40" />
          <p class="text-sm font-medium">No worktrees registered</p>
          <p class="max-w-[220px] text-xs">
            Run <code class="font-mono">wt register</code> inside a worktree to
            see it here.
          </p>
        </div>

        <Card v-for="project in projects" v-else :key="project.project_path">
          <CardHeader>
            <CardTitle class="truncate text-[13px]">
              {{ project.display_name }}
            </CardTitle>
          </CardHeader>
          <CardContent class="flex flex-col gap-0.5">
            <WorktreeItem
              v-for="wt in project.worktrees"
              :key="wt.path"
              :worktree="wt"
              :openers="availableOpeners"
            />
          </CardContent>
        </Card>
      </div>
    </ScrollArea>
  </div>
</template>
