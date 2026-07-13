<script setup lang="ts">
import { CheckCircle2, Loader2, XCircle } from "lucide-vue-next";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import type { StatusCheck } from "@/types";

defineProps<{ checks: StatusCheck[]; loading: boolean }>();
</script>

<template>
  <div class="flex flex-col gap-3 p-3">
    <p class="px-1 text-xs text-muted-foreground">
      Tools <code class="font-mono">wt</code> works with on this machine.
    </p>

    <Card>
      <CardContent class="flex flex-col gap-0.5 px-2 py-2">
        <div
          v-if="loading"
          class="flex items-center gap-2 px-2 py-3 text-sm text-muted-foreground"
        >
          <Loader2 class="size-4 animate-spin" />
          Checking…
        </div>

        <div
          v-for="check in checks"
          v-else
          :key="check.id"
          class="flex items-start gap-2.5 rounded-md px-2 py-2"
        >
          <component
            :is="check.ok ? CheckCircle2 : XCircle"
            class="mt-0.5 size-4 shrink-0"
            :class="check.ok ? 'text-emerald-500' : 'text-destructive'"
          />
          <div class="flex min-w-0 flex-col">
            <span class="text-sm font-medium">{{ check.label }}</span>
            <span
              class="truncate font-mono text-[11px]"
              :class="check.ok ? 'text-muted-foreground' : 'text-destructive/80'"
              :title="check.detail"
            >
              {{ check.detail }}
            </span>
          </div>
          <Badge
            class="ml-auto"
            :variant="check.ok ? 'secondary' : 'destructive'"
          >
            {{ check.ok ? "Found" : "Missing" }}
          </Badge>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
