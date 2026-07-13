<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { DropdownMenuContent, DropdownMenuPortal } from "reka-ui";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes["class"];
    align?: "start" | "center" | "end";
    sideOffset?: number;
  }>(),
  { align: "end", sideOffset: 4 },
);
</script>

<template>
  <DropdownMenuPortal>
    <DropdownMenuContent
      :align="props.align"
      :side-offset="props.sideOffset"
      :class="
        cn(
          'z-50 min-w-[11rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md',
          props.class,
        )
      "
    >
      <slot />
    </DropdownMenuContent>
  </DropdownMenuPortal>
</template>
