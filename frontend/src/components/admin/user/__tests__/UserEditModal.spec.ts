import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, beforeEach, vi } from 'vitest'
import { defineComponent, h } from 'vue'

import UserEditModal from '../UserEditModal.vue'

const mockGetAllGroups = vi.fn()
const mockUpdateUser = vi.fn()
const mockUpdateUserAttributes = vi.fn()
const mockShowError = vi.fn()
const mockShowSuccess = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      getAll: (...args: unknown[]) => mockGetAllGroups(...args),
    },
    users: {
      update: (...args: unknown[]) => mockUpdateUser(...args),
    },
    userAttributes: {
      updateUserAttributeValues: (...args: unknown[]) => mockUpdateUserAttributes(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mockShowError,
    showSuccess: mockShowSuccess,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn(),
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
  },
  setup(_, { slots }) {
    return () => h('div', {}, [slots.default?.(), slots.footer?.()])
  },
})

const UserAttributeFormStub = defineComponent({
  name: 'UserAttributeFormStub',
  props: {
    modelValue: {
      type: Object,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue'],
  setup() {
    return () => h('div', { class: 'user-attribute-form-stub' })
  },
})

const TierGroupChainEditorStub = defineComponent({
  name: 'TierGroupChainEditorStub',
  props: {
    groups: {
      type: Array,
      default: () => [],
    },
  },
  setup(props) {
    return () => h('div', {
      class: 'tier-group-chain-editor-stub',
      'data-groups': JSON.stringify(props.groups),
    })
  },
})

describe('UserEditModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not clear default tier groups when group loading fails', async () => {
    mockGetAllGroups.mockRejectedValue(new Error('load failed'))
    mockUpdateUser.mockResolvedValue({})
    mockUpdateUserAttributes.mockResolvedValue({})

    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: {
          id: 42,
          email: 'user@example.com',
          username: 'tester',
          notes: 'original notes',
          concurrency: 3,
          rpm_limit: 0,
          default_tier_group_ids: [11, 12],
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          UserAttributeForm: UserAttributeFormStub,
          TierGroupChainEditor: TierGroupChainEditorStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.tier-group-chain-editor-stub').exists()).toBe(false)

    await wrapper.find('textarea').setValue('updated notes')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockUpdateUser).toHaveBeenCalledTimes(1)
    expect(mockUpdateUser).toHaveBeenCalledWith(42, {
      email: 'user@example.com',
      username: 'tester',
      notes: 'updated notes',
      concurrency: 3,
      rpm_limit: 0,
    })
  })

  it('passes only active OpenAI groups to the tier editor', async () => {
    mockGetAllGroups.mockResolvedValue([
      {
        id: 1,
        name: 'Anthropic Group',
        platform: 'anthropic',
        status: 'active',
      },
      {
        id: 2,
        name: 'OpenAI Active',
        platform: 'openai',
        status: 'active',
      },
      {
        id: 3,
        name: 'OpenAI Disabled',
        platform: 'openai',
        status: 'disabled',
      },
    ])

    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user: {
          id: 42,
          email: 'user@example.com',
          username: 'tester',
          notes: 'original notes',
          concurrency: 3,
          rpm_limit: 0,
          default_tier_group_ids: [],
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          UserAttributeForm: UserAttributeFormStub,
          TierGroupChainEditor: TierGroupChainEditorStub,
          Icon: true,
        },
      },
    })

    await flushPromises()

    const editor = wrapper.find('.tier-group-chain-editor-stub')
    expect(editor.exists()).toBe(true)
    expect(JSON.parse(editor.attributes('data-groups') ?? '[]')).toEqual([
      { value: 2, label: 'OpenAI Active' },
    ])
  })
})
