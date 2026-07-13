<script setup lang="ts">
import {
  ChevronDown,
  Code2,
  GitBranch,
  Link2,
  SquareTerminal,
  Star,
} from "lucide-vue-next";
import { invoke } from "@tauri-apps/api/core";
import { openUrl, revealItemInDir } from "@tauri-apps/plugin-opener";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { EntryPoint, Opener, Worktree } from "@/types";

const props = defineProps<{ worktree: Worktree; openers: Opener[] }>();

// wt stores bare hosts like "acme.test"; make them a real URL before opening.
function openEntryPoint(ep: EntryPoint) {
  if (!ep.url) return;
  const href = /^[a-z][a-z0-9+.-]*:\/\//i.test(ep.url)
    ? ep.url
    : `http://${ep.url}`;
  void openUrl(href);
}

// Reveal the worktree folder in Finder so people can jump straight to it.
function openInFinder() {
  void revealItemInDir(props.worktree.path);
}

// Terminals get a terminal glyph; editors get a code glyph.
function iconFor(id: string) {
  return id === "terminal" || id === "warp" ? SquareTerminal : Code2;
}

function openWith(id: string) {
  void invoke("open_in_app", { id, path: props.worktree.path });
}
</script>

<template>
  <div class="flex flex-col gap-1.5 rounded-md px-2.5 py-2">
    <div class="flex items-center gap-2">
      <component
        :is="worktree.special ? Star : GitBranch"
        class="size-3.5 shrink-0 text-muted-foreground"
      />
      <span class="truncate text-sm font-medium">{{ worktree.name }}</span>

      <div class="ml-auto flex items-center gap-1.5">
        <Badge v-if="worktree.special" variant="secondary">main</Badge>

        <DropdownMenu v-if="openers.length">
          <DropdownMenuTrigger
            class="flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus:outline-none focus-visible:ring-1 focus-visible:ring-ring data-[state=open]:bg-accent data-[state=open]:text-foreground"
            title="Open with…"
          >
            Open with
            <ChevronDown class="size-3" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Open with</DropdownMenuLabel>
            <DropdownMenuItem
              v-for="opener in openers"
              :key="opener.id"
              @select="openWith(opener.id)"
            >
              <component
                :is="iconFor(opener.id)"
                class="size-3.5 text-muted-foreground"
              />
              {{ opener.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>

    <p
      v-if="worktree.description"
      class="pl-5.5 text-xs text-muted-foreground line-clamp-2"
    >
      {{ worktree.description }}
    </p>

    <button
      type="button"
      :title="`Reveal in Finder — ${worktree.path}`"
      class="block w-full cursor-pointer truncate border-0 bg-transparent pl-5.5 text-left font-mono text-[11px] text-muted-foreground/70 transition-colors hover:text-foreground hover:underline"
      @click="openInFinder"
    >
      {{ worktree.path }}
    </button>

    <div
      v-if="worktree.entry_points.length"
      class="flex flex-wrap gap-1 pl-5.5 pt-0.5"
    >
      <component
        v-for="ep in worktree.entry_points"
        :is="ep.url ? 'button' : 'span'"
        :key="ep.name"
        :type="ep.url ? 'button' : undefined"
        :title="ep.url ? `Open ${ep.url}` : undefined"
        class="inline-flex border-0 bg-transparent p-0"
        @click="ep.url && openEntryPoint(ep)"
      >
        <Badge
          variant="outline"
          class="gap-1"
          :class="
            ep.url &&
            'cursor-pointer transition-colors hover:bg-accent hover:text-accent-foreground'
          "
        >
          <Link2 class="size-3" />
          <span class="font-medium">{{ ep.name }}</span>
          <span v-if="ep.url" class="text-muted-foreground"
            >· {{ ep.url }}</span
          >
        </Badge>
      </component>
    </div>
  </div>
</template>
