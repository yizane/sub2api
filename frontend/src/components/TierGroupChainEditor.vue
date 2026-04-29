<template>
  <div class="space-y-2">
    <!-- Chain items (draggable) -->
    <VueDraggable
      v-if="localChain.length > 0"
      v-model="localChain"
      :animation="200"
      handle=".drag-handle"
      class="space-y-2"
      @end="emitUpdate"
    >
      <div
        v-for="(gid, idx) in localChain"
        :key="gid"
        class="flex items-center gap-2"
      >
        <div class="drag-handle flex cursor-grab items-center text-gray-300 hover:text-gray-500 active:cursor-grabbing dark:text-dark-600 dark:hover:text-dark-400">
          <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path d="M7 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4z"/>
          </svg>
        </div>
        <span class="w-5 text-center text-xs font-medium text-gray-400 dark:text-gray-500">{{ idx + 1 }}</span>
        <div class="min-w-0 flex-1">
          <Select
            :model-value="gid"
            :options="availableOptionsFor(gid)"
            :placeholder="placeholderText"
            :searchable="true"
            :search-placeholder="searchPlaceholderText"
            @update:model-value="(v) => updateItem(idx, v as number)"
          />
        </div>
        <button
          type="button"
          @click="removeItem(idx)"
          class="flex-shrink-0 rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20"
        >
          <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
          </svg>
        </button>
      </div>
    </VueDraggable>

    <!-- Add button -->
    <button
      v-if="canAddMore"
      type="button"
      @click="addItem"
      class="flex items-center gap-1.5 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
    >
      <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd"/>
      </svg>
      {{ addButtonText }}
    </button>

    <p v-if="hintText" class="input-hint">{{ hintText }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import Select from '@/components/common/Select.vue'

interface GroupOption {
  value: number
  label: string
  [key: string]: unknown
}

const props = withDefaults(defineProps<{
  modelValue: number[]
  groups: GroupOption[]
  excludeIds?: number[]
  maxItems?: number
  placeholderText?: string
  searchPlaceholderText?: string
  addButtonText?: string
  hintText?: string
}>(), {
  excludeIds: () => [],
  maxItems: 16,
  placeholderText: 'Select group',
  searchPlaceholderText: 'Search groups',
  addButtonText: 'Add fallback group',
  hintText: ''
})

const emit = defineEmits<{
  'update:modelValue': [value: number[]]
}>()

const sameNumberArray = (a: number[], b: number[]) =>
  a.length === b.length && a.every((value, index) => value === b[index])

function sanitizeChain(chain: number[]): number[] {
  const excluded = new Set(props.excludeIds)
  const seen = new Set<number>()

  return chain.filter((groupID) => {
    if (groupID <= 0 || excluded.has(groupID) || seen.has(groupID)) {
      return false
    }
    seen.add(groupID)
    return true
  })
}

const localChain = ref<number[]>(sanitizeChain(props.modelValue))

watch(() => props.modelValue, (val) => {
  const next = sanitizeChain(val)
  localChain.value = next
  if (!sameNumberArray(next, val)) {
    emit('update:modelValue', [...next])
  }
}, { immediate: true })

watch(
  () => [props.excludeIds, props.groups] as const,
  () => {
    const next = sanitizeChain(localChain.value)
    if (!sameNumberArray(next, localChain.value)) {
      localChain.value = next
      emitUpdate()
    }
  },
  { deep: true }
)

watch(localChain, (val) => {
  const next = sanitizeChain(val)
  if (!sameNumberArray(next, val)) {
    localChain.value = next
  }
})

const canAddMore = computed(() =>
  localChain.value.length < props.maxItems && availableGlobalOptions.value.length > 0
)

const usedIds = computed(() => new Set([...localChain.value, ...props.excludeIds]))

const availableGlobalOptions = computed(() =>
  props.groups.filter(g => !usedIds.value.has(g.value))
)

function availableOptionsFor(currentId: number): GroupOption[] {
  const others = new Set([...props.excludeIds, ...localChain.value.filter(id => id !== currentId)])
  const options = props.groups.filter(g => !others.has(g.value))
  if (options.some((group) => group.value === currentId) || currentId <= 0) {
    return options
  }
  return [
    { value: currentId, label: `#${currentId} (unavailable)` },
    ...options,
  ]
}

function addItem() {
  const first = availableGlobalOptions.value[0]
  if (!first) return
  localChain.value = [...localChain.value, first.value]
  emitUpdate()
}

function removeItem(idx: number) {
  const copy = [...localChain.value]
  copy.splice(idx, 1)
  localChain.value = copy
  emitUpdate()
}

function updateItem(idx: number, newId: number) {
  const copy = [...localChain.value]
  copy[idx] = newId
  localChain.value = copy
  emitUpdate()
}

function emitUpdate() {
  emit('update:modelValue', [...localChain.value])
}
</script>
