import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'

import TierGroupChainEditor from '../TierGroupChainEditor.vue'

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: Number,
      default: undefined,
    },
    options: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue'],
  setup(props) {
    return () => h('div', {
      class: 'select-stub',
      'data-model-value': props.modelValue,
      'data-options-count': Array.isArray(props.options) ? props.options.length : 0,
    })
  },
})

const DraggableStub = defineComponent({
  name: 'VueDraggableStub',
  props: {
    modelValue: {
      type: Array,
      default: () => [],
    },
  },
  setup(_, { slots }) {
    return () => h('div', { class: 'draggable-stub' }, slots.default?.())
  },
})

describe('TierGroupChainEditor', () => {
  it('preserves unavailable existing ids instead of silently dropping them on mount', () => {
    const wrapper = mount(TierGroupChainEditor, {
      props: {
        modelValue: [2, 999],
        groups: [
          { value: 1, label: 'primary-a' },
          { value: 2, label: 'fallback-b' },
        ],
        excludeIds: [1],
      },
      global: {
        stubs: {
          Select: SelectStub,
          VueDraggable: DraggableStub,
        },
      },
    })

    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    const selects = wrapper.findAll('.select-stub')
    expect(selects).toHaveLength(2)
    expect(selects[0].attributes('data-model-value')).toBe('2')
    expect(selects[1].attributes('data-model-value')).toBe('999')
  })

  it('drops excluded ids when parent excludeIds changes', async () => {
    const wrapper = mount(TierGroupChainEditor, {
      props: {
        modelValue: [2, 3],
        groups: [
          { value: 1, label: 'primary-a' },
          { value: 2, label: 'fallback-b' },
          { value: 3, label: 'fallback-c' },
        ],
        excludeIds: [1],
      },
      global: {
        stubs: {
          Select: SelectStub,
          VueDraggable: DraggableStub,
        },
      },
    })

    await wrapper.setProps({ excludeIds: [2] })

    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toBeTruthy()
    expect(emitted?.at(-1)).toEqual([[3]])
  })
})
